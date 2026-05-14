# spec-collector — Full Computer Spec Tool

## Architecture

Standalone agent. No server needed. Run `agent.exe` on any Windows machine, it prints full spec to terminal.

```
agent.exe
↓
Gather:  WMI + gopsutil
         → CPU, RAM, OS, disks, GPU, mobo, BIOS,
           network, monitors, RAM slots, battery
↓
Print:   Beautiful formatted terminal output
         (Vietnamese labels, section-based layout)
↓
Exit
```

## Tech Stack

| Layer | Choice |
|-------|--------|
| Language | Go 1.22+ |
| HW queries | `gopsutil/v3` + `StackExchange/wmi` |
| Output | Terminal with ASCII formatting |
| Build | Single .exe, no dependencies, no CGO |

## Project Structure

```
get-spec/
├── main.go                       # Entry point
├── go.mod / go.sum
├── Makefile
├── README.md
├── internal/
│   ├── models/
│   │   └── spec.go               # Data structures
│   ├── agent/
│   │   ├── basic.go              # Gather CPU, RAM, OS, disks, GPU, system info
│   │   ├── detailed.go           # Gather mobo, BIOS, network, monitors, RAM slots, battery
│   │   └── formatter.go          # Beautiful terminal output formatting
│   └── platform/
│       ├── wmi.go                # WMI stub (non-Windows)
│       └── wmi_windows.go        # WMI implementation (Windows)
```

## Build

```bash
# Build for current platform
make

# Build Windows agent (cross-compile)
make win64

# Build Linux agent
make linux

# Run
make run
```

## Output Format

```
==============================================================
          THÔNG TIN HỆ THỐNG - DESKTOP-4UE0421
                Ngày xuất: 14/05/2026 23:21
==============================================================

[THÔNG TIN MÁY]
  Tên máy:      DESKTOP-4UE0421
  User:         Administrator
  IP:           192.168.1.37
  Hãng SX:      ASUS (ASUS)
  Model:        System Product Name
  Service Tag:  System Serial Number
  Hệ điều hành: Windows 11 Pro 23H2 (64-bit) - Build 22631.6199
  BIOS:         American Megatrends Inc. 4505 (2025-11-28)
  Web hỗ trợ:   https://www.asus.com/support/searchproduct?searchKey=System%20Product%20Name

[BỘ XỬ LÝ - CPU]
  12th Gen Intel Core i9-12900K
  16 Nhân / 24 Luồng @ 3.2 GHz

[CARD ĐỒ HỌA - GPU]
  Intel(R) UHD Graphics 770 (2 GB) [Dr: 32.0.101.7082]
  NVIDIA GeForce RTX 5080 [Dr: 32.0.15.9621]

[BỘ NHỚ RAM - Tổng: 47.7 GB | 3/4 khe]
  Slot 1: 16GB DDR5 @ 6000MHz | F5-6400J3239G16G
  Slot 2: 16GB DDR5 @ 6000MHz | D5U1662320B-K66
  Slot 3: 16GB DDR5 @ 6000MHz | F5-6400J3239G16G
  >>> Còn trống: 1 khe

[Ổ LƯU TRỮ - 3 ổ]
  Disk 0: INTEL SSDPEKNW512G8 - 477 GB
  Disk 1: INTEL SSDPEKNW010T8 - 954 GB
  Disk 2: SK hynix PC801 HFS001TEJ9X101N - 954 GB

==============================================================
[MẠNG & BẢO MẬT]
  IP LAN/Wi-Fi: 192.168.1.37
  MAC Address:  00:91:9E:7C:5E:D4
  BitLocker:    Tắt 🔓
  Pin (Battery): N/A - Chai: N/A
```

## Dependencies

- `github.com/shirou/gopsutil/v3` — Cross-platform hardware info
- `github.com/StackExchange/wmi` — WMI queries (Windows only)
- `golang.org/x/sys` — Low-level system interfaces
