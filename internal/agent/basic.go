package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"

	"spec-collector/internal/models"
	"spec-collector/internal/platform"
)

type wmiVideoController struct {
	Name          string
	AdapterRAM    uint64
	DriverVersion string
}

type win32ComputerSystem struct {
	Manufacturer string
	Model        string
	UserName     string
}

type win32ComputerSystemProduct struct {
	IdentifyingNumber string
}

type win32OS struct {
	Caption        string
	Version        string
	BuildNumber    string
	OSArchitecture string
}

type win32Processor struct {
	Name              string
	NumberOfCores     uint32
	NumberOfLogicalProcessors uint32
	MaxClockSpeed     uint32
	ThermalDesignPower uint32
}

type win32DiskDrive struct {
	Model         string
	Size          uint64
	InterfaceType string
}

type win32LogicalDisk struct {
	DeviceID  string
	Size      uint64
	FreeSpace uint64
	DriveType uint32
}

func Gather() (*models.Machine, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("host info: %w", err)
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("cpu info: %w", err)
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("memory info: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("hostname: %w", err)
	}

	ipAddr := getIPAddress()
	macAddr := getMACAddress()
	disks := getDisks()

	gpuList := getGPUs()
	sysInfo := getSystemInfo()

	var cpuCores int
	if len(cpuInfo) > 0 {
		cpuCores = int(cpuInfo[0].Cores)
	}
	cpuThreads, _ := cpu.Counts(true)

	var cpuModel, cpuClock string
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
		cpuClock = fmt.Sprintf("%.0f", cpuInfo[0].Mhz)
	}

	cpuTDP := getCPUTDP()

	machineID := generateMachineID(hostname, macAddr)

	osInfo := getOSInfo(hostInfo)

	m := &models.Machine{
		ID:        machineID,
		Hostname:  hostname,
		IPAddress: ipAddr,
		OS:        osInfo,
		CPU: models.CPUInfo{
			Model:    cpuModel,
			Cores:    cpuCores,
			Threads:  cpuThreads,
			MaxClock: cpuClock,
			TDPW:     cpuTDP,
		},
		RAM: models.RAMInfo{
			TotalGB: float64(memInfo.Total) / (1024 * 1024 * 1024),
		},
		Disks:  disks,
		GPU:    gpuList,
		System: sysInfo,
	}

	mobo, bios, netAdapters, monitors, software, sensors, ramSlots, battery := GatherDetailed()
	m.Motherboard = mobo
	m.BIOS = bios
	m.Network = netAdapters
	m.Monitors = monitors
	m.Software = software
	m.Sensors = sensors
	m.RAMSlots = ramSlots
	m.RAM.Slots = getRAMSlotCount()
	m.Battery = battery
	m.BitLocker = getBitLocker()

	return m, nil
}

func getOSInfo(hostInfo *host.InfoStat) models.OSInfo {
	osInfo := models.OSInfo{
		Name:          hostInfo.Platform + " " + hostInfo.PlatformVersion,
		Version:       hostInfo.PlatformVersion,
		Architecture:  runtime.GOARCH,
		LastBoot:      time.Unix(int64(hostInfo.BootTime), 0).Format(time.RFC3339),
	}

	if platform.IsWindows() {
		var wmiOS []win32OS
		if err := platform.QueryWMI("SELECT Caption, Version, BuildNumber, OSArchitecture FROM Win32_OperatingSystem", &wmiOS); err == nil && len(wmiOS) > 0 {
			osInfo.BuildNumber = wmiOS[0].BuildNumber
			if wmiOS[0].OSArchitecture != "" {
				osInfo.Architecture = wmiOS[0].OSArchitecture
			}
			caption := strings.TrimPrefix(wmiOS[0].Caption, "Microsoft ")
			osInfo.Name = caption

			verParts := strings.Split(wmiOS[0].Version, ".")
			if len(verParts) >= 4 {
				osInfo.UBR = verParts[3]
			} else if len(verParts) >= 3 {
				osInfo.UBR = verParts[2]
			}

			displayVer, err := platform.ReadRegistryString(
				`SOFTWARE\Microsoft\Windows NT\CurrentVersion`, "DisplayVersion")
			if err == nil && displayVer != "" {
				osInfo.DisplayName = caption + " " + displayVer
			} else {
				osInfo.DisplayName = caption
			}
		}
	}

	return osInfo
}

func getCPUTDP() int {
	if !platform.IsWindows() {
		return 0
	}
	var proc []win32Processor
	if err := platform.QueryWMI("SELECT ThermalDesignPower FROM Win32_Processor", &proc); err != nil || len(proc) == 0 {
		return 0
	}
	return int(proc[0].ThermalDesignPower)
}

func getSystemInfo() *models.SystemInfo {
	if !platform.IsWindows() {
		return nil
	}
	si := &models.SystemInfo{}
	var cs []win32ComputerSystem
	if err := platform.QueryWMI("SELECT Manufacturer, Model, UserName FROM Win32_ComputerSystem", &cs); err == nil && len(cs) > 0 {
		si.Manufacturer = cs[0].Manufacturer
		si.Model = cs[0].Model
		si.UserName = cs[0].UserName
	}
	var csp []win32ComputerSystemProduct
	if err := platform.QueryWMI("SELECT IdentifyingNumber FROM Win32_ComputerSystemProduct", &csp); err == nil && len(csp) > 0 {
		si.Serial = csp[0].IdentifyingNumber
	}
	if si.Manufacturer == "" && si.Model == "" && si.Serial == "" && si.UserName == "" {
		return nil
	}
	return si
}

func getDisks() []models.DiskInfo {
	partitions, err := disk.Partitions(false)
	if err == nil {
		var disks []models.DiskInfo
		for _, p := range partitions {
			usage, err := disk.Usage(p.Mountpoint)
			if err != nil {
				continue
			}
			disks = append(disks, models.DiskInfo{
				Model:      p.Device,
				SizeGB:     float64(usage.Total) / (1024 * 1024 * 1024),
				MountPoint: p.Mountpoint,
			})
		}
		if len(disks) > 0 {
			return disks
		}
	}

	if platform.IsWindows() {
		return getDisksWMI()
	}
	return nil
}

func getDisksWMI() []models.DiskInfo {
	var drives []win32DiskDrive
	err := platform.QueryWMI("SELECT Model, Size, InterfaceType FROM Win32_DiskDrive", &drives)
	if err != nil {
		return nil
	}
	var disks []models.DiskInfo
	for i, d := range drives {
		if d.Size == 0 {
			continue
		}
		t := ""
		model := d.Model
		if strings.Contains(strings.ToLower(model), "nvme") {
			t = "NVMe"
		} else if strings.Contains(strings.ToLower(d.InterfaceType), "usb") {
			t = "USB"
		}
		disks = append(disks, models.DiskInfo{
			Model:  model,
			SizeGB: float64(d.Size) / (1024 * 1024 * 1024),
			Type:   t,
		})
		_ = i
	}
	return disks
}

func getGPUs() []models.GPUInfo {
	if !platform.IsWindows() {
		return nil
	}
	var wmiGPUs []wmiVideoController
	err := platform.QueryWMI("SELECT Name, AdapterRAM, DriverVersion FROM Win32_VideoController", &wmiGPUs)
	if err != nil {
		return nil
	}

	vramMap := getGPUVRAMFromCLI()
	regMap := getGPUVRAMFromRegistry()
	for k, v := range regMap {
		if _, exists := vramMap[k]; !exists {
			vramMap[k] = v
		}
	}

	var gpus []models.GPUInfo
	for _, g := range wmiGPUs {
		brand := parseGPUBrand(g.Name)
		vramMB := int(g.AdapterRAM / (1024 * 1024))
		if vramMB <= 0 || vramMB > 262144 {
			vramMB = 0
		}
		if vramMB == 0 && len(vramMap) > 0 {
			nameKey := strings.ToLower(strings.TrimSpace(g.Name))
			if v, ok := vramMap[nameKey]; ok && v > 0 && v < 262144*1024*1024 {
				vramMB = int(v / (1024 * 1024))
			}
		}
		gpus = append(gpus, models.GPUInfo{
			Brand:  brand,
			Model:  g.Name,
			VRAMMB: vramMB,
			Driver: g.DriverVersion,
		})
	}
	return gpus
}

func parseGPUBrand(model string) string {
	m := strings.ToLower(model)
	switch {
	case strings.Contains(m, "nvidia"):
		return "NVIDIA"
	case strings.Contains(m, "amd"), strings.Contains(m, "radeon"), strings.Contains(m, "ati"):
		return "AMD"
	case strings.Contains(m, "intel"):
		return "Intel"
	case strings.Contains(m, "microsoft"):
		return "Microsoft"
	case strings.Contains(m, "vmware"), strings.Contains(m, "virtualbox"), strings.Contains(m, "qemu"):
		return "Virtual"
	default:
		return ""
	}
}

func getIPAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() || ipnet.IP.To4() == nil {
				continue
			}
			return ipnet.IP.String()
		}
	}
	return ""
}

func getMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		mac := iface.HardwareAddr.String()
		if mac != "" {
			return mac
		}
	}
	return ""
}

func generateMachineID(hostname, mac string) string {
	h := sha256.Sum256([]byte(hostname + mac))
	return hex.EncodeToString(h[:])
}
