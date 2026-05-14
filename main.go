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

	s.WriteString(`<div class="header">`)
	s.WriteString(fmt.Sprintf(`<h1>🖥 %s</h1>`, m.Hostname))
	s.WriteString(fmt.Sprintf(`<p class="subtitle">Report generated: %s</p>`, now))
	s.WriteString(`</div>`)

	s.WriteString(`<div class="report-content">`)

	s.WriteString(`<div class="author-bar">`)
	s.WriteString(`<div class="author-item"><span class="author-label">Cung cấp bởi</span><span class="author-value">V-Computer</span></div>`)
	s.WriteString(`<div class="author-item"><span class="author-label">Liên hệ</span><span class="author-value">038.928.6768 - Vương Nguyễn</span></div>`)
	s.WriteString(`<div class="author-item"><span class="author-label">Địa chỉ</span><span class="author-value">14/1 Nguyễn Đình Chiểu, P9, Đà Lạt</span></div>`)
	s.WriteString(`</div>`)

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

	s.WriteString(`</div>`)

	s.WriteString(`<div class="toolbar">`)
	s.WriteString(`<button onclick="saveImage()">🖼 Lưu thành hình ảnh</button>`)
	s.WriteString(`<button onclick="saveMinimal()">📋 Lưu phiên bản tối giản</button>`)
	s.WriteString(`</div>`)

	s.WriteString(`<div class="footer">`)
	s.WriteString(`<p>Tool created by <strong>V-Computer</strong></p>`)
	s.WriteString(`<p>SDT: 038.928.6768 - Vương Nguyễn (software engineer)</p>`)
	s.WriteString(`<p>Địa chỉ: 14/1 Nguyễn Đình Chiểu, P9, Đà Lạt</p>`)
	s.WriteString(`</div>`)

	s.WriteString(buildMinimalHTML(m))
	s.WriteString(`<a href="https://www.facebook.com/nguyenvanvuong972/" target="_blank" class="fb-btn" title="Facebook Vương Nguyễn"><svg viewBox="0 0 24 24" fill="white"><path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/></svg></a>`)
	s.WriteString(`<script src="https://cdn.jsdelivr.net/npm/html2canvas@1.4.1/dist/html2canvas.min.js"></script>`)
	s.WriteString(fmt.Sprintf(`<script>
function saveImage(){html2canvas(document.querySelector(".report-content"),{scale:2,useCORS:true,backgroundColor:"#ffffff"}).then(function(c){var a=document.createElement("a");a.href=c.toDataURL("image/png");a.download="%s-spec.png";a.click()})}
function saveMinimal(){var e=document.getElementById("minimal-report");e.style.display="block";html2canvas(e,{scale:2,useCORS:true,backgroundColor:"#ffffff"}).then(function(c){e.style.display="none";var a=document.createElement("a");a.href=c.toDataURL("image/png");a.download="%s-minimal.png";a.click()})}
</script>`, m.Hostname, m.Hostname))
	s.WriteString(`</body></html>`)

	return s.String()
}

func buildMinimalHTML(m *models.Machine) string {
	s := strings.Builder{}
	s.WriteString(`<div id="minimal-report"><div class="minimal-inner">`)
	s.WriteString(fmt.Sprintf(`<div class="minimal-title">🖥 %s</div>`, m.Hostname))
	s.WriteString(fmt.Sprintf(`<div class="minimal-date">%s</div>`, time.Now().Format("02/01/2006 15:04")))
	s.WriteString(`<table class="minimal-table">`)
	minimalRow(&s, "CPU", m.CPU.Model)
	if m.CPU.MaxClock != "" {
		minimalRow(&s, "", fmt.Sprintf("%dC/%dT @ %s GHz", m.CPU.Cores, m.CPU.Threads, m.CPU.MaxClock))
	}
	minimalRow(&s, "RAM", fmt.Sprintf("%.1f GB (", m.RAM.TotalGB))
	for i, slot := range m.RAMSlots {
		if i > 0 { s.WriteString(", ") }
		s.WriteString(fmt.Sprintf("%.0fGB", slot.SizeGB))
	}
	s.WriteString(")")
	if m.RAM.Slots > 0 { minimalRow(&s, "Khe", fmt.Sprintf("%d/%d", len(m.RAMSlots), m.RAM.Slots)) }
	for _, g := range m.GPU {
		vram := ""
		if g.VRAMMB > 0 { vram = fmt.Sprintf(" - %d GB", g.VRAMMB/1024) }
		minimalRow(&s, "GPU", g.Model+vram)
	}
	for i, d := range m.Disks {
		minimalRow(&s, fmt.Sprintf("Disk %d", i), fmt.Sprintf("%s - %.0f GB", d.Model, d.SizeGB))
	}
	osName := m.OS.Name
	if m.OS.DisplayName != "" { osName = m.OS.DisplayName }
	minimalRow(&s, "OS", fmt.Sprintf("%s (%s-bit)", osName, m.OS.Architecture))
	s.WriteString(`</table>`)
	s.WriteString(`<div class="minimal-contact">V-Computer • 038.928.6768</div>`)
	s.WriteString(`</div></div>`)
	return s.String()
}

func minimalRow(s *strings.Builder, label, value string) {
	if value == "" {
		return
	}
	if label == "" {
		s.WriteString(fmt.Sprintf(`<tr><td colspan="2" class="mval">%s</td></tr>`, value))
	} else {
		s.WriteString(fmt.Sprintf(`<tr><td class="mlabel">%s</td><td class="mval">%s</td></tr>`, label, value))
	}
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
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:linear-gradient(135deg,#e8edf5 0%,#f5f7fa 100%);color:#333;padding:20px;max-width:860px;margin:0 auto}
.header{background:linear-gradient(135deg,#1a1a2e 0%,#16213e 100%);color:#fff;border-radius:12px;padding:20px 24px;margin-bottom:16px}
.header h1{font-size:24px;margin-bottom:2px}
.header .subtitle{font-size:13px;color:#8899b0;margin:0}
.author-bar{background:#fff;border-radius:12px;padding:12px 20px;margin-bottom:16px;box-shadow:0 1px 3px rgba(0,0,0,.08);display:flex;flex-wrap:wrap;gap:16px 32px}
.author-item{display:flex;flex-direction:column;gap:1px}
.author-label{font-size:11px;color:#999;text-transform:uppercase;letter-spacing:.5px}
.author-value{font-size:14px;font-weight:600;color:#1a1a2e}
.card{background:#fff;border-radius:12px;padding:16px 20px;margin-bottom:16px;box-shadow:0 1px 3px rgba(0,0,0,.08)}
h2{font-size:16px;margin-bottom:12px;color:#1a1a2e;border-bottom:2px solid #e8e8e8;padding-bottom:8px}
table{width:100%;border-collapse:collapse}
td{padding:6px 0;font-size:14px;border-bottom:1px solid #f0f0f0}
td.label{width:140px;font-weight:600;color:#555;white-space:nowrap}
.toolbar{display:flex;gap:10px;margin-bottom:16px;flex-wrap:wrap}
.toolbar button{flex:1;min-width:180px;padding:12px 16px;border:none;border-radius:10px;font-size:14px;font-weight:600;cursor:pointer;transition:transform .15s,box-shadow .15s;color:#fff}
.toolbar button:first-child{background:linear-gradient(135deg,#2563eb,#1d4ed8)}
.toolbar button:last-child{background:linear-gradient(135deg,#059669,#047857)}
.toolbar button:hover{transform:translateY(-1px);box-shadow:0 4px 12px rgba(0,0,0,.15)}
.footer{text-align:center;color:#999;font-size:12px;margin-top:20px;border-top:1px solid #ddd;padding-top:16px}
.footer p{margin:2px 0}
#minimal-report{display:none;position:fixed;top:0;left:0;width:400px;background:#fff;padding:12px;font-family:'Segoe UI',Arial,sans-serif;z-index:-1}
.minimal-inner{border:2px solid #1a1a2e;border-radius:6px;padding:10px 12px}
.minimal-title{font-size:15px;font-weight:700;color:#1a1a2e;margin-bottom:1px}
.minimal-date{font-size:10px;color:#999;margin-bottom:6px}
.minimal-table{width:100%;border-collapse:collapse;font-size:11px}
.minimal-table td{padding:2px 4px;border-bottom:1px solid #eee;line-height:1.3}
.minimal-table td.mlabel{width:48px;font-weight:700;color:#555;white-space:nowrap;font-size:10px;text-transform:uppercase}
.minimal-table td.mval{color:#1a1a2e}
.minimal-contact{text-align:center;font-size:10px;color:#888;margin-top:6px;padding-top:4px;border-top:1px solid #ddd}
.fb-btn{position:fixed;right:12px;top:50%;transform:translateY(-50%);width:48px;height:48px;background:#1877f2;border-radius:50%;display:flex;align-items:center;justify-content:center;box-shadow:0 4px 12px rgba(24,119,242,.4);transition:transform .2s,box-shadow .2s;z-index:999}
.fb-btn:hover{transform:translateY(-50%) scale(1.1);box-shadow:0 6px 20px rgba(24,119,242,.5)}
.fb-btn svg{width:26px;height:26px}
</style>
`

func pause() {
	if runtime.GOOS == "windows" {
		fmt.Print("\nNhấn Enter để thoát...")
		fmt.Scanln()
	}
}
