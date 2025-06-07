package wasmtime

import (
	"runtime"
	"unsafe"
)

// ValKind enumeration of different kinds of value types
type ValKind uint8

const (
	// KindI32 is the types i32 classify 32 bit integers. Integers are not inherently signed or unsigned, their interpretation is determined by individual operations.
	KindI32 ValKind = 0
	// KindI64 is the types i64 classify 64 bit integers. Integers are not inherently signed or unsigned, their interpretation is determined by individual operations.
	KindI64 ValKind = 1
	// KindF32 is the types f32 classify 32 bit floating-point data. They correspond to the respective binary floating-point representations, also known as single and double precision, as defined by the IEEE 754-2019 standard.
	KindF32 ValKind = 2
	// KindF64 is the types f64 classify 64 bit floating-point data. They correspond to the respective binary floating-point representations, also known as single and double precision, as defined by the IEEE 754-2019 standard.
	KindF64 ValKind = 3
	// TODO: Unknown
	KindExternref ValKind = 128
	// KindFuncref is the infinite union of all function types.
	KindFuncref ValKind = 129
)

// String renders this kind as a string, similar to the `*.wat` format
func (ty ValKind) String() string {
	switch ty {
	case KindI32:
		return "i32"
	case KindI64:
		return "i64"
	case KindF32:
		return "f32"
	case KindF64:
		return "f64"
	case KindExternref:
		return "externref"
	case KindFuncref:
		return "funcref"
	}
	panic("unknown kind")
}

type wasm_valtype_vec_t struct {
	size uint32          // C.size_t
	data *unsafe.Pointer // *C.wasm_valtype_t
}

type wasm_valtype_t struct {
	_kind uint8   // C.wasm_valkind_t
	_     [3]byte // C.wasm_valtype_reserved_t
}

// ValType means one of the value types, which classify the individual values that WebAssembly code can compute with and the values that a variable accepts.
type ValType struct {
	_ptr   uintptr // *wasm_valtype_t
	_owner interface{}
}

// NewValType creates a new `ValType` with the `kind` provided
func NewValType(kind ValKind) *ValType {
	ptr := wasm_valtype_new(uint8(kind))
	return mkValType(ptr, nil)
}

func mkValType(ptr uintptr, owner interface{}) *ValType {
	valtype := &ValType{_ptr: ptr, _owner: owner}
	if owner == nil {
		runtime.SetFinalizer(valtype, func(valtype *ValType) {
			valtype.Close()
		})
	}
	return valtype
}

// Kind returns the corresponding `ValKind` for this `ValType`
func (t *ValType) Kind() ValKind {
	ret := ValKind(wasm_valtype_kind(t.ptr()))
	runtime.KeepAlive(t)
	return ret
}

// Converts this `ValType` into a string according to the string representation
// of `ValKind`.
func (t *ValType) String() string {
	return t.Kind().String()
}

func (t *ValType) ptr() uintptr {
	ret := t._ptr
	if ret == uintptr(0) {
		panic("object has been closed already")
	}
	//maybeGC()
	return ret
}

// Close will deallocate this type's state explicitly.
//
// For more information see the documentation for engine.Close()
func (ty *ValType) Close() {
	if ty._ptr == uintptr(0) || ty._owner != nil {
		return
	}
	runtime.SetFinalizer(ty, nil)
	wasm_valtype_delete(ty._ptr)
	ty._ptr = uintptr(0)
}
