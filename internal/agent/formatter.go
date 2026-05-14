package agent

import (
	"fmt"
	"strings"
	"time"

	"spec-collector/internal/models"
)

func FormatSpec(m *models.Machine) string {
	var b strings.Builder
	now := time.Now().Format("02/01/2006 15:04")
	hostname := m.Hostname

	sep := strings.Repeat("─", 52)

	b.WriteString(sep + "\n")
	b.WriteString(fmt.Sprintf("  %-48s\n", "THÔNG TIN HỆ THỐNG - "+hostname))
	b.WriteString(fmt.Sprintf("  %-48s\n", "Ngày xuất: "+now))
	b.WriteString(sep + "\n")

	b.WriteString(fmt.Sprintf("\n  📋 %s\n", "THÔNG TIN MÁY"))

	manu, model, serial := "", "", ""
	if m.System != nil {
		manu = pickValue(m.System.Manufacturer)
		model = pickValue(m.System.Model)
		serial = pickValue(m.System.Serial)
	}
	if m.Motherboard != nil && (isGeneric(manu) || isGeneric(model) || isGeneric(serial)) {
		manu = pickValue(manu, m.Motherboard.Manufacturer)
		model = pickValue(model, m.Motherboard.Model)
		serial = pickValue(serial, m.Motherboard.Serial)
	}
	if manu != "" {
		b.WriteString(kvLine("Hãng SX", manu))
	}
	if model != "" {
		b.WriteString(kvLine("Model", model))
	}
	if serial != "" {
		b.WriteString(kvLine("Service Tag", serial))
	}
	b.WriteString(kvLine("Hệ điều hành", formatOS(m.OS)))
	if m.BIOS != nil {
		b.WriteString(kvLine("BIOS", m.BIOS.Vendor+" "+m.BIOS.Version+" ("+m.BIOS.Date+")"))
	}

	b.WriteString(fmt.Sprintf("\n  💻 %s\n", "BỘ XỬ LÝ - CPU"))
	tdpStr := ""
	if m.CPU.TDPW > 0 {
		tdpStr = fmt.Sprintf(" (TDP: %dW)", m.CPU.TDPW)
	}
	b.WriteString(fmt.Sprintf("    %s\n", m.CPU.Model))
	b.WriteString(fmt.Sprintf("    %d Nhân / %d Luồng @ Max %s GHz%s\n",
		m.CPU.Cores, m.CPU.Threads, formatClock(m.CPU.MaxClock), tdpStr))

	if len(m.GPU) > 0 {
		b.WriteString(fmt.Sprintf("\n  🎮 %s\n", "CARD ĐỒ HỌA - GPU"))
		for _, g := range m.GPU {
			vramStr := ""
			if g.VRAMMB > 0 {
				vramStr = fmt.Sprintf(" - %s", formatVRAM(g.VRAMMB))
			}
			drvStr := ""
			if g.Driver != "" {
				drvStr = fmt.Sprintf(" [Dr: %s]", g.Driver)
			}
			displayModel := g.Model
			if g.Brand != "" && !strings.HasPrefix(g.Model, g.Brand) {
				displayModel = g.Brand + " " + g.Model
			}
			b.WriteString(fmt.Sprintf("    %s%s%s\n", displayModel, vramStr, drvStr))
		}
	}

	usedSlots := len(m.RAMSlots)
	emptySlots := m.RAM.Slots - usedSlots
	if emptySlots < 0 {
		emptySlots = 0
	}
	ramHeader := fmt.Sprintf("🧠 BỘ NHỚ RAM - %.1f GB", m.RAM.TotalGB)
	if m.RAM.Slots > 0 {
		ramHeader += fmt.Sprintf(" (%d/%d khe)", usedSlots, m.RAM.Slots)
	}
	b.WriteString(fmt.Sprintf("\n  %s\n", ramHeader))
	if len(m.RAMSlots) > 0 {
		for i, slot := range m.RAMSlots {
			line := fmt.Sprintf("    Slot %d: %.0fGB %s @ %dMHz", i+1, slot.SizeGB, slot.MemoryType, slot.SpeedMHz)
			if slot.PartNumber != "" && slot.PartNumber != "Undefined" {
				line += " | " + slot.PartNumber
			}
			b.WriteString(line + "\n")
		}
		if emptySlots > 0 {
			b.WriteString(fmt.Sprintf("    >>> Còn trống: %d khe\n", emptySlots))
		}
	} else {
		b.WriteString(fmt.Sprintf("    Tổng: %.1f GB\n", m.RAM.TotalGB))
	}

	if len(m.Disks) > 0 {
		b.WriteString(fmt.Sprintf("\n  💾 %s\n", "Ổ LƯU TRỮ"))
		for i, d := range m.Disks {
			sizeStr := fmt.Sprintf("%.0f GB", d.SizeGB)
			model := d.Model
			parts := strings.Split(model, "/")
			if len(parts) > 0 {
				model = parts[len(parts)-1]
			}
			b.WriteString(fmt.Sprintf("    Disk %d: %s - %s\n", i, model, sizeStr))
		}
	}

	b.WriteString(fmt.Sprintf("\n  🔒 %s\n", "MẠNG & BẢO MẬT"))
	if m.IPAddress != "" {
		b.WriteString(kvLine("IP LAN/Wi-Fi", m.IPAddress))
	}
	if len(m.Network) > 0 {
		b.WriteString(kvLine("MAC Address", m.Network[0].MAC))
	}
	if m.BitLocker != nil {
		b.WriteString(kvLine("BitLocker", m.BitLocker.Status))
	} else {
		b.WriteString(kvLine("BitLocker", "Tắt \U0001F513"))
	}
	if m.Battery != nil && m.Battery.Present {
		b.WriteString(kvLine("Pin (Battery)", fmt.Sprintf("%d%% - Chai: %s", m.Battery.Capacity, batteryHealth(m.Battery))))
	} else {
		b.WriteString(kvLine("Pin (Battery)", "N/A - Chai: N/A"))
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s\n", "Tool created by V-Computer"))
	b.WriteString(fmt.Sprintf("  %s\n", "SDT: 038.928.6768 - Vương Nguyễn (software engineer)"))
	b.WriteString(fmt.Sprintf("  %s\n", "Địa chỉ: 14/1 Nguyễn Đình Chiểu p9 Đà Lạt"))
	b.WriteString(fmt.Sprintf("  %s\n", "FB: https://www.facebook.com/nguyenvanvuong972/"))
	b.WriteString(sep + "\n")
	return b.String()
}

func kvLine(key, value string) string {
	return fmt.Sprintf("    %-16s%s\n", key+":", value)
}

var genericStrings = []string{
	"System Product Name", "Product Name", "To Be Filled By O.E.M.",
	"Default string", "Not Specified", "To be filled by O.E.M.",
	"System Serial Number", "Serial Number", "None", "0000000",
}

func isGeneric(s string) bool {
	if s == "" {
		return true
	}
	lower := strings.ToLower(s)
	for _, g := range genericStrings {
		if strings.ToLower(g) == lower {
			return true
		}
	}
	return false
}

func pickValue(vals ...string) string {
	for _, v := range vals {
		if !isGeneric(v) {
			return v
		}
	}
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func formatOS(os models.OSInfo) string {
	name := os.Name
	if os.DisplayName != "" {
		name = os.DisplayName
	}
	arch := strings.Replace(os.Architecture, "-bit", "", 1)
	s := fmt.Sprintf("%s (%s-bit)", name, arch)
	build := os.BuildNumber
	if build == "" && strings.Contains(os.Version, ".") {
		segments := strings.Split(os.Version, ".")
		if len(segments) >= 3 {
			build = segments[2]
		}
	}
	if build != "" {
		verStr := build
		if os.UBR != "" {
			verStr += "." + os.UBR
		}
		s += fmt.Sprintf(" - Build %s", verStr)
	}
	return s
}

func formatClock(mhz string) string {
	var f float64
	if n, _ := fmt.Sscanf(mhz, "%f", &f); n == 1 && f >= 1000 {
		return fmt.Sprintf("%.1f", f/1000)
	}
	return mhz
}

func formatVRAM(vramMB int) string {
	if vramMB <= 0 {
		return "0 GB"
	}
	if vramMB < 1024 {
		return fmt.Sprintf("%d MB", vramMB)
	}
	gb := (vramMB + 512) / 1024
	return fmt.Sprintf("%d GB", gb)
}

func batteryHealth(b *models.BatteryInfo) string {
	if b.DesignCapacity <= 0 {
		return "N/A"
	}
	pct := float64(b.Capacity) / float64(b.DesignCapacity) * 100
	return fmt.Sprintf("%.0f%%", pct)
}
