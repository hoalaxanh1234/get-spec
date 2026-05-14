package agent

import (
	"os/exec"
	"strconv"
	"strings"
)

func getGPUVRAMFromCLI() map[string]uint64 {
	result := getGPUVRAMFromNvidiaSMI()
	amd := getGPUVRAMFromAMD()
	for k, v := range amd {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}
	return result
}

func getGPUVRAMFromNvidiaSMI() map[string]uint64 {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader,nounits")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	return parseCommaCSV(string(out))
}

func getGPUVRAMFromAMD() map[string]uint64 {
	cmd := exec.Command("rocm-smi", "--showmeminfo", "vram", "--csv")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	return parseROCMCSV(string(out))
}

func parseCommaCSV(out string) map[string]uint64 {
	result := make(map[string]uint64)
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ",", 2)
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		memStr := strings.TrimSpace(parts[1])
		memMiB, err := strconv.ParseUint(memStr, 10, 64)
		if err != nil || memMiB == 0 {
			continue
		}
		result[strings.ToLower(name)] = memMiB * 1024 * 1024
	}
	return result
}

func parseROCMCSV(out string) map[string]uint64 {
	result := make(map[string]uint64)
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "GPU[") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) > 3 {
			name := strings.TrimSpace(parts[1])
			memStr := strings.TrimSpace(parts[3])
			memBytes, err := strconv.ParseUint(memStr, 10, 64)
			if err != nil || memBytes == 0 {
				continue
			}
			result[strings.ToLower(name)] = memBytes
		}
	}
	return result
}
