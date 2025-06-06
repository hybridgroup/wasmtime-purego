package wasmtime

import (
	"runtime"
	"unsafe"
)

type wasmtime_instance_t struct {
	/// Internal identifier of what store this belongs to, never zero.
	store_id uint64
	/// Private data for use in Wasmtime.
	__private uint32
}

// Instance is an instantiated module instance.
// Once a module has been instantiated as an Instance, any exported function can be invoked externally via its function address funcaddr in the store S and an appropriate list val∗ of argument values.
type Instance struct {
	val wasmtime_instance_t
}

// NewInstance instantiates a WebAssembly `module` with the `imports` provided.
//
// This function will attempt to create a new wasm instance given the provided
// imports. This can fail if the wrong number of imports are specified, the
// imports aren't of the right type, or for other resource-related issues.
//
// This will also run the `start` function of the instance, returning an error
// if it traps.
func NewInstance(store Storelike, module *Module, imports []AsExtern) (*Instance, error) {
	importsRaw := make([]wasmtime_extern_t, len(imports))
	for i, imp := range imports {
		importsRaw[i] = imp.AsExtern()
	}
	var val wasmtime_instance_t
	err := enterWasm(store, func(trap **wasm_trap_t) uintptr {
		var imports *wasmtime_extern_t
		if len(importsRaw) > 0 {
			imports = (*wasmtime_extern_t)(unsafe.SliceData(importsRaw))
		}
		return wasmtime_instance_new(
			uintptr(store.Context()),
			module.ptr(),
			imports,
			len(importsRaw),
			&val,
			trap,
		)
	})
	runtime.KeepAlive(store)
	runtime.KeepAlive(module)
	runtime.KeepAlive(imports)
	runtime.KeepAlive(importsRaw)
	if err != nil {
		return nil, err
	}
	return mkInstance(val), nil
}

func mkInstance(val wasmtime_instance_t) *Instance {
	return &Instance{val}
}
