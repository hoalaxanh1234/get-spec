//go:build windows

package agent

import (
	"strings"

	"golang.org/x/sys/windows/registry"
)

func getGPUVRAMFromRegistry() map[string]uint64 {
	baseKey := `SYSTEM\CurrentControlSet\Control\Class\{4d36e968-e325-11ce-bfc1-08002be10318}`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, baseKey, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil
	}
	defer k.Close()

	subKeys, _ := k.ReadSubKeyNames(-1)
	result := make(map[string]uint64)
	for _, sk := range subKeys {
		if len(sk) != 4 {
			continue
		}
		adapterKey := baseKey + `\` + sk
		ak, err := registry.OpenKey(registry.LOCAL_MACHINE, adapterKey, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		name, _, err := ak.GetStringValue("DriverDesc")
		if err != nil || name == "" {
			ak.Close()
			continue
		}
		vram, _, err := ak.GetIntegerValue("HardwareInformation.qwMemorySize")
		if err != nil {
			vram, _, err = ak.GetIntegerValue("HardwareInformation.AdapterRAM")
			if err != nil {
				ak.Close()
				continue
			}
		}
		ak.Close()
		result[strings.ToLower(strings.TrimSpace(name))] = vram
	}
	return result
}
