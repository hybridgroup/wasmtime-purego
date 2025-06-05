//go:build windows

package wasmtime

import "syscall"

func load(name string) (uintptr, error) {
	// Use [syscall.LoadLibrary] here to avoid external dependencies.
	// For actual use cases, [golang.org/x/sys/windows.NewLazySystemDLL] is recommended.
	handle, err := syscall.LoadLibrary(name)
	return uintptr(handle), err
}
