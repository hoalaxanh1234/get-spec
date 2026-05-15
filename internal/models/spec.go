package models

import "time"

type Machine struct {
	ID           string    `json:"id"`
	Hostname     string    `json:"hostname"`
	LastReported time.Time `json:"last_reported"`
	IPAddress    string    `json:"ip_address"`

	OS  OSInfo  `json:"os"`
	CPU CPUInfo `json:"cpu"`
	RAM RAMInfo `json:"ram"`
	Disks []DiskInfo `json:"disks"`
	GPU  []GPUInfo  `json:"gpu"`

	Motherboard *MotherboardInfo `json:"motherboard,omitempty"`
	BIOS        *BIOSInfo        `json:"bios,omitempty"`
	Network     []NetAdapter     `json:"network,omitempty"`
	Monitors    []MonitorInfo    `json:"monitors,omitempty"`
	Software    []SoftwareInfo   `json:"software,omitempty"`
	Sensors     *SensorInfo      `json:"sensors,omitempty"`
	System      *SystemInfo      `json:"system,omitempty"`
	RAMSlots    []RAMSlot        `json:"ram_slots,omitempty"`
	Battery     *BatteryInfo     `json:"battery,omitempty"`
	BitLocker   *BitLockerInfo   `json:"bit_locker,omitempty"`
}

type OSInfo struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Architecture  string `json:"architecture"`
	BuildNumber   string `json:"build_number"`
	UBR           string `json:"ubr"`
	DisplayName   string `json:"display_name"`
	InstallDate   string `json:"install_date"`
	LastBoot      string `json:"last_boot"`
}

type CPUInfo struct {
	Model        string `json:"model"`
	Cores        int    `json:"cores"`
	Threads      int    `json:"threads"`
	MaxClock     string `json:"max_clock"`
	TDPW         int    `json:"tdp_w"`
	Architecture string `json:"architecture"`
}

type RAMInfo struct {
	TotalGB    float64 `json:"total_gb"`
	FormFactor string  `json:"form_factor,omitempty"`
	Slots      int     `json:"slots,omitempty"`
}

type DiskInfo struct {
	Model      string  `json:"model"`
	SizeGB     float64 `json:"size_gb"`
	Type       string  `json:"type"`
	MountPoint string  `json:"mount_point"`
	HealthPct  int     `json:"health_pct"`
}

type GPUInfo struct {
	Brand  string `json:"brand"`
	Model  string `json:"model"`
	VRAMMB int    `json:"vram_mb"`
	Driver string `json:"driver"`
}

type MotherboardInfo struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Serial       string `json:"serial"`
}

type BIOSInfo struct {
	Vendor  string `json:"vendor"`
	Version string `json:"version"`
	Date    string `json:"date"`
}

type NetAdapter struct {
	Name        string   `json:"name"`
	MAC         string   `json:"mac"`
	IPAddresses []string `json:"ip_addresses"`
	SpeedMbps   int      `json:"speed_mbps,omitempty"`
}

type MonitorInfo struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Resolution   string `json:"resolution"`
	Serial       string `json:"serial"`
}

type SoftwareInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Publisher   string `json:"publisher"`
	InstallDate string `json:"install_date"`
}

type SensorInfo struct {
	CPUTemp   float64 `json:"cpu_temp,omitempty"`
	GPUTemp   float64 `json:"gpu_temp,omitempty"`
	CPUFanRPM int     `json:"cpu_fan_rpm,omitempty"`
	GPUFanRPM int     `json:"gpu_fan_rpm,omitempty"`
}

type SystemInfo struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Serial       string `json:"serial"`
	UserName     string `json:"user_name"`
}

type RAMSlot struct {
	SizeGB       float64 `json:"size_gb"`
	SpeedMHz     int     `json:"speed_mhz"`
	MemoryType   string  `json:"memory_type"`
	Manufacturer string  `json:"manufacturer"`
	PartNumber   string  `json:"part_number"`
}

type BatteryInfo struct {
	Present         bool   `json:"present"`
	Capacity        int    `json:"capacity"`
	DesignCapacity  int    `json:"design_capacity"`
	Chemistry       string `json:"chemistry"`
	Status          string `json:"status"`
}

type BitLockerInfo struct {
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
}
