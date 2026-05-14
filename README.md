# spec-collector — Full Computer Spec Tool

Standalone Windows agent. Run `agent.exe` — it gathers full hardware specs, prints to terminal, and opens a polished HTML report in your browser.

```
agent.exe
↓
Gather:  WMI + gopsutil + registry + nvidia-smi
         → CPU, RAM, OS, disks, GPU, mobo, BIOS,
           network, monitors, RAM slots, battery, BitLocker
↓
Print:   Beautiful terminal output (Vietnamese labels, box-drawing layout)
↓
Open:    HTML report in default browser with:
         • Save as image (PNG)
         • Customizable minimal sticker (select fields + add note)
         • Floating Facebook button
         • Author info header
```

## Tech Stack

| Layer | Choice |
|-------|--------|
| Language | Go 1.22+ |
| HW queries | `gopsutil/v3` + `StackExchange/wmi` + registry |
| Output | Terminal (UTF-8 box-drawing) + HTML report |
| Build | Single .exe, no dependencies, no CGO, no WebView2 |

## Project Structure

```
get-spec/
├── main.go                       # Entry point (gather, format, HTML, interactive menu)
├── go.mod / go.sum
├── Makefile
├── README.md
├── internal/
│   ├── models/
│   │   └── spec.go               # Data structures
│   ├── agent/
│   │   ├── basic.go              # Gather CPU, RAM, OS, disks, GPU, system info
│   │   ├── basic_windows.go      # GPU VRAM from registry (Windows)
│   │   ├── basic_other.go        # GPU VRAM stub (non-Windows)
│   │   ├── detailed.go           # Gather mobo, BIOS, network, monitors, RAM slots, battery, BitLocker
│   │   ├── gpu_nvidia.go         # nvidia-smi / rocm-smi CLI parsers (cross-platform)
│   │   └── formatter.go          # Terminal output formatting (Vietnamese, box-drawing)
│   └── platform/
│       ├── wmi.go                # WMI stub (non-Windows)
│       ├── wmi_windows.go        # WMI + QueryWMINamespace (Windows)
│       ├── registry.go           # Registry stub (non-Windows)
│       └── registry_windows.go   # Registry reads (Windows)
```

## Build

```bash
# Build Windows agent (cross-compile from Linux)
make win64

# Build Linux agent
make linux

# Build both
make all

# Run (Linux)
make run
```

## Features

### Terminal Output
- Vietnamese labels (THÔNG TIN MÁY, BỘ XỬ LÝ, CARD ĐỒ HỌA, ...)
- Unicode box-drawing separators
- BitLocker status, battery health, RAM slot info
- Author/contact info footer

### HTML Report
- Auto-opens in default browser
- Dark gradient header + author info bar
- **Save as image** — captures full report as PNG via html2canvas
- **Minimal sticker** — compact card for printing & taping on PC case
  - Checkboxes to select which fields appear
  - Optional custom note (shown italic)
  - Downloaded as PNG
- Floating Facebook button (right side, centered)
- Support link (manufacturer website)

### Interactive Menu (Windows)
After output, choose:
- `[C]` Copy to clipboard
- `[S]` Save to text file
- `[R]` Regenerate HTML report
- `[E]` Edit in Notepad
- `[Enter]` Exit

## Dependencies

- `github.com/shirou/gopsutil/v3` — Cross-platform hardware info
- `github.com/StackExchange/wmi` — WMI queries (Windows only)
- `golang.org/x/sys` — Low-level system interfaces
