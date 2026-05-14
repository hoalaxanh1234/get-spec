# spec-collector — Full Computer Spec Tool

## Architecture

```
┌──────────────────────┐       HTTP POST (JSON)       ┌──────────────────────┐
│  Agent (Wails GUI)   │ ────────────────────────────  │  Central Server      │
│  Gather → Show → Send                                  │  (Go + SQLite)       │
└──────────────────────┘                                 │  - REST API          │
                                                          │  - Web Dashboard     │
                                                          │  - Serves agent .exe │
                                                          └──────────────────────┘
```

## Tech Stack

| Layer | Choice |
|-------|--------|
| Server | Go 1.22+, `chi` router, `modernc.org/sqlite` (pure Go, no CGO) |
| Agent | Go + Wails v2, single Windows .exe with native GUI |
| DB | SQLite (single file, portable) |
| Agent UI | Plain HTML + JavaScript + Tailwind CSS (CDN), WebView2 runtime |
| Server Frontend | Server-rendered HTML + vanilla JS + Tailwind CSS (CDN) |
| HW queries | `gopsutil/v3` + `StackExchange/wmi` + `golang.org/x/sys/windows/registry` |

## Project Structure

```
get-spec/
├── main.go                       # Wails GUI entry point (agent)
├── app.go                        # Wails Go bindings
├── go.mod / go.sum
├── Makefile
├── README.md
├── todo.md
├── wails.json                    # Wails project config
├── cmd/
│   └── server/
│       └── main.go               # Server CLI entry point
├── frontend/
│   └── dist/
│       ├── index.html            # Agent GUI page
│       └── src/
│           ├── main.js           # JS logic for spec display + send
│           └── style.css         # Custom styles
├── internal/
│   ├── models/
│   │   └── spec.go               # Shared data structures
│   ├── agent/
│   │   ├── basic.go              # Basic spec: CPU, RAM, OS, disk, GPU
│   │   ├── detailed.go           # Detailed spec: mobo, BIOS, PCI, NIC, sensors, software
│   │   └── reporter.go           # HTTP client — POST spec to server
│   ├── server/
│   │   ├── db.go                 # SQLite schema + CRUD
│   │   ├── handlers.go           # REST API handlers
│   │   └── templates/            # HTML templates
│   │       ├── base.html         # Layout
│   │       ├── index.html        # Machine list
│   │       └── machine.html      # Single machine detail
│   └── platform/
│       ├── wmi.go                # WMI stub (non-Windows)
│       └── wmi_windows.go        # WMI implementation (Windows)
```

## Build Commands

```bash
# Build Linux server
make server

# Build Windows agent GUI (requires Wails CLI)
make agent

# Install Wails CLI first
make install-wails

# Build both
make all

# Run server
make run-server
```

## REST API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/report` | Agent submits spec (JSON body) |
| GET | `/` | Dashboard — list all machines |
| GET | `/machines/{id}` | Dashboard — single machine detail |
| GET | `/api/agent/download` | Download agent .exe |
| GET | `/api/machines` | JSON list of all machines |
| GET | `/api/machines/{id}` | JSON machine detail |

## CLI Reference

```bash
# Start server
./spec-collector-server server -port 8080 -db ./data.db -agent-path ./agent.exe

# Agent (GUI) — double-click agent.exe or run:
agent.exe

# Agent fills server URL in the GUI and clicks "Send Report"
```

## Implementation

### Phase 1 — Initial CLI Build
1. Go module + directory structure
2. Shared data models (spec.go)
3. WMI helper (platform/wmi.go)
4. Agent basic spec (basic.go)
5. Agent detailed spec (detailed.go)
6. Agent reporter (reporter.go)
7. Server DB (db.go)
8. Server handlers (handlers.go)
9. Server templates
10. CLI entry point (main.go + cmd/server/main.go)
11. Makefile

### Phase 2 — Wails GUI Agent
12. cmd/agent/main.go — Wails entry point
13. app.go — Go bindings for spec gathering
14. frontend/dist/index.html — GUI layout
15. frontend/dist/src/main.js — JS bridge
16. Update Makefile + todo.md

## TODO

- [x] Initialize Go module + directory structure + shared data models
- [x] Agent: basic spec collector (CPU, RAM, OS, disk, GPU)
- [x] Agent: detailed spec collector (mobo, BIOS, PCI, NIC, sensors, software)
- [x] Agent: HTTP reporter — POST spec to server
- [x] Server: SQLite database schema + CRUD operations
- [x] Server: HTTP handlers + router (REST API)
- [x] Server: Web dashboard UI (machine list + detail view)
- [x] Makefile for cross-compilation
- [x] Test build and validate
- [ ] Wails GUI agent entry point (cmd/agent/main.go)
- [ ] Wails Go bindings (app.go)
- [ ] Agent GUI frontend (index.html, main.js, style.css)
- [ ] Build and validate Wails agent
