package agent

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"spec-collector/internal/models"
)

//go:embed templates/*
var templateFS embed.FS

type gpuRow struct {
	Name   string
	VRAM   string
	Driver string
}

type ramSlotRow struct {
	Slot   string
	Detail string
}

type diskRow struct {
	Label string
	Info  string
}

type reportData struct {
	Hostname     string
	Date         string
	Manufacturer string
	Model        string
	Serial       string
	OS           string
	BIOS         string
	CPUModel     string
	CPUExtra     string
	GPUs         []gpuRow
	RAMHeader    string
	RAMTotal     string
	RAMSlots     []ramSlotRow
	EmptySlots   int
	Disks        []diskRow
	IP           string
	MAC          string
	BitLocker    string
	Battery      string
	MachineJSON    string
	MinimalContent template.HTML
	Html2Canvas    template.JS
}

type minimalData struct {
	Model  string
	CPU    string
	CPU2   string
	RAM    string
	RAM2   string
	GPUs   []string
	Disks  []string
	OS     string
}

var tmplReport = template.Must(
	template.ParseFS(templateFS, "templates/report.html"),
)
var tmplMinimal = template.Must(
	template.ParseFS(templateFS, "templates/minimal.html"),
)

func buildMinimalHTML(m *models.Machine) string {
	var gpus []string
	for _, g := range m.GPU {
		vram := ""
		if g.VRAMMB > 0 {
			vram = fmt.Sprintf(" - %d GB", g.VRAMMB/1024)
		}
		gpus = append(gpus, g.Model+vram)
	}

	var disks []string
	for _, d := range m.Disks {
		disks = append(disks, fmt.Sprintf("%s - %.0f GB", d.Model, d.SizeGB))
	}

	var ramSlots []string
	for _, s := range m.RAMSlots {
		ramSlots = append(ramSlots, fmt.Sprintf("%.0fGB", s.SizeGB))
	}
	ramTotal := fmt.Sprintf("%.1f GB (%s)", m.RAM.TotalGB, strings.Join(ramSlots, ", "))
	ramSlotsStr := ""
	if m.RAM.Slots > 0 {
		ramSlotsStr = fmt.Sprintf("%d/%d", len(m.RAMSlots), m.RAM.Slots)
	}

	model := ""
	manu := ""
	if m.System != nil {
		manu = m.System.Manufacturer
		model = m.System.Model
	}
	if m.Motherboard != nil && (manu == "" || isGeneric(manu)) {
		manu = m.Motherboard.Manufacturer
	}
	if m.Motherboard != nil && (model == "" || isGeneric(model)) {
		model = m.Motherboard.Model
	}
	if isGeneric(model) {
		model = ""
	}
	if model != "" && !isGeneric(manu) {
		model = manu + " " + model
	}

	cpu2 := ""
	if m.CPU.Cores > 0 {
		cpu2 = fmt.Sprintf("%d nhân / %d luồng", m.CPU.Cores, m.CPU.Threads)
	}

	osName := m.OS.Name
	if m.OS.DisplayName != "" {
		osName = m.OS.DisplayName
	}

	data := minimalData{
		Model: model,
		CPU:   m.CPU.Model,
		CPU2:  cpu2,
		RAM:   ramTotal,
		RAM2:  ramSlotsStr,
		GPUs:  gpus,
		Disks: disks,
		OS:    fmt.Sprintf("%s (%s-bit)", osName, m.OS.Architecture),
	}

	var buf strings.Builder
	tmplMinimal.Execute(&buf, data)
	return buf.String()
}

func buildHTML(m *models.Machine) string {
	now := time.Now().Format("02/01/2006 15:04")

	manu, model, serial := "", "", ""
	if m.System != nil {
		manu = m.System.Manufacturer
		model = m.System.Model
		serial = m.System.Serial
	}
	if m.Motherboard != nil {
		if manu == "" || isGeneric(manu) {
			manu = m.Motherboard.Manufacturer
		}
		if model == "" || isGeneric(model) {
			model = m.Motherboard.Model
		}
		if serial == "" || isGeneric(serial) {
			serial = m.Motherboard.Serial
		}
	}

	osName := m.OS.Name
	if m.OS.DisplayName != "" {
		osName = m.OS.DisplayName
	}
	osStr := fmt.Sprintf("%s (%s-bit) - Build %s.%s", osName, m.OS.Architecture, m.OS.BuildNumber, m.OS.UBR)

	biosStr := ""
	if m.BIOS != nil {
		biosStr = fmt.Sprintf("%s %s (%s)", m.BIOS.Vendor, m.BIOS.Version, m.BIOS.Date)
	}

	cpuExtra := fmt.Sprintf("%d Nhân / %d Luồng @ Max %s GHz", m.CPU.Cores, m.CPU.Threads, m.CPU.MaxClock)
	if m.CPU.TDPW > 0 {
		cpuExtra += fmt.Sprintf(" (TDP: %dW)", m.CPU.TDPW)
	}

	var gpus []gpuRow
	for _, g := range m.GPU {
		vram := ""
		if g.VRAMMB > 0 {
			vram = fmt.Sprintf(" - %d GB", g.VRAMMB/1024)
		}
		drv := ""
		if g.Driver != "" {
			drv = fmt.Sprintf(" [Dr: %s]", g.Driver)
		}
		displayModel := g.Model
		if g.Brand != "" && !strings.HasPrefix(g.Model, g.Brand) {
			displayModel = g.Brand + " " + g.Model
		}
		gpus = append(gpus, gpuRow{Name: displayModel, VRAM: vram, Driver: drv})
	}

	ramHeader := fmt.Sprintf("🧠 BỘ NHỚ RAM - %.1f GB", m.RAM.TotalGB)
	if m.RAM.Slots > 0 {
		ramHeader += fmt.Sprintf(" (%d/%d khe)", len(m.RAMSlots), m.RAM.Slots)
	}
	ramTotal := fmt.Sprintf("%.1f GB", m.RAM.TotalGB)

	var ramSlots []ramSlotRow
	for i, slot := range m.RAMSlots {
		line := fmt.Sprintf("%.0fGB %s @ %dMHz", slot.SizeGB, slot.MemoryType, slot.SpeedMHz)
		if slot.PartNumber != "" && slot.PartNumber != "Undefined" {
			line += " | " + slot.PartNumber
		}
		ramSlots = append(ramSlots, ramSlotRow{Slot: fmt.Sprintf("Slot %d", i+1), Detail: line})
	}
	emptySlots := m.RAM.Slots - len(m.RAMSlots)
	if emptySlots < 0 {
		emptySlots = 0
	}

	var disks []diskRow
	for i, d := range m.Disks {
		disks = append(disks, diskRow{
			Label: fmt.Sprintf("Disk %d", i),
			Info:  fmt.Sprintf("%s - %.0f GB", d.Model, d.SizeGB),
		})
	}

	bitLocker := "Tắt 🔓"
	if m.BitLocker != nil {
		bitLocker = m.BitLocker.Status
	}

	battery := "N/A"
	if m.Battery != nil && m.Battery.Present {
		battery = fmt.Sprintf("%d%%", m.Battery.Capacity)
	}

	mac := ""
	if len(m.Network) > 0 {
		mac = m.Network[0].MAC
	}

	h2cBytes, _ := templateFS.ReadFile("templates/html2canvas.min.js")

	data := reportData{
		Hostname:       m.Hostname,
		Date:           now,
		Manufacturer:   manu,
		Model:          model,
		Serial:         serial,
		OS:             osStr,
		BIOS:           biosStr,
		CPUModel:       m.CPU.Model,
		CPUExtra:       cpuExtra,
		GPUs:           gpus,
		RAMHeader:      ramHeader,
		RAMTotal:       ramTotal,
		RAMSlots:       ramSlots,
		EmptySlots:     emptySlots,
		Disks:          disks,
		IP:             m.IPAddress,
		MAC:            mac,
		BitLocker:      bitLocker,
		Battery:        battery,
		MachineJSON:    machineJSON(m),
		MinimalContent: template.HTML(buildMinimalHTML(m)),
		Html2Canvas:    template.JS(h2cBytes),
	}

	var buf strings.Builder
	tmplReport.Execute(&buf, data)
	return buf.String()
}

func machineJSON(m *models.Machine) string {
	type data struct {
		Hostname string   `json:"hostname"`
		Date     string   `json:"date"`
		CPU      string   `json:"cpu"`
		CPU2     string   `json:"cpu2"`
		RAM      string   `json:"ram"`
		RAM2     string   `json:"ram2"`
		GPU      []string `json:"gpu"`
		Disk     []string `json:"disk"`
		OS       string   `json:"os"`
	}
	d := data{
		Hostname: m.Hostname,
		Date:     time.Now().Format("02/01/2006 15:04"),
		CPU:      m.CPU.Model,
	}
	if m.CPU.MaxClock != "" {
		d.CPU2 = fmt.Sprintf("%dC/%dT @ %s GHz", m.CPU.Cores, m.CPU.Threads, m.CPU.MaxClock)
	}
	var ramSlots []string
	for _, s := range m.RAMSlots {
		ramSlots = append(ramSlots, fmt.Sprintf("%.0fGB", s.SizeGB))
	}
	d.RAM = fmt.Sprintf("%.1f GB (%s)", m.RAM.TotalGB, strings.Join(ramSlots, ", "))
	if m.RAM.Slots > 0 {
		d.RAM2 = fmt.Sprintf("%d/%d", len(m.RAMSlots), m.RAM.Slots)
	}
	for _, g := range m.GPU {
		vram := ""
		if g.VRAMMB > 0 {
			vram = fmt.Sprintf(" - %d GB", g.VRAMMB/1024)
		}
		d.GPU = append(d.GPU, g.Model+vram)
	}
	for _, disk := range m.Disks {
		d.Disk = append(d.Disk, fmt.Sprintf("%s - %.0f GB", disk.Model, disk.SizeGB))
	}
	osName := m.OS.Name
	if m.OS.DisplayName != "" {
		osName = m.OS.DisplayName
	}
	d.OS = fmt.Sprintf("%s (%s-bit)", osName, m.OS.Architecture)
	b, _ := json.Marshal(d)
	return string(b)
}

func GenerateHTML(m *models.Machine) string {
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
