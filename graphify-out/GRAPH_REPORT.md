# Graph Report - .  (2026-05-15)

## Corpus Check
- Corpus is ~6,953 words - fits in a single context window. You may not need a graph.

## Summary
- 132 nodes · 183 edges · 16 communities (13 shown, 3 thin omitted)
- Extraction: 88% EXTRACTED · 11% INFERRED · 1% AMBIGUOUS · INFERRED: 21 edges (avg confidence: 0.85)
- Token cost: 15,200 input · 1,800 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Hardware Detail Gathering|Hardware Detail Gathering]]
- [[_COMMUNITY_Basic System Gathering|Basic System Gathering]]
- [[_COMMUNITY_Architecture & Design Concepts|Architecture & Design Concepts]]
- [[_COMMUNITY_Data Models & Structs|Data Models & Structs]]
- [[_COMMUNITY_Output & Interactivity|Output & Interactivity]]
- [[_COMMUNITY_Formatting Utilities|Formatting Utilities]]
- [[_COMMUNITY_Windows Registry|Windows Registry]]
- [[_COMMUNITY_GPU VRAM Parsers|GPU VRAM Parsers]]
- [[_COMMUNITY_README Documentation|README Documentation]]
- [[_COMMUNITY_CLAUDE.md Instructions|CLAUDE.md Instructions]]
- [[_COMMUNITY_Todo  Future Plans|Todo / Future Plans]]

## God Nodes (most connected - your core abstractions)
1. `Gather()` - 13 edges
2. `FormatSpec()` - 9 edges
3. `GatherDetailed()` - 8 edges
4. `Basic Collector Helpers` - 8 edges
5. `main()` - 6 edges
6. `interactive()` - 6 edges
7. `buildHTML()` - 5 edges
8. `getBattery()` - 5 edges
9. `WMI Windows Implementation` - 5 edges
10. `Registry Windows Implementation` - 5 edges

## Surprising Connections (you probably didn't know these)
- `Terminal Output Formatter` --semantically_similar_to--> `HTML Report Generator`  [INFERRED] [semantically similar]
  internal/agent/formatter.go → main.go
- `main()` --calls--> `Gather()`  [INFERRED]
  main.go → internal/agent/basic.go
- `main()` --calls--> `FormatSpec()`  [INFERRED]
  main.go → internal/agent/formatter.go
- `Vietnamese-Language UI` --rationale_for--> `HTML Report Generator`  [INFERRED]
  internal/agent/formatter.go → main.go
- `Gather()` --calls--> `GatherDetailed()`  [INFERRED]
  internal/agent/basic.go → internal/agent/detailed.go

## Hyperedges (group relationships)
- **Spec Collection Pipeline** — gather_orchestrator, basic_collector_helpers, detailed_collector, machine_data_models [INFERRED 0.90]
- **Platform Abstraction Layer** — wmi_stub, wmi_windows, registry_stub, registry_windows, platform_abstraction_pattern [INFERRED 0.95]
- **Output Formatters** — terminal_formatter, html_report_generator, machine_data_models [INFERRED 0.85]

## Communities (16 total, 3 thin omitted)

### Community 0 - "Hardware Detail Gathering"
Cohesion: 0.13
Nodes (23): acpiBatteryStatus(), batteryChemistryString(), batteryStatusString(), formatResolution(), GatherDetailed(), getBattery(), getBIOS(), getMonitors() (+15 more)

### Community 1 - "Basic System Gathering"
Cohesion: 0.14
Nodes (20): Gather(), generateMachineID(), getCPUTDP(), getDisks(), getDisksWMI(), getGPUs(), getIPAddress(), getMACAddress() (+12 more)

### Community 2 - "Architecture & Design Concepts"
Cohesion: 0.2
Nodes (18): Basic Collector Helpers, Detailed Spec Collector, Gather Orchestrator, GPU VRAM CLI Parser, GPU VRAM Registry (Agent), GPU VRAM Registry Stub, HTML Report Generator, Interactive Menu (+10 more)

### Community 3 - "Data Models & Structs"
Cohesion: 0.12
Nodes (16): BatteryInfo, BIOSInfo, BitLockerInfo, CPUInfo, DiskInfo, GPUInfo, Machine, MonitorInfo (+8 more)

### Community 4 - "Output & Interactivity"
Cohesion: 0.33
Nodes (11): addRow(), buildHTML(), copyToClipboard(), editInEditor(), generateHTML(), interactive(), isGeneric(), machineJSON() (+3 more)

### Community 5 - "Formatting Utilities"
Cohesion: 0.44
Nodes (8): batteryHealth(), formatClock(), formatOS(), FormatSpec(), formatVRAM(), isGeneric(), kvLine(), pickValue()

### Community 6 - "Windows Registry"
Cohesion: 0.47
Nodes (3): findGPUNameMapping(), ReadRegistryDWORD(), ReadRegistryQWORD()

### Community 7 - "GPU VRAM Parsers"
Cohesion: 0.6
Nodes (5): getGPUVRAMFromAMD(), getGPUVRAMFromCLI(), getGPUVRAMFromNvidiaSMI(), parseCommaCSV(), parseROCMCSV()

## Ambiguous Edges - Review These
- `Registry Windows Implementation` → `GPU VRAM Registry (Agent)`  [AMBIGUOUS]
  internal/platform/registry_windows.go · relation: conceptually_related_to

## Knowledge Gaps
- **38 isolated node(s):** `Machine`, `OSInfo`, `CPUInfo`, `RAMInfo`, `DiskInfo` (+33 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **3 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `Registry Windows Implementation` and `GPU VRAM Registry (Agent)`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **Why does `Gather()` connect `Basic System Gathering` to `Hardware Detail Gathering`, `Output & Interactivity`?**
  _High betweenness centrality (0.201) - this node is a cross-community bridge._
- **Why does `main()` connect `Output & Interactivity` to `Basic System Gathering`, `Formatting Utilities`?**
  _High betweenness centrality (0.132) - this node is a cross-community bridge._
- **Why does `GatherDetailed()` connect `Hardware Detail Gathering` to `Basic System Gathering`?**
  _High betweenness centrality (0.068) - this node is a cross-community bridge._
- **Are the 4 inferred relationships involving `Gather()` (e.g. with `main()` and `GatherDetailed()`) actually correct?**
  _`Gather()` has 4 INFERRED edges - model-reasoned connections that need verification._
- **Are the 2 inferred relationships involving `Basic Collector Helpers` (e.g. with `GPU VRAM Registry Stub` and `Two-Phase Gathering Design`) actually correct?**
  _`Basic Collector Helpers` has 2 INFERRED edges - model-reasoned connections that need verification._
- **Are the 2 inferred relationships involving `main()` (e.g. with `Gather()` and `FormatSpec()`) actually correct?**
  _`main()` has 2 INFERRED edges - model-reasoned connections that need verification._