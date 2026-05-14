//go:build !windows

package agent

func getGPUVRAMFromRegistry() map[string]uint64 {
	return nil
}
