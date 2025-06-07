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
var wasmtime_store_context func(ptr uintptr) uintptr // returns *wasmtime_context_t
var wasmtime_module_new func(ptr uintptr, data []byte, size int, rtn *uintptr) uintptr
var wasmtime_module_delete func(ptr uintptr)
var wasmtime_error_delete func(ptr uintptr)
var wasmtime_wat2wasm func(wat string, size int, retVec *wasm_byte_vec_t) uintptr
var wasm_byte_vec_delete func(vec *wasm_byte_vec_t)
var wasm_valtype_new func(kind uint8) uintptr // *wasm_valtype_t
var wasm_valtype_kind func(ptr uintptr) uint8
var wasm_valtype_delete func(ptr uintptr)
var wasm_functype_new func(params, results *wasm_valtype_vec_t) uintptr
var wasm_valtype_vec_new_uninitialized func(vec *wasm_valtype_vec_t, size int) uintptr
var wasm_functype_delete func(ptr uintptr)            // *wasm_functype_t
var wasmtime_externref_data func(ptr uintptr) uintptr // returns *interface{} (externref data)
var wasmtime_func_new func(store uintptr, ty uintptr, callback uintptr, env int, wrap int, ret *wasmtime_func_t)
var wasmtime_caller_context func(caller uintptr) uintptr
var wasmtime_trap_new func(message string, size int) uintptr
var wasm_trap_delete func(ptr uintptr)
var wasmtime_val_unroot func(store uintptr, val uintptr) uintptr // returns *wasmtime_val_t
var wasm_functype_params func(ptr uintptr) uintptr
var wasm_functype_results func(ptr uintptr) uintptr
var wasm_externtype_delete func(ptr uintptr)                    // ExternType
var wasm_functype_as_externtype_const func(ptr uintptr) uintptr //*wasm_externtype_t // ExternType
var wasmtime_context_get_data func(ptr uintptr) uintptr         // returns *interface{} (context data)
var wasm_externtype_as_functype func(ptr uintptr) uintptr       // ExternType
var wasmtime_extern_delete func(ptr uintptr)                    // ExternType
var wasmtime_extern_type func(ctx uintptr, ptr uintptr) uintptr // returns *wasm_externtype_t
var wasmtime_instance_new func(
	store uintptr,
	module uintptr,
	imports *wasmtime_extern_t,
	len int,
	instance *wasmtime_instance_t,
	trap **wasm_trap_t,
) uintptr                                                                           // returns *wasmtime_error_t
var wasm_globaltype_new func(content uintptr, mutability wasm_mutability_t) uintptr // returns *wasm_globaltype_t
var wasm_globaltype_content func(ptr uintptr) uintptr                               // returns *wasm_valtype_t
var wasm_globaltype_mutability func(ptr uintptr) wasm_mutability_t                  // returns wasm_mutability_t
var wasm_globaltype_delete func(ptr uintptr)                                        // *wasm_globaltype_t
var wasm_externtype_as_globaltype func(ptr uintptr) uintptr                         // returns *wasm_globaltype_t
var wasm_externtype_as_functype_const func(ptr uintptr) uintptr                     // returns *wasm_functype_t
var wasm_globaltype_as_externtype_const func(ptr uintptr) uintptr                   // returns *wasm_externtype_t
var wasmtime_memorytype_new func(minimum uint64, hasMax bool, max uint64, is64bit bool, shared bool) uintptr
var wasmtime_memorytype_minimum func(ptr uintptr) uint32
var wasmtime_memorytype_maximum func(ptr uintptr, size *uint64) bool // returns bool, size uint32
var wasmtime_memorytype_is64 func(ptr uintptr) bool
var wasmtime_memorytype_isshared func(ptr uintptr) bool
var wasm_memorytype_as_externtype_const func(ptr uintptr) uintptr       // returns *wasm_externtype_t
var wasm_memorytype_delete func(ptr uintptr)                            // *wasm_memorytype_t
var wasm_externtype_as_memorytype func(ptr uintptr) uintptr             // returns *wasm_memorytype_t
var wasm_tabletype_new func(ptr uintptr, limits *wasm_limits_t) uintptr // returns *wasm_tabletype_t
var wasm_tabletype_element func(ptr uintptr) uintptr                    // *wasm_valtype_t
var wasm_tabletype_limits func(ptr uintptr) uintptr                     // returns *wasm_limits_t
var wasm_tabletype_delete func(ptr uintptr)                             // *wasm_tabletype_t
var wasm_tabletype_as_externtype_const func(ptr uintptr) uintptr        // returns *wasm_externtype_t
var wasm_externtype_as_tabletype func(ptr uintptr) uintptr              // returns *wasm_tabletype_t

var libshimsptr uintptr
var go_wasmtime_val_i32_set func(ptr *wasmtime_val_t, val int32)
var go_wasmtime_val_i64_set func(ptr *wasmtime_val_t, val int64)
var go_wasmtime_val_f32_set func(ptr *wasmtime_val_t, val float32)
var go_wasmtime_val_f64_set func(ptr *wasmtime_val_t, val float64)
var go_wasmtime_val_funcref_set func(ptr *wasmtime_val_t, val uintptr)   //val *Func)
var go_wasmtime_val_externref_set func(ptr *wasmtime_val_t, val uintptr) //val interface{})
var go_wasmtime_val_i32_get func(ptr *wasmtime_val_t) int32
var go_wasmtime_val_i64_get func(ptr *wasmtime_val_t) int64
var go_wasmtime_val_f32_get func(ptr *wasmtime_val_t) float32
var go_wasmtime_val_f64_get func(ptr *wasmtime_val_t) float64
var go_wasmtime_val_funcref_get func(ptr *wasmtime_val_t) *wasmtime_func_t        // *Func
var go_wasmtime_val_externref_get func(ptr *wasmtime_val_t) uintptr               // interface{}
var go_wasmtime_extern_func_get func(ptr *wasmtime_extern_t) uintptr              // returns *wasmtime_func_t
var go_wasmtime_extern_func_set func(ptr *wasmtime_extern_t, val uintptr) uintptr // returns *wasmtime_error_t

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
	purego.RegisterLibFunc(&wasmtime_store_context, libptr, "wasmtime_store_context")
	purego.RegisterLibFunc(&wasmtime_module_new, libptr, "wasmtime_module_new")
	purego.RegisterLibFunc(&wasmtime_module_delete, libptr, "wasmtime_module_delete")
	purego.RegisterLibFunc(&wasmtime_error_delete, libptr, "wasmtime_error_delete")
	purego.RegisterLibFunc(&wasmtime_wat2wasm, libptr, "wasmtime_wat2wasm")
	purego.RegisterLibFunc(&wasm_byte_vec_delete, libptr, "wasm_byte_vec_delete")
	purego.RegisterLibFunc(&wasm_valtype_new, libptr, "wasm_valtype_new")
	purego.RegisterLibFunc(&wasm_valtype_kind, libptr, "wasm_valtype_kind")
	purego.RegisterLibFunc(&wasm_valtype_delete, libptr, "wasm_valtype_delete")
	purego.RegisterLibFunc(&wasm_functype_new, libptr, "wasm_functype_new")
	purego.RegisterLibFunc(&wasm_valtype_vec_new_uninitialized, libptr, "wasm_valtype_vec_new_uninitialized")
	purego.RegisterLibFunc(&wasm_functype_delete, libptr, "wasm_functype_delete")
	purego.RegisterLibFunc(&wasmtime_externref_data, libptr, "wasmtime_externref_data")
	purego.RegisterLibFunc(&wasmtime_func_new, libptr, "wasmtime_func_new")
	purego.RegisterLibFunc(&wasmtime_caller_context, libptr, "wasmtime_caller_context")
	purego.RegisterLibFunc(&wasmtime_trap_new, libptr, "wasmtime_trap_new")
	purego.RegisterLibFunc(&wasm_trap_delete, libptr, "wasm_trap_delete")
	purego.RegisterLibFunc(&wasmtime_val_unroot, libptr, "wasmtime_val_unroot")
	purego.RegisterLibFunc(&wasm_functype_params, libptr, "wasm_functype_params")
	purego.RegisterLibFunc(&wasm_functype_results, libptr, "wasm_functype_results")
	purego.RegisterLibFunc(&wasm_externtype_delete, libptr, "wasm_externtype_delete")
	purego.RegisterLibFunc(&wasm_functype_as_externtype_const, libptr, "wasm_functype_as_externtype_const")
	purego.RegisterLibFunc(&wasmtime_context_get_data, libptr, "wasmtime_context_get_data")
	purego.RegisterLibFunc(&wasm_externtype_as_functype, libptr, "wasm_externtype_as_functype")
	purego.RegisterLibFunc(&wasmtime_extern_delete, libptr, "wasmtime_extern_delete")
	purego.RegisterLibFunc(&wasmtime_extern_type, libptr, "wasmtime_extern_type")
	purego.RegisterLibFunc(&wasmtime_instance_new, libptr, "wasmtime_instance_new")
	purego.RegisterLibFunc(&wasm_globaltype_new, libptr, "wasm_globaltype_new")
	purego.RegisterLibFunc(&wasm_globaltype_content, libptr, "wasm_globaltype_content")
	purego.RegisterLibFunc(&wasm_globaltype_mutability, libptr, "wasm_globaltype_mutability")
	purego.RegisterLibFunc(&wasm_globaltype_delete, libptr, "wasm_globaltype_delete")
	purego.RegisterLibFunc(&wasm_externtype_as_globaltype, libptr, "wasm_externtype_as_globaltype")
	purego.RegisterLibFunc(&wasm_externtype_as_functype_const, libptr, "wasm_externtype_as_functype_const")
	purego.RegisterLibFunc(&wasm_globaltype_as_externtype_const, libptr, "wasm_globaltype_as_externtype_const")
	purego.RegisterLibFunc(&wasmtime_memorytype_new, libptr, "wasmtime_memorytype_new")
	purego.RegisterLibFunc(&wasmtime_memorytype_minimum, libptr, "wasmtime_memorytype_minimum")
	purego.RegisterLibFunc(&wasmtime_memorytype_maximum, libptr, "wasmtime_memorytype_maximum")
	purego.RegisterLibFunc(&wasmtime_memorytype_is64, libptr, "wasmtime_memorytype_is64")
	purego.RegisterLibFunc(&wasmtime_memorytype_isshared, libptr, "wasmtime_memorytype_isshared")
	purego.RegisterLibFunc(&wasm_memorytype_as_externtype_const, libptr, "wasm_memorytype_as_externtype_const")
	purego.RegisterLibFunc(&wasm_memorytype_delete, libptr, "wasm_memorytype_delete")
	purego.RegisterLibFunc(&wasm_externtype_as_memorytype, libptr, "wasm_externtype_as_memorytype")
	purego.RegisterLibFunc(&wasm_tabletype_new, libptr, "wasm_tabletype_new")
	purego.RegisterLibFunc(&wasm_tabletype_element, libptr, "wasm_tabletype_element")
	purego.RegisterLibFunc(&wasm_tabletype_limits, libptr, "wasm_tabletype_limits")
	purego.RegisterLibFunc(&wasm_tabletype_delete, libptr, "wasm_tabletype_delete")
	purego.RegisterLibFunc(&wasm_tabletype_as_externtype_const, libptr, "wasm_tabletype_as_externtype_const")
	purego.RegisterLibFunc(&wasmtime_context_get_data, libptr, "wasmtime_context_get_data")
	purego.RegisterLibFunc(&wasm_externtype_as_tabletype, libptr, "wasm_externtype_as_tabletype")

	libshims, err := findWasmtimeShims()
	if err != nil {
		panic(err)
	}
	if libshimsptr, err = load(libshims); err != nil {
		panic(err)
	}

	purego.RegisterLibFunc(&go_wasmtime_val_i32_set, libshimsptr, "go_wasmtime_val_i32_set")
	purego.RegisterLibFunc(&go_wasmtime_val_i64_set, libshimsptr, "go_wasmtime_val_i64_set")
	purego.RegisterLibFunc(&go_wasmtime_val_f32_set, libshimsptr, "go_wasmtime_val_f32_set")
	purego.RegisterLibFunc(&go_wasmtime_val_f64_set, libshimsptr, "go_wasmtime_val_f64_set")
	purego.RegisterLibFunc(&go_wasmtime_val_funcref_set, libshimsptr, "go_wasmtime_val_funcref_set")
	purego.RegisterLibFunc(&go_wasmtime_val_externref_set, libshimsptr, "go_wasmtime_val_externref_set")
	purego.RegisterLibFunc(&go_wasmtime_val_i32_get, libshimsptr, "go_wasmtime_val_i32_get")
	purego.RegisterLibFunc(&go_wasmtime_val_i64_get, libshimsptr, "go_wasmtime_val_i64_get")
	purego.RegisterLibFunc(&go_wasmtime_val_f32_get, libshimsptr, "go_wasmtime_val_f32_get")
	purego.RegisterLibFunc(&go_wasmtime_val_f64_get, libshimsptr, "go_wasmtime_val_f64_get")
	purego.RegisterLibFunc(&go_wasmtime_val_funcref_get, libshimsptr, "go_wasmtime_val_funcref_get")
	purego.RegisterLibFunc(&go_wasmtime_val_externref_get, libshimsptr, "go_wasmtime_val_externref_get")
	purego.RegisterLibFunc(&go_wasmtime_extern_func_get, libshimsptr, "go_wasmtime_extern_func_get")
	purego.RegisterLibFunc(&go_wasmtime_extern_func_set, libshimsptr, "go_wasmtime_extern_func_set")
}

// findWasmtime searches for the dynamic library in standard system paths.
func findWasmtime() (string, error) {
	switch runtime.GOOS {
	case "windows":
		// TODO: also handle libwasmtime.dll.a?
		return findLibrary("wasmtime.dll", runtime.GOOS)
	case "darwin":
		return findLibrary("libwasmtime.dylib", runtime.GOOS)
	default:
		return findLibrary("libwasmtime.so", runtime.GOOS)
	}
}

// findWasmtimeShims searches for the dynamic library in standard system paths.
func findWasmtimeShims() (string, error) {
	switch runtime.GOOS {
	case "windows":
		// TODO: also handle libwasmtime.dll.a?
		return findLibrary("wasmtime-shims.dll", runtime.GOOS)
	case "darwin":
		return findLibrary("libwasmtime-shims.dylib", runtime.GOOS)
	default:
		return findLibrary("libwasmtime-shims.so", runtime.GOOS)
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
