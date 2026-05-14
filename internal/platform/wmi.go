//go:build !windows

package platform

import (
	"fmt"
	"runtime"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func QueryWMI(query string, dst interface{}) error {
	return fmt.Errorf("WMI queries are only supported on Windows")
}

func QueryWMINamespace(query string, dst interface{}, namespace string) error {
	return fmt.Errorf("WMI queries are only supported on Windows")
}
