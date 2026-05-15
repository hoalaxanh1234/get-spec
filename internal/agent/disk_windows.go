//go:build windows

package agent

import (
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"

	"spec-collector/internal/models"
	"spec-collector/internal/platform"
)

const (
	ioctlStorageQueryProperty = 0x002D1400
	storageDeviceProperty     = 0
	propertyStandardQuery     = 0
)

type storagePropertyQuery struct {
	PropertyId uint32
	QueryType  uint32
	_          [4]byte
}

type storageDeviceDescriptor struct {
	Version               uint32
	Size                  uint32
	DeviceType            byte
	DeviceTypeModifier    byte
	CommandQueueing       byte
	_                     byte
	VendorIdOffset        int32
	ProductIdOffset       int32
	ProductRevisionOffset int32
	SerialNumberOffset    int32
	BusType               int32
	RawPropertiesLength   uint32
}

func busTypeString(busType int32) string {
	switch busType {
	case 3, 11:
		return "SATA"
	case 7:
		return "USB"
	case 9:
		return "SAS"
	case 10, 17:
		return "NVMe"
	case 12:
		return "SD"
	case 13:
		return "eMMC"
	case 14:
		return "Virtual"
	default:
		return ""
	}
}

func getDisksIOCTL() []models.DiskInfo {
	var winDrives []struct {
		Model         string
		Size          uint64
		Index         uint32
		InterfaceType string
	}
	err := platform.QueryWMI(
		"SELECT Model, Size, Index, InterfaceType FROM Win32_DiskDrive",
		&winDrives,
	)
	if err != nil || len(winDrives) == 0 {
		return nil
	}

	var disks []models.DiskInfo
	ioctlOk := false
	for _, d := range winDrives {
		if d.Size == 0 {
			continue
		}

		busType := detectBusTypeIOCTL(d.Index)
		diskType := ""
		if busType != "" {
			ioctlOk = true
			diskType = classifyDiskType(busType, d.Model)
		} else {
			continue
		}

		disks = append(disks, models.DiskInfo{
			Model:  d.Model,
			SizeGB: float64(d.Size) / (1024 * 1024 * 1024),
			Type:   diskType,
		})
	}
	if !ioctlOk {
		return nil
	}
	return disks
}

func detectBusTypeIOCTL(index uint32) string {
	path := fmt.Sprintf(`\\.\PhysicalDrive%d`, index)
	handle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(path),
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(handle)

	query := storagePropertyQuery{
		PropertyId: storageDeviceProperty,
		QueryType:  propertyStandardQuery,
	}

	var buf [4096]byte
	var returned uint32

	err = windows.DeviceIoControl(
		handle,
		ioctlStorageQueryProperty,
		(*byte)(unsafe.Pointer(&query)),
		uint32(unsafe.Sizeof(query)),
		&buf[0],
		uint32(len(buf)),
		&returned,
		nil,
	)
	if err != nil {
		return ""
	}

	desc := (*storageDeviceDescriptor)(unsafe.Pointer(&buf[0]))
	return busTypeString(desc.BusType)
}

func classifyDiskType(busType, model string) string {
	m := strings.ToLower(model)
	switch busType {
	case "NVMe":
		gen := nvmeGen(model)
		if gen != "" {
			return "NVMe " + gen
		}
		return "NVMe"
	case "SATA":
		if strings.Contains(m, "ssd") {
			return "SATA SSD"
		}
		if strings.Contains(m, "hdd") || strings.Contains(m, "hard") {
			return "SATA HDD"
		}
		return "SATA SSD"
	case "USB":
		return "USB"
	case "SAS":
		return "SAS"
	default:
		if strings.Contains(m, "sata") {
			if strings.Contains(m, "ssd") {
				return "SATA SSD"
			}
			if strings.Contains(m, "hdd") || strings.Contains(m, "hard") {
				return "SATA HDD"
			}
			return "SATA SSD"
		}
		if strings.Contains(m, "nvme") {
			gen := nvmeGen(model)
			if gen != "" {
				return "NVMe " + gen
			}
			return "NVMe"
		}
		if strings.Contains(m, "ssd") {
			return "SSD"
		}
		return busType
	}
}
