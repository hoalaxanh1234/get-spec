package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"spec-collector/internal/agent"
	"spec-collector/internal/models"
)

func main() {
	m, err := agent.Gather()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error gathering spec: %v\n", err)
		pause()
		os.Exit(1)
	}

	output := agent.FormatSpec(m)
	fmt.Print(output)

	if runtime.GOOS == "windows" {
		if path := generateHTML(m); path != "" {
			fmt.Printf("📄 Report: %s\n", path)
			exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", path).Start()
		}
		interactive(m, output)
	} else {
		pause()
	}
}

func interactive(m *models.Machine, output string) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n[C] Copy  [S] Save to file  [R] HTML report  [E] Edit  [Enter] Exit: ")
		if !scanner.Scan() {
			return
		}
		cmd := strings.ToLower(strings.TrimSpace(scanner.Text()))
		switch cmd {
		case "c":
			if copyToClipboard(output) {
				fmt.Println("✓ Copied to clipboard!")
			} else {
				fmt.Println("✗ Clipboard failed")
			}
		case "s":
			if path := saveToFile(m, output); path != "" {
				fmt.Printf("✓ Saved to %s\n", path)
			} else {
				fmt.Println("✗ Save failed")
			}
		case "r":
			if path := generateHTML(m); path != "" {
				fmt.Printf("✓ Report: %s\n", path)
				exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", path).Start()
			} else {
				fmt.Println("✗ Report failed")
			}
		case "e":
			if editInEditor(output) {
				fmt.Println("✓ Opened in editor")
			} else {
				fmt.Println("✗ Editor failed")
			}
		case "":
			return
		}
	}
}

func copyToClipboard(text string) bool {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("clip")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run() == nil
	default:
		for _, prog := range []string{"xclip", "xsel", "wl-copy"} {
			cmd := exec.Command(prog)
			cmd.Stdin = strings.NewReader(text)
			if cmd.Run() == nil {
				return true
			}
		}
		return false
	}
}

func editInEditor(text string) bool {
	tmp := filepath.Join(os.TempDir(),
		fmt.Sprintf("spec-%s.txt", time.Now().Format("20060102-150405")))
	if err := os.WriteFile(tmp, []byte(text), 0644); err != nil {
		return false
	}

	switch runtime.GOOS {
	case "windows":
		return exec.Command("notepad.exe", tmp).Start() == nil
	default:
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano"
		}
		return exec.Command(editor, tmp).Start() == nil
	}
}

func saveToFile(m *models.Machine, output string) string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exe)
	hostname := m.Hostname
	if hostname == "" {
		hostname = "unknown"
	}
	name := fmt.Sprintf("spec-%s-%s.txt", hostname, time.Now().Format("20060102-150405"))
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(output), 0644); err != nil {
		return ""
	}
	return path
}

func generateHTML(m *models.Machine) string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exe)
	hostname := m.Hostname
	if hostname == "" {
		hostname = "unknown"
	}
	name := fmt.Sprintf("spec-%s-%s.html", hostname, time.Now().Format("20060102-150405"))
	path := filepath.Join(dir, name)

	html := buildHTML(m)
	if err := os.WriteFile(path, []byte(html), 0644); err != nil {
		return ""
	}
	return path
}

func buildHTML(m *models.Machine) string {
	s := strings.Builder{}
	now := time.Now().Format("02/01/2006 15:04")

	s.WriteString(htmlHead)
	s.WriteString(fmt.Sprintf("<title>%s - System Spec</title></head><body>\n", m.Hostname))
	manu, model, serial := "", "", ""
	if m.System != nil {
		manu = m.System.Manufacturer
		model = m.System.Model
		serial = m.System.Serial
	}
	if m.Motherboard != nil {
		if manu == "" || isGeneric(manu) { manu = m.Motherboard.Manufacturer }
		if model == "" || isGeneric(model) { model = m.Motherboard.Model }
		if serial == "" || isGeneric(serial) { serial = m.Motherboard.Serial }
	}

	s.WriteString(fmt.Sprintf("<h1>🖥 %s</h1>\n", m.Hostname))
	s.WriteString(fmt.Sprintf("<p class=\"subtitle\">Report generated: %s</p>\n", now))

	s.WriteString(`<div class="card">`)
	s.WriteString(`<h2>📋 THÔNG TIN MÁY</h2><table>`)
	if manu != "" { addRow(&s, "Hãng SX", manu) }
	if model != "" { addRow(&s, "Model", model) }
	if serial != "" { addRow(&s, "Service Tag", serial) }
	osName := m.OS.Name
	if m.OS.DisplayName != "" { osName = m.OS.DisplayName }
	osStr := fmt.Sprintf("%s (%s-bit) - Build %s.%s", osName, m.OS.Architecture, m.OS.BuildNumber, m.OS.UBR)
	addRow(&s, "Hệ điều hành", osStr)
	if m.BIOS != nil {
		addRow(&s, "BIOS", fmt.Sprintf("%s %s (%s)", m.BIOS.Vendor, m.BIOS.Version, m.BIOS.Date))
	}
	s.WriteString(`</table></div>`)

	s.WriteString(`<div class="card">`)
	s.WriteString(`<h2>💻 BỘ XỬ LÝ - CPU</h2><table>`)
	addRow(&s, "Model", m.CPU.Model)
	cpuExtra := fmt.Sprintf("%d Nhân / %d Luồng @ Max %s GHz", m.CPU.Cores, m.CPU.Threads, m.CPU.MaxClock)
	if m.CPU.TDPW > 0 { cpuExtra += fmt.Sprintf(" (TDP: %dW)", m.CPU.TDPW) }
	addRow(&s, "Cores/Threads", cpuExtra)
	s.WriteString(`</table></div>`)

	if len(m.GPU) > 0 {
		s.WriteString(`<div class="card"><h2>🎮 CARD ĐỒ HỌA - GPU</h2><table>`)
		for _, g := range m.GPU {
			vram := ""
			if g.VRAMMB > 0 { vram = fmt.Sprintf(" - %d GB", g.VRAMMB/1024) }
			drv := ""
			if g.Driver != "" { drv = fmt.Sprintf(" [Dr: %s]", g.Driver) }
			addRow(&s, g.Model, fmt.Sprintf("%s%s", vram, drv))
		}
		s.WriteString(`</table></div>`)
	}

	s.WriteString(`<div class="card">`)
	ramHeader := fmt.Sprintf("🧠 BỘ NHỚ RAM - %.1f GB", m.RAM.TotalGB)
	if m.RAM.Slots > 0 {
		ramHeader += fmt.Sprintf(" (%d/%d khe)", len(m.RAMSlots), m.RAM.Slots)
	}
	s.WriteString(fmt.Sprintf("<h2>%s</h2><table>", ramHeader))
	for i, slot := range m.RAMSlots {
		line := fmt.Sprintf("%.0fGB %s @ %dMHz", slot.SizeGB, slot.MemoryType, slot.SpeedMHz)
		if slot.PartNumber != "" && slot.PartNumber != "Undefined" {
			line += " | " + slot.PartNumber
		}
		addRow(&s, fmt.Sprintf("Slot %d", i+1), line)
	}
	emptySlots := m.RAM.Slots - len(m.RAMSlots)
	if emptySlots > 0 {
		addRow(&s, "Còn trống", fmt.Sprintf("%d khe", emptySlots))
	}
	s.WriteString(`</table></div>`)

	if len(m.Disks) > 0 {
		s.WriteString(`<div class="card"><h2>💾 Ổ LƯU TRỮ</h2><table>`)
		for i, d := range m.Disks {
			addRow(&s, fmt.Sprintf("Disk %d", i), fmt.Sprintf("%s - %.0f GB", d.Model, d.SizeGB))
		}
		s.WriteString(`</table></div>`)
	}

	s.WriteString(`<div class="card"><h2>🔒 MẠNG & BẢO MẬT</h2><table>`)
	if m.IPAddress != "" { addRow(&s, "IP LAN/Wi-Fi", m.IPAddress) }
	if len(m.Network) > 0 && m.Network[0].MAC != "" { addRow(&s, "MAC Address", m.Network[0].MAC) }
	if m.BitLocker != nil { addRow(&s, "BitLocker", m.BitLocker.Status) }
	if m.Battery != nil && m.Battery.Present {
		addRow(&s, "Pin", fmt.Sprintf("%d%%", m.Battery.Capacity))
	} else {
		addRow(&s, "Pin", "N/A")
	}
	s.WriteString(`</table></div>`)

	s.WriteString(`<div class="footer">`)
	s.WriteString(`<p>Tool created by V-Computer</p>`)
	s.WriteString(`<p>SDT: 038.928.6768 - Vương Nguyễn (software engineer)</p>`)
	s.WriteString(`<p>Địa chỉ: 14/1 Nguyễn Đình Chiểu p9 Đà Lạt</p>`)
	s.WriteString(`</div>`)
	s.WriteString(`</body></html>`)

	return s.String()
}

func isGeneric(s string) bool {
	if s == "" { return true }
	lower := strings.ToLower(s)
	for _, g := range []string{"system product name", "product name", "to be filled by o.e.m.",
		"default string", "not specified", "system serial number", "serial number", "none", "0000000"} {
		if g == lower { return true }
	}
	return false
}

func addRow(s *strings.Builder, label, value string) {
	if value == "" || value == "N/A" || value == "0 GB" {
		return
	}
	s.WriteString(fmt.Sprintf("<tr><td class=\"label\">%s</td><td>%s</td></tr>\n", label, value))
}

const htmlHead = `<!DOCTYPE html>
<html lang="vi">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#f0f2f5;color:#333;padding:20px;max-width:860px;margin:0 auto}
h1{font-size:24px;margin-bottom:4px;color:#1a1a2e}
.subtitle{font-size:13px;color:#888;margin-bottom:24px}
.card{background:#fff;border-radius:12px;padding:16px 20px;margin-bottom:16px;box-shadow:0 1px 3px rgba(0,0,0,.08)}
h2{font-size:16px;margin-bottom:12px;color:#1a1a2e;border-bottom:2px solid #e8e8e8;padding-bottom:8px}
table{width:100%;border-collapse:collapse}
td{padding:6px 0;font-size:14px;border-bottom:1px solid #f0f0f0}
td.label{width:140px;font-weight:600;color:#555}
.footer{text-align:center;color:#aaa;font-size:12px;margin-top:20px}
</style>
`

func pause() {
	if runtime.GOOS == "windows" {
		fmt.Print("\nNhấn Enter để thoát...")
		fmt.Scanln()
	}
}
