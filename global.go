package wasmtime

import (
	"runtime"
	"unsafe"
)

type wasmtime_global_t struct {
	/// Internal identifier of what store this belongs to, never zero.
	store_id uint64
	/// Private field for Wasmtime.
	__private1 uint32
	/// Private field for Wasmtime.
	__private2 uint32
	/// Private field for Wasmtime.
	__private3 uint32
}

// Global is a global instance, which is the runtime representation of a global variable.
// It holds an individual value and a flag indicating whether it is mutable.
// Read more in [spec](https://webassembly.github.io/spec/core/exec/runtime.html#global-instances)
type Global struct {
	val uintptr // C.wasmtime_global_t
}

// NewGlobal creates a new `Global` in the given `Store` with the specified `ty` and
// initial value `val`.
func NewGlobal(
	store Storelike,
	ty *GlobalType,
	val Val,
) (*Global, error) {
	var ret wasmtime_global_t
	var raw_val wasmtime_val_t
	val.initialize(store, &raw_val)
	err := wasmtime_global_new(
		uintptr(store.Context()),
		ty.ptr(),
		&raw_val,
		&ret,
	)
	wasmtime_val_unroot(uintptr(store.Context()), uintptr(unsafe.Pointer(&raw_val)))
	runtime.KeepAlive(store)
	runtime.KeepAlive(ty)
	if err != uintptr(0) {
		return nil, mkError(err)
	}

	return mkGlobal(uintptr(unsafe.Pointer(&ret))), nil
}

func mkGlobal(val uintptr) *Global {
	return &Global{val}
}

// Type returns the type of this global
func (g *Global) Type(store Storelike) *GlobalType {
	ptr := wasmtime_global_type(uintptr(store.Context()), g.val)
	runtime.KeepAlive(store)
	return mkGlobalType(ptr, nil)
}

// Get gets the value of this global
func (g *Global) Get(store Storelike) Val {
	ret := wasmtime_val_t{}
	wasmtime_global_get(uintptr(store.Context()), g.val, &ret)
	runtime.KeepAlive(store)
	return takeVal(store, &ret)
}

// Set sets the value of this global
func (g *Global) Set(store Storelike, val Val) error {
	var raw_val wasmtime_val_t
	val.initialize(store, &raw_val)
	err := wasmtime_global_set(uintptr(store.Context()), g.val, &raw_val)
	wasmtime_val_unroot(uintptr(store.Context()), uintptr(unsafe.Pointer(&raw_val)))
	runtime.KeepAlive(store)
	if err == uintptr(0) {
		return nil
	}

	return mkError(err)
}

func (g *Global) AsExtern() wasmtime_extern_t {
	ret := wasmtime_extern_t{kind: WASMTIME_EXTERN_GLOBAL}
	go_wasmtime_extern_global_set(&ret, g.val)
	return ret
}
