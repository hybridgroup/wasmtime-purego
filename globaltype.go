package wasmtime

import "runtime"

type wasm_mutability_t uint8

const (
	WASM_CONST wasm_mutability_t = 0 // Immutable global
	WASM_VAR   wasm_mutability_t = 1 // Mutable global
)

// GlobalType is a ValType, which classify global variables and hold a value and can either be mutable or immutable.
type GlobalType struct {
	_ptr   uintptr // *C.wasm_globaltype_t
	_owner interface{}
}

// NewGlobalType creates a new `GlobalType` with the `kind` provided and whether it's
// `mutable` or not
func NewGlobalType(content *ValType, mutable bool) *GlobalType {
	mutability := WASM_CONST
	if mutable {
		mutability = WASM_VAR
	}
	contentPtr := wasm_valtype_new(wasm_valtype_kind(content.ptr()))
	runtime.KeepAlive(content)
	ptr := wasm_globaltype_new(contentPtr, wasm_mutability_t(mutability))

	return mkGlobalType(ptr, nil)
}

func mkGlobalType(ptr uintptr, owner interface{}) *GlobalType {
	globaltype := &GlobalType{_ptr: ptr, _owner: owner}
	if owner == nil {
		runtime.SetFinalizer(globaltype, func(globaltype *GlobalType) {
			globaltype.Close()
		})
	}
	return globaltype
}

func (ty *GlobalType) ptr() uintptr {
	ret := ty._ptr
	if ret == uintptr(0) {
		panic("object has been closed already")
	}
	//maybeGC()
	return ret
}

func (ty *GlobalType) owner() interface{} {
	if ty._owner != nil {
		return ty._owner
	}
	return ty
}

// Close will deallocate this type's state explicitly.
//
// For more information see the documentation for engine.Close()
func (ty *GlobalType) Close() {
	if ty._ptr == uintptr(0) || ty._owner != nil {
		return
	}
	runtime.SetFinalizer(ty, nil)
	wasm_globaltype_delete(ty._ptr)
	ty._ptr = uintptr(0)
}

// Content returns the type of value stored in this global
func (ty *GlobalType) Content() *ValType {
	ptr := wasm_globaltype_content(ty.ptr())
	return mkValType(ptr, ty.owner())
}

// Mutable returns whether this global type is mutable or not
func (ty *GlobalType) Mutable() bool {
	ret := wasm_globaltype_mutability(ty.ptr()) == WASM_VAR
	runtime.KeepAlive(ty)
	return ret
}

// AsExternType converts this type to an instance of `ExternType`
func (ty *GlobalType) AsExternType() *ExternType {
	ptr := wasm_globaltype_as_externtype_const(ty.ptr())
	return mkExternType(ptr, ty.owner())
}
