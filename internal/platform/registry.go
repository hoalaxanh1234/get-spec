//go:build !windows

package platform

import "fmt"

var errNotSupported = fmt.Errorf("registry access is only supported on Windows")

func ReadRegistryString(keyPath, valueName string) (string, error) {
	return "", errNotSupported
}

func ReadRegistryDWORD(keyPath, valueName string) (uint64, error) {
	return 0, errNotSupported
}

func ReadRegistryQWORD(keyPath, valueName string) (uint64, error) {
	return 0, errNotSupported
}
