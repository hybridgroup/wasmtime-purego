package wasmtime

import (
	"runtime"
	"unsafe"
)

// ExportType is one of the exports component.
// A module defines a set of exports that become accessible to the host environment once the module has been instantiated.
type ExportType struct {
	_ptr   uintptr // *C.wasm_exporttype_t
	_owner interface{}
}

// NewExportType creates a new `ExportType` with the `name` and the type provided.
func NewExportType(name string, ty AsExternType) *ExportType {
	nameVec := stringToByteVec(name)

	// Creating an export type requires taking ownership, so create a copy
	// so we don't have to invalidate pointers here. Shouldn't be too
	// costly in theory anyway.
	extern := ty.AsExternType()
	ptr := wasm_externtype_copy(extern.ptr())
	runtime.KeepAlive(extern)

	// And once we've got all that create the export type!
	exportPtr := wasm_exporttype_new(&nameVec, ptr)

	return mkExportType(exportPtr, nil)
}

func mkExportType(ptr uintptr, owner interface{}) *ExportType {
	exporttype := &ExportType{_ptr: ptr, _owner: owner}
	if owner == nil {
		runtime.SetFinalizer(exporttype, func(exporttype *ExportType) {
			exporttype.Close()
		})
	}
	return exporttype
}

func (ty *ExportType) ptr() uintptr {
	ret := ty._ptr
	if ret == uintptr(0) {
		panic("object has been closed already")
	}
	// maybeGC()
	return ret
}

func (ty *ExportType) owner() interface{} {
	if ty._owner != nil {
		return ty._owner
	}
	return ty
}

// Close will deallocate this type's state explicitly.
//
// For more information see the documentation for engine.Close()
func (ty *ExportType) Close() {
	if ty._ptr == uintptr(0) || ty._owner != nil {
		return
	}
	runtime.SetFinalizer(ty, nil)
	wasm_exporttype_delete(ty._ptr)
	ty._ptr = uintptr(0)
}

// Name returns the name in the module this export type is exporting
func (ty *ExportType) Name() string {
	ret := (*wasm_byte_vec_t)(unsafe.Pointer(wasm_exporttype_name(ty.ptr())))
	runtime.KeepAlive(ty)

	return unsafe.String(ret.data, ret.size)
}

// Type returns the type of item this export type expects
func (ty *ExportType) Type() *ExternType {
	ptr := wasm_exporttype_type(ty.ptr())
	return mkExternType(ptr, ty.owner())
}
