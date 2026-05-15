package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
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
	fmt.Fprint(os.Stderr, "  ✓ Đang thu thập thông tin máy tính...\n")

	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("cpu info: %w", err)
	}
	fmt.Fprint(os.Stderr, "  ✓ Bộ xử lý (CPU)\n")

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("memory info: %w", err)
	}
	fmt.Fprint(os.Stderr, "  ✓ Bộ nhớ RAM\n")

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("hostname: %w", err)
	}

	ipAddr := getIPAddress()
	macAddr := getMACAddress()
	disks := getDisks()
	fmt.Fprint(os.Stderr, "  ✓ Ổ đĩa\n")

	gpuList := getGPUs()
	fmt.Fprint(os.Stderr, "  ✓ Card đồ họa (GPU)\n")
	sysInfo := getSystemInfo()
	fmt.Fprint(os.Stderr, "  ✓ Thông tin hệ thống\n")

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
	fmt.Fprint(os.Stderr, "  ✓ Hoàn tất!\n")

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

func mergeDiskTypesWithPartitions(physicalDrives []models.DiskInfo) []models.DiskInfo {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return physicalDrives
	}
	var disks []models.DiskInfo
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		t := ""
		partSize := float64(usage.Total)
		for _, pd := range physicalDrives {
			diff := math.Abs(partSize - pd.SizeGB*1024*1024*1024)
			if diff < partSize*0.05 {
				t = pd.Type
				break
			}
		}
		disks = append(disks, models.DiskInfo{
			Model:      p.Device,
			SizeGB:     partSize / (1024 * 1024 * 1024),
			Type:       t,
			MountPoint: p.Mountpoint,
		})
	}
	if len(disks) > 0 {
		return disks
	}
	return physicalDrives
}

func getDisks() []models.DiskInfo {
	if platform.IsWindows() {
		if disks := runEmbeddedCrystalDiskInfo(); len(disks) > 0 {
			disks = mergeDiskTypesWithPartitions(disks)
			if len(disks) > 0 {
				return disks
			}
		}

		physicalDrives := getDisksWMI()
		return mergeDiskTypesWithPartitions(physicalDrives)
	}

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
	return nil
}

var nvmeGenPatterns = []struct {
	gen  string
	keywords []string
}{
	{"Gen5", []string{"9100", "sn8100", "t700", "t705", "pcie gen5", "gen5"}},
	{"Gen4", []string{"980 pro", "990 pro", "sn850", "sn810", "pm9a1", "pcie gen4", "gen4", "p41", "p44", "p51"}},
	{"Gen3", []string{"960", "970 evo", "970 pro", "980", "pm981", "sn550", "sn570", "sn750", "sn720", "pcie gen3", "gen3"}},
}

func nvmeGen(model string) string {
	m := strings.ToLower(model)
	for _, p := range nvmeGenPatterns {
		for _, kw := range p.keywords {
			if strings.Contains(m, kw) {
				return p.gen
			}
		}
	}
	return ""
}

func busTypeName(busType uint32) string {
	switch busType {
	case 3:
		return "SATA"
	case 7:
		return "USB"
	case 9:
		return "SAS"
	case 10:
		return "NVMe"
	case 17:
		return "NVMe"
	case 11:
		return "SATA"
	case 12:
		return "SD"
	case 13:
		return "eMMC"
	default:
		return ""
	}
}

type msftDisk struct {
	FriendlyName string
	Size         uint64
	BusType      uint32
	Number       uint32
}

func getDisksPowerShell() []models.DiskInfo {
	script := `$drives = Get-PhysicalDisk -ErrorAction SilentlyContinue; if ($drives) { foreach ($d in $drives) { Write-Output (($d.FriendlyName -replace '\|','-') + '|' + $d.Size + '|' + [int]$d.BusType + '|' + $d.MediaType) } }`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	out, err := cmd.Output()
	if err != nil || len(out) < 5 {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var disks []models.DiskInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		name := parts[0]
		size := parseUint64(parts[1])
		busType := parseUint32(parts[2])
		mediaType := strings.ToLower(strings.TrimSpace(parts[3]))
		if size == 0 {
			continue
		}
		t := detectDiskTypeFromBusAndMedia(name, busType, mediaType)
		disks = append(disks, models.DiskInfo{
			Model:  name,
			SizeGB: float64(size) / (1024 * 1024 * 1024),
			Type:   t,
		})
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func detectDiskTypeFromBusAndMedia(model string, busType uint32, mediaType string) string {
	bus := busTypeName(busType)
	m := strings.ToLower(model)

	switch bus {
	case "NVMe":
		gen := nvmeGen(model)
		if gen != "" {
			return "NVMe " + gen
		}
		return "NVMe"
	case "SATA":
		if mediaType == "ssd" {
			return "SATA SSD"
		}
		if mediaType == "hdd" {
			return "SATA HDD"
		}
		if strings.Contains(m, "ssd") {
			return "SATA SSD"
		}
		if strings.Contains(m, "hdd") || strings.Contains(m, "hard") {
			return "SATA HDD"
		}
		return "SATA SSD"
	case "USB":
		return "USB"
	default:
		if strings.Contains(m, "nvme") {
			gen := nvmeGen(model)
			if gen != "" {
				return "NVMe " + gen
			}
			return "NVMe"
		}
		if mediaType == "ssd" || strings.Contains(m, "ssd") {
			return "SSD"
		}
		if bus != "" {
			return bus
		}
		return "SCSI"
	}
}

func parseUint64(s string) uint64 {
	s = strings.TrimSpace(s)
	var v uint64
	fmt.Sscanf(s, "%d", &v)
	return v
}

func parseUint32(s string) uint32 {
	s = strings.TrimSpace(s)
	var v uint32
	fmt.Sscanf(s, "%d", &v)
	return v
}

func getDisksWMI() []models.DiskInfo {
	var storageDisks []msftDisk
	err := platform.QueryWMINamespace(
		"SELECT FriendlyName, Size, BusType, Number FROM MSFT_Disk",
		&storageDisks, `ROOT\Microsoft\Windows\Storage`,
	)

	if err == nil && len(storageDisks) > 0 {
		var disks []models.DiskInfo
		for _, d := range storageDisks {
			if d.Size == 0 {
				continue
			}
			t := detectDiskTypeFromBus(d.FriendlyName, d.BusType)
			disks = append(disks, models.DiskInfo{
				Model:  d.FriendlyName,
				SizeGB: float64(d.Size) / (1024 * 1024 * 1024),
				Type:   t,
			})
		}
		if len(disks) > 0 {
			return disks
		}
	}

	if ps := getDisksPowerShell(); len(ps) > 0 {
		return ps
	}

	var drives []win32DiskDrive
	err = platform.QueryWMI("SELECT Model, Size, InterfaceType FROM Win32_DiskDrive", &drives)
	if err != nil {
		return nil
	}
	var disks []models.DiskInfo
	for _, d := range drives {
		if d.Size == 0 {
			continue
		}
		disks = append(disks, models.DiskInfo{
			Model:  d.Model,
			SizeGB: float64(d.Size) / (1024 * 1024 * 1024),
			Type:   detectDiskTypeFromModel(d.Model, d.InterfaceType),
		})
	}
	return disks
}

func detectDiskTypeFromBus(model string, busType uint32) string {
	bus := busTypeName(busType)
	modelLower := strings.ToLower(model)

	switch bus {
	case "NVMe":
		gen := nvmeGen(model)
		if gen != "" {
			return "NVMe " + gen
		}
		return "NVMe"
	case "SATA":
		if strings.Contains(modelLower, "ssd") {
			return "SATA SSD"
		}
		if strings.Contains(modelLower, "hdd") || strings.Contains(modelLower, "hard") {
			return "SATA HDD"
		}
		return "SATA SSD"
	case "USB":
		return "USB"
	default:
		if strings.Contains(modelLower, "nvme") {
			gen := nvmeGen(model)
			if gen != "" {
				return "NVMe " + gen
			}
			return "NVMe"
		}
		if strings.Contains(modelLower, "sata") {
			if strings.Contains(modelLower, "ssd") {
				return "SATA SSD"
			}
			if strings.Contains(modelLower, "hdd") || strings.Contains(modelLower, "hard") {
				return "SATA HDD"
			}
			return "SATA SSD"
		}
		if strings.Contains(modelLower, "ssd") {
			return "SSD"
		}
		if strings.Contains(modelLower, "hdd") || strings.Contains(modelLower, "hard") {
			return "HDD"
		}
		if bus != "" {
			return bus
		}
		return "SCSI"
	}
}

func detectDiskTypeFromModel(model, ifaceType string) string {
	modelLower := strings.ToLower(model)
	ifaceLower := strings.ToLower(ifaceType)

	switch {
	case strings.Contains(modelLower, "nvme") || ifaceLower == "nvme":
		gen := nvmeGen(model)
		if gen != "" {
			return "NVMe " + gen
		}
		return "NVMe"
	case ifaceLower == "usb":
		return "USB"
	case strings.Contains(ifaceLower, "serial ata") || ifaceLower == "sata":
		if strings.Contains(modelLower, "ssd") {
			return "SATA SSD"
		}
		if strings.Contains(modelLower, "hdd") || strings.Contains(modelLower, "hard") {
			return "SATA HDD"
		}
		return "SATA SSD"
	case ifaceLower == "ide":
		if strings.Contains(modelLower, "ssd") {
			return "IDE SSD"
		}
		return "IDE HDD"
	default:
		if strings.Contains(modelLower, "sata") {
			if strings.Contains(modelLower, "ssd") {
				return "SATA SSD"
			}
			if strings.Contains(modelLower, "hdd") || strings.Contains(modelLower, "hard") {
				return "SATA HDD"
			}
			return "SATA SSD"
		}
		if strings.Contains(modelLower, "ssd") {
			return "SSD"
		}
		if strings.Contains(modelLower, "nvme") {
			gen := nvmeGen(model)
			if gen != "" {
				return "NVMe " + gen
			}
			return "NVMe"
		}
		return ifaceType
	}
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
