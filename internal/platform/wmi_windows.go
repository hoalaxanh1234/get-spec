//go:build windows

package platform

import (
	"runtime"

	"github.com/StackExchange/wmi"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func QueryWMI(query string, dst interface{}) error {
	return wmi.Query(query, dst)
}

func QueryWMINamespace(query string, dst interface{}, namespace string) error {
	return wmi.QueryNamespace(query, dst, namespace)
}
