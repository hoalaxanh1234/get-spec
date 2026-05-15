package agent

import (
	"fmt"
	"os"
	"strings"

	"spec-collector/internal/models"
	"spec-collector/internal/platform"
)

type win32BaseBoard struct {
	Manufacturer string
	Product      string
	SerialNumber string
}

type win32BIOS struct {
	Manufacturer      string
	SMBIOSBIOSVersion string
	ReleaseDate       string
}

type win32NetAdapter struct {
	Name       string
	MACAddress string
	Speed      uint64
}

type win32NetAdapterConfig struct {
	Description string
	MACAddress  string
	IPAddress   []string
}

type win32DesktopMonitor struct {
	MonitorManufacturerID string
	Name                  string
	ScreenWidth           uint32
	ScreenHeight          uint32
	SerialNumberID        string
}

type win32PhysicalMemoryArray struct {
	MemoryDevices uint32
}

type win32PhysicalMemory struct {
	Capacity    uint64
	Speed       uint32
	MemoryType  uint16
	Manufacturer string
	PartNumber  string
}

type win32EncryptableVolume struct {
	DriveLetter      string
	ProtectionStatus *uint16
}

type win32Battery struct {
	EstimatedChargeRemaining uint16
	DesignCapacity           uint32
	Chemistry                uint16
	BatteryStatus            uint16
}

func GatherDetailed() (*models.MotherboardInfo, *models.BIOSInfo, []models.NetAdapter, []models.MonitorInfo, []models.SoftwareInfo, *models.SensorInfo, []models.RAMSlot, *models.BatteryInfo) {
	if !platform.IsWindows() {
		return nil, nil, nil, nil, nil, nil, nil, nil
	}

	mobo := getMotherboard()
	fmt.Fprint(os.Stderr, "  ✓ Mainboard\n")
	bios := getBIOS()
	fmt.Fprint(os.Stderr, "  ✓ BIOS\n")
	netAdapters := getNetworkAdapters()
	fmt.Fprint(os.Stderr, "  ✓ Mạng\n")
	monitors := getMonitors()
	fmt.Fprint(os.Stderr, "  ✓ Màn hình\n")
	ramSlots := getRAMSlots()
	fmt.Fprint(os.Stderr, "  ✓ Khe RAM\n")
	battery := getBattery()
	fmt.Fprint(os.Stderr, "  ✓ Pin\n")

	return mobo, bios, netAdapters, monitors, nil, nil, ramSlots, battery
}

func getMotherboard() *models.MotherboardInfo {
	var boards []win32BaseBoard
	err := platform.QueryWMI("SELECT Manufacturer, Product, SerialNumber FROM Win32_BaseBoard", &boards)
	if err != nil || len(boards) == 0 {
		return nil
	}
	return &models.MotherboardInfo{
		Manufacturer: boards[0].Manufacturer,
		Model:        boards[0].Product,
		Serial:       boards[0].SerialNumber,
	}
}

func getBIOS() *models.BIOSInfo {
	var biosList []win32BIOS
	err := platform.QueryWMI("SELECT Manufacturer, SMBIOSBIOSVersion, ReleaseDate FROM Win32_BIOS", &biosList)
	if err != nil || len(biosList) == 0 {
		return nil
	}
	return &models.BIOSInfo{
		Vendor:  biosList[0].Manufacturer,
		Version: biosList[0].SMBIOSBIOSVersion,
		Date:    biosList[0].ReleaseDate,
	}
}

func getNetworkAdapters() []models.NetAdapter {
	var adapters []win32NetAdapter
	err := platform.QueryWMI("SELECT Name, MACAddress, Speed FROM Win32_NetworkAdapter WHERE PhysicalAdapter=TRUE", &adapters)
	if err != nil {
		return nil
	}
	var configs []win32NetAdapterConfig
	platform.QueryWMI("SELECT Description, MACAddress, IPAddress FROM Win32_NetworkAdapterConfiguration WHERE IPEnabled=TRUE", &configs)

	ipByMAC := make(map[string][]string)
	for _, c := range configs {
		mac := strings.ToUpper(c.MACAddress)
		ipByMAC[mac] = c.IPAddress
	}

	var result []models.NetAdapter
	for _, a := range adapters {
		mac := strings.ToUpper(a.MACAddress)
		ips := ipByMAC[mac]
		speedMbps := int(a.Speed / 1000000)
		result = append(result, models.NetAdapter{
			Name:        a.Name,
			MAC:         a.MACAddress,
			IPAddresses: ips,
			SpeedMbps:   speedMbps,
		})
	}
	return result
}

func getMonitors() []models.MonitorInfo {
	var monitors []win32DesktopMonitor
	err := platform.QueryWMI("SELECT MonitorManufacturerID, Name, ScreenWidth, ScreenHeight, SerialNumberID FROM Win32_DesktopMonitor", &monitors)
	if err != nil {
		return nil
	}
	var result []models.MonitorInfo
	for _, m := range monitors {
		resolution := ""
		if m.ScreenWidth > 0 && m.ScreenHeight > 0 {
			resolution = formatResolution(int(m.ScreenWidth), int(m.ScreenHeight))
		}
		result = append(result, models.MonitorInfo{
			Manufacturer: m.MonitorManufacturerID,
			Model:        m.Name,
			Resolution:   resolution,
			Serial:       m.SerialNumberID,
		})
	}
	return result
}

func getRAMSlotCount() int {
	var arr []win32PhysicalMemoryArray
	if err := platform.QueryWMI("SELECT MemoryDevices FROM Win32_PhysicalMemoryArray", &arr); err == nil && len(arr) > 0 && arr[0].MemoryDevices > 0 {
		return int(arr[0].MemoryDevices)
	}
	type win32CS struct {
		NumberOfMemorySlots int
	}
	var cs []win32CS
	if err := platform.QueryWMI("SELECT NumberOfMemorySlots FROM Win32_ComputerSystem", &cs); err == nil && len(cs) > 0 && cs[0].NumberOfMemorySlots > 0 {
		return cs[0].NumberOfMemorySlots
	}
	return 0
}

func getRAMSlots() []models.RAMSlot {
	var mems []win32PhysicalMemory
	err := platform.QueryWMI("SELECT Capacity, Speed, MemoryType, Manufacturer, PartNumber FROM Win32_PhysicalMemory", &mems)
	if err != nil {
		return nil
	}
	var slots []models.RAMSlot
	for _, m := range mems {
		slots = append(slots, models.RAMSlot{
			SizeGB:       float64(m.Capacity) / (1024 * 1024 * 1024),
			SpeedMHz:     int(m.Speed),
			MemoryType:   memoryTypeString(m.MemoryType),
			Manufacturer: m.Manufacturer,
			PartNumber:   m.PartNumber,
		})
	}
	return slots
}

type wmiBatteryStatic struct {
	EstimatedChargeRemaining uint32
	DesignCapacity           uint32
	BatteryStatus            uint32
}

func getBattery() *models.BatteryInfo {
	var bats []win32Battery
	err := platform.QueryWMI("SELECT EstimatedChargeRemaining, DesignCapacity, Chemistry, BatteryStatus FROM Win32_Battery", &bats)
	if err == nil && len(bats) > 0 && bats[0].DesignCapacity > 0 {
		return &models.BatteryInfo{
			Present:        true,
			Capacity:       int(bats[0].EstimatedChargeRemaining),
			DesignCapacity: int(bats[0].DesignCapacity),
			Chemistry:      batteryChemistryString(bats[0].Chemistry),
			Status:         batteryStatusString(bats[0].BatteryStatus),
		}
	}
	var acpi []wmiBatteryStatic
	if err := platform.QueryWMINamespace("SELECT EstimatedChargeRemaining, DesignCapacity, BatteryStatus FROM BatteryStatic WHERE Active=true", &acpi, "ROOT\\WMI"); err == nil && len(acpi) > 0 && acpi[0].DesignCapacity > 0 {
		return &models.BatteryInfo{
			Present:        true,
			Capacity:       int(acpi[0].EstimatedChargeRemaining),
			DesignCapacity: int(acpi[0].DesignCapacity),
			Status:         acpiBatteryStatus(acpi[0].BatteryStatus),
		}
	}
	return nil
}

func acpiBatteryStatus(s uint32) string {
	switch s {
	case 0:
		return "Discharging"
	case 1:
		return "On AC"
	case 2:
		return "Fully Charged"
	case 3:
		return "Low"
	case 4:
		return "Critical"
	case 5:
		return "Charging"
	case 6:
		return "Charging High"
	case 7:
		return "Charging Low"
	case 8:
		return "Charging Critical"
	case 9:
		return "Undefined"
	case 10:
		return "Partially Charged"
	default:
		return ""
	}
}

func getBitLocker() *models.BitLockerInfo {
	if !platform.IsWindows() {
		return nil
	}
	var vols []win32EncryptableVolume
	err := platform.QueryWMINamespace(
		"SELECT DriveLetter, ProtectionStatus FROM Win32_EncryptableVolume",
		&vols, "ROOT\\CIMV2\\Security\\MicrosoftVolumeEncryption")
	if err != nil {
		return nil
	}
	for _, v := range vols {
		if v.ProtectionStatus != nil && *v.ProtectionStatus == 1 {
			return &models.BitLockerInfo{
				Enabled: true,
				Status:  "Bật \U0001F512",
			}
		}
	}
	return &models.BitLockerInfo{
		Enabled: false,
		Status:  "Tắt \U0001F513",
	}
}

func memoryTypeString(t uint16) string {
	switch t {
	case 20:
		return "DDR"
	case 21:
		return "DDR2"
	case 24:
		return "DDR3"
	case 26:
		return "DDR4"
	case 34:
		return "DDR5"
	default:
		return ""
	}
}

func batteryChemistryString(c uint16) string {
	switch c {
	case 1:
		return "Lead Acid"
	case 2:
		return "NiCad"
	case 3:
		return "NiMH"
	case 4:
		return "Li-ion"
	case 5:
		return "LiPo"
	default:
		return ""
	}
}

func batteryStatusString(s uint16) string {
	switch s {
	case 1:
		return "Discharging"
	case 2:
		return "On AC"
	case 3:
		return "Fully Charged"
	case 4:
		return "Low"
	case 5:
		return "Critical"
	case 6:
		return "Charging"
	case 7:
		return "Charging High"
	case 8:
		return "Charging Low"
	case 9:
		return "Charging Critical"
	case 10:
		return "Undefined"
	case 11:
		return "Partially Charged"
	default:
		return ""
	}
}

func formatResolution(w, h int) string {
	return itoa(w) + "x" + itoa(h)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
