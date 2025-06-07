package wasmtime

import (
	"runtime"
)

type wasm_externtype_t struct{}

// ExternType means one of external types which classify imports and external values with their respective types.
type ExternType struct {
	_ptr   uintptr //*wasm_externtype_t
	_owner interface{}
}

// AsExternType is an interface for all types which can be ExternType.
type AsExternType interface {
	AsExternType() *ExternType
}

func mkExternType(ptr uintptr, owner interface{}) *ExternType {
	externtype := &ExternType{_ptr: ptr, _owner: owner}
	if owner == nil {
		runtime.SetFinalizer(externtype, func(externtype *ExternType) {
			externtype.Close()
		})
	}
	return externtype
}

func (ty *ExternType) ptr() uintptr {
	ret := ty._ptr
	if ret == uintptr(0) {
		panic("object has been closed already")
	}
	//maybeGC()
	return ret
}

func (ty *ExternType) owner() interface{} {
	if ty._owner != nil {
		return ty._owner
	}
	return ty
}

// Close will deallocate this type's state explicitly.
//
// For more information see the documentation for engine.Close()
func (ty *ExternType) Close() {
	if ty._ptr == uintptr(0) || ty._owner != nil {
		return
	}
	runtime.SetFinalizer(ty, nil)
	wasm_externtype_delete(ty._ptr)
	ty._ptr = uintptr(0)
}

// FuncType returns the underlying `FuncType` for this `ExternType` if it's a function
// type. Otherwise returns `nil`.
func (ty *ExternType) FuncType() *FuncType {
	ptr := wasm_externtype_as_functype(ty.ptr())
	if ptr == uintptr(0) {
		return nil
	}
	return mkFuncType(ptr, ty.owner())
}

// GlobalType returns the underlying `GlobalType` for this `ExternType` if it's a *global* type.
// Otherwise returns `nil`.
func (ty *ExternType) GlobalType() *GlobalType {
	ptr := wasm_externtype_as_globaltype(ty.ptr())
	if ptr == uintptr(0) {
		return nil
	}
	return mkGlobalType(ptr, ty.owner())
}

// // TableType returns the underlying `TableType` for this `ExternType` if it's a *table* type.
// // Otherwise returns `nil`.
// func (ty *ExternType) TableType() *TableType {
// 	ptr := wasm_externtype_as_tabletype(ty.ptr())
// 	if ptr == nil {
// 		return nil
// 	}
// 	return mkTableType(ptr, ty.owner())
// }

// MemoryType returns the underlying `MemoryType` for this `ExternType` if it's a *memory* type.
// Otherwise returns `nil`.
func (ty *ExternType) MemoryType() *MemoryType {
	ptr := wasm_externtype_as_memorytype(ty.ptr())
	if ptr == uintptr(0) {
		return nil
	}
	return mkMemoryType(ptr, ty.owner())
}

// AsExternType returns this type itself
func (ty *ExternType) AsExternType() *ExternType {
	return ty
}
