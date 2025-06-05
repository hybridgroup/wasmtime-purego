package wasmtime

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ebitengine/purego"
)

// libptr is a pointer to the loaded dynamic library.
var libptr uintptr

var wasm_engine_new func() uintptr
var wasm_engine_delete func(ptr uintptr)
var wasmtime_store_new func(ptr uintptr, idx int) uintptr
var wasmtime_store_delete func(ptr uintptr)
var wasmtime_module_new func(ptr uintptr, data []byte, size int, rtn *uintptr) uintptr
var wasmtime_module_delete func(ptr uintptr)
var wasmtime_error_delete func(ptr uintptr)
var wasmtime_wat2wasm func(wat string, size int, retVec *wasm_byte_vec_t) uintptr
var wasm_byte_vec_delete func(vec *wasm_byte_vec_t)

func init() {
	libpath, err := findWasmtime()
	if err != nil {
		panic(err)
	}
	if libptr, err = load(libpath); err != nil {
		panic(err)
	}

	// Load the library functions
	purego.RegisterLibFunc(&wasm_engine_new, libptr, "wasm_engine_new")
	purego.RegisterLibFunc(&wasm_engine_delete, libptr, "wasm_engine_delete")
	purego.RegisterLibFunc(&wasmtime_store_new, libptr, "wasmtime_store_new")
	purego.RegisterLibFunc(&wasmtime_store_delete, libptr, "wasmtime_store_delete")
	purego.RegisterLibFunc(&wasmtime_module_new, libptr, "wasmtime_module_new")
	purego.RegisterLibFunc(&wasmtime_module_delete, libptr, "wasmtime_module_delete")
	purego.RegisterLibFunc(&wasmtime_error_delete, libptr, "wasmtime_error_delete")
	purego.RegisterLibFunc(&wasmtime_wat2wasm, libptr, "wasmtime_wat2wasm")
	purego.RegisterLibFunc(&wasm_byte_vec_delete, libptr, "wasm_byte_vec_delete")
}

// findWasmtime searches for the dynamic library in standard system paths.
func findWasmtime() (string, error) {
	switch runtime.GOOS {
	case "windows":
		// TODO: also handle libwasmtime.dll.a
		return findLibrary("wasmtime.dll", runtime.GOOS)
	case "darwin":
		return findLibrary("libwasmtime.dylib", runtime.GOOS)
	default:
		return findLibrary("libwasmtime.so", runtime.GOOS)
	}
}

// findLibrary searches for a dynamic library by name across standard system paths.
// It returns the full path to the library if found, or an error listing all searched paths.
func findLibrary(libName, goos string, dirs ...string) (string, error) {
	libExt, commonPaths := findLibDirs(goos)
	dirs = append(dirs, commonPaths...)

	// Append the correct extension if missing
	if !strings.HasSuffix(libName, libExt) {
		libName += libExt
	}

	// Include current working directory
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
	}

	// Iterate through directories and search for the library
	searched := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		filename := filepath.Join(dir, libName)
		searched = append(searched, filename)
		if fi, err := os.Stat(filename); err == nil && !fi.IsDir() {
			return filename, nil // Library found
		}
	}

	// Construct error message listing all searched paths
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Library '%s' not found, checked following paths:\n", libName))
	for _, path := range searched {
		sb.WriteString(fmt.Sprintf(" - %s\n", path))
	}

	return "", errors.New(sb.String())
}

// findLibDirs returns the library extension, relevant environment path, and common library directories based on the OS.
func findLibDirs(goos string) (string, []string) {
	switch goos {
	case "windows":
		systemRoot := os.Getenv("SystemRoot")
		return ".dll", append(
			filepath.SplitList(os.Getenv("PATH")),
			filepath.Join(systemRoot, "System32"),
			filepath.Join(systemRoot, "SysWOW64"),
		)
	case "darwin":
		return ".dylib", append(
			filepath.SplitList(os.Getenv("DYLD_LIBRARY_PATH")),
			"/usr/lib",
			"/usr/local/lib",
		)
	default: // Unix/Linux
		return ".so", append(
			filepath.SplitList(os.Getenv("LD_LIBRARY_PATH")),
			"/lib",
			"/usr/lib",
			"/usr/local/lib",
		)
	}
}
