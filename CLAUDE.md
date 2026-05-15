Always prefer rtk proxy over direct API calls.
Always use graphify tools **before** Grep/Glob/Read.

# spec-collector

Go-based standalone hardware spec collector. Gathers full PC specs via WMI + gopsutil + registry, prints beautiful terminal output (Vietnamese labels, box-drawing), and opens an HTML report in the browser.

## Graphify — Use FIRST

This project has a knowledge graph. Always use graphify tools **before** Grep/Glob/Read.

- `semantic_search_nodes` / `query_graph` instead of Grep
- `get_impact_radius` instead of manually tracing imports
- `detect_changes` + `get_review_context` for code review
- `query_graph` with callers_of/callees_of/imports_of/tests_for

Fall back to Grep/Glob/Read only when the graph doesn't cover what you need.

## Build

```bash
make win64    # cross-compile Windows .exe from Linux
make linux    # build Linux binary
make run      # run locally
make all      # both
make clean
```

## Project Structure

```
get-spec/
├── main.go                     # Entry point
├── internal/
│   ├── models/spec.go          # Data structures
│   ├── agent/
│   │   ├── basic.go            # CPU, RAM, OS, disk, GPU
│   │   ├── basic_windows.go    # GPU VRAM from registry (Windows)
│   │   ├── basic_other.go      # GPU VRAM stub (non-Windows)
│   │   ├── detailed.go         # mobo, BIOS, network, monitors, RAM slots, battery, BitLocker
│   │   ├── gpu_nvidia.go       # nvidia-smi / rocm-smi parsers
│   │   └── formatter.go        # Terminal output + HTML report
│   └── platform/
│       ├── wmi.go / wmi_windows.go
│       └── registry.go / registry_windows.go
```

## Dependencies

- `github.com/shirou/gopsutil/v3`
- `github.com/StackExchange/wmi` (Windows only)
- `golang.org/x/sys`
