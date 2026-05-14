//go:build windows

package platform

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func ReadRegistryString(keyPath, valueName string) (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("open key %s: %w", keyPath, err)
	}
	defer k.Close()
	val, _, err := k.GetStringValue(valueName)
	if err != nil {
		return "", fmt.Errorf("read %s\\%s: %w", keyPath, valueName, err)
	}
	return val, nil
}

func ReadRegistryDWORD(keyPath, valueName string) (uint64, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE)
	if err != nil {
		return 0, fmt.Errorf("open key %s: %w", keyPath, err)
	}
	defer k.Close()
	val, _, err := k.GetIntegerValue(valueName)
	if err != nil {
		return 0, fmt.Errorf("read %s\\%s: %w", keyPath, valueName, err)
	}
	return val, nil
}

func ReadRegistryQWORD(keyPath, valueName string) (uint64, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE)
	if err != nil {
		return 0, fmt.Errorf("open key %s: %w", keyPath, err)
	}
	defer k.Close()
	val, _, err := k.GetIntegerValue(valueName)
	if err != nil {
		return 0, fmt.Errorf("read %s\\%s: %w", keyPath, valueName, err)
	}
	return val, nil
}

func FindGPUVRAM() ([]uint64, error) {
	baseKey := `SYSTEM\CurrentControlSet\Control\Class\{4d36e968-e325-11ce-bfc1-08002be10318}`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, baseKey, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, fmt.Errorf("open GPU class key: %w", err)
	}
	defer k.Close()

	subKeys, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return nil, fmt.Errorf("enum GPU subkeys: %w", err)
	}

	var results []uint64
	for _, sk := range subKeys {
		adapterKey := baseKey + `\` + sk
		vram, err := func() (uint64, error) {
			ak, err := registry.OpenKey(registry.LOCAL_MACHINE, adapterKey, registry.QUERY_VALUE)
			if err != nil {
				return 0, err
			}
			defer ak.Close()
			val, _, err := ak.GetIntegerValue("HardwareInformation.qwMemorySize")
			if err != nil {
				val, _, err = ak.GetIntegerValue("HardwareInformation.AdapterRAM")
				if err != nil {
					return 0, err
				}
			}
			return val, nil
		}()
		if err == nil && vram > 0 {
			results = append(results, vram)
		}
	}
	return results, nil
}

// findGPUNameMapping reads GPU adapter names from registry to match VRAM entries.
func findGPUNameMapping() map[string]uint64 {
	result := make(map[string]uint64)
	baseKey := `SYSTEM\CurrentControlSet\Control\Class\{4d36e968-e325-11ce-bfc1-08002be10318}`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, baseKey, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return result
	}
	defer k.Close()

	subKeys, _ := k.ReadSubKeyNames(-1)
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
		ak.Close()
		if err != nil || name == "" {
			continue
		}
		vram, err := ReadRegistryQWORD(adapterKey, "HardwareInformation.qwMemorySize")
		if err != nil {
			vram, err = ReadRegistryDWORD(adapterKey, "HardwareInformation.AdapterRAM")
			if err != nil {
				continue
			}
		}
		key := strings.ToLower(strings.TrimSpace(name))
		result[key] = vram
	}
	return result
}
