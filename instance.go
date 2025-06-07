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
	val uintptr // wasmtime_instance_t
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
	return mkInstance(uintptr(unsafe.Pointer(&val))), nil
}

func mkInstance(val uintptr) *Instance {
	return &Instance{val}
}

// Exports returns a list of exports from this instance.
//
// Each export is returned as a `*Extern` and lines up with the exports list of
// the associated `Module`.
func (instance *Instance) Exports(store Storelike) []*Extern {
	ret := make([]*Extern, 0)
	var name string
	var name_len int
	for i := 0; ; i++ {
		var item wasmtime_extern_t
		ok := wasmtime_instance_export_nth(
			store.Context(),
			instance.val,
			i,
			&name,
			&name_len,
			&item,
		)
		if !ok {
			break
		}
		ret = append(ret, mkExtern(uintptr(unsafe.Pointer(&item))))
	}
	runtime.KeepAlive(store)
	return ret
}

// GetExport attempts to find an export on this instance by `name`
//
// May return `nil` if this instance has no export named `name`
func (i *Instance) GetExport(store Storelike, name string) *Extern {
	var item wasmtime_extern_t
	ok := wasmtime_instance_export_get(
		uintptr(store.Context()),
		uintptr(unsafe.Pointer(&i.val)),
		name,
		len(name),
		&item,
	)
	runtime.KeepAlive(store)
	runtime.KeepAlive(name)
	if ok {
		return mkExtern(uintptr(unsafe.Pointer(&item)))
	}
	return nil
}

// GetFunc attempts to find a function on this instance by `name`.
//
// May return `nil` if this instance has no function named `name`,
// it is not a function, etc.
func (i *Instance) GetFunc(store Storelike, name string) *Func {
	f := i.GetExport(store, name)
	if f == nil {
		return nil
	}
	return f.Func()
}
