package wasmtime

import (
	"runtime"
	"unsafe"
)

type wasm_functype_t struct {
	ext unsafe.Pointer // *C.wasm_externtype_t
	// size_t params_count;
	// wasm_valtype_t* params_data;
	// size_t results_count;
	// wasm_valtype_t* results_data;
	// C.wasm_functype_t
	// Note: The actual structure is defined in the C library, this is just a placeholder.
}

// FuncType is one of function types which classify the signature of functions, mapping a vector of parameters to a vector of results.
// They are also used to classify the inputs and outputs of instructions.
type FuncType struct {
	_ptr   uintptr //*wasm_functype_t
	_owner interface{}
}

// NewFuncType creates a new `FuncType` with the `kind` provided
func NewFuncType(params, results []*ValType) *FuncType {
	paramVec := mkValTypeList(params)
	resultVec := mkValTypeList(results)

	ptr := wasm_functype_new(&paramVec, &resultVec)

	return mkFuncType(ptr, nil)
}

func mkValTypeList(tys []*ValType) wasm_valtype_vec_t {
	var vec wasm_valtype_vec_t
	wasm_valtype_vec_new_uninitialized(&vec, len(tys))
	base := unsafe.Pointer(vec.data)
	for i, ty := range tys {
		ptr := wasm_valtype_new(wasm_valtype_kind(ty.ptr()))
		*(**wasm_valtype_t)(unsafe.Pointer(uintptr(base) + unsafe.Sizeof(ptr)*uintptr(i))) = (*wasm_valtype_t)(unsafe.Pointer(ptr))
	}
	runtime.KeepAlive(tys)
	return vec
}

func mkFuncType(ptr uintptr, owner interface{}) *FuncType {
	functype := &FuncType{_ptr: ptr, _owner: owner}
	if owner == nil {
		runtime.SetFinalizer(functype, func(functype *FuncType) {
			functype.Close()
		})
	}
	return functype
}

func (ty *FuncType) ptr() uintptr {
	ret := ty._ptr
	if ret == uintptr(0) {
		panic("object has been closed already")
	}
	//maybeGC()
	return ret
}

func (ty *FuncType) owner() interface{} {
	if ty._owner != nil {
		return ty._owner
	}
	return ty
}

// Close will deallocate this type's state explicitly.
//
// For more information see the documentation for engine.Close()
func (ty *FuncType) Close() {
	if ty._ptr == uintptr(0) || ty._owner != nil {
		return
	}
	runtime.SetFinalizer(ty, nil)
	wasm_functype_delete(ty._ptr)
	ty._ptr = uintptr(0)
}

// Params returns the parameter types of this function type
func (ty *FuncType) Params() []*ValType {
	ptr := wasm_functype_params(uintptr(ty.ptr()))
	return ty.convertTypeList((*wasm_valtype_vec_t)(unsafe.Pointer(ptr)))
}

// Results returns the result types of this function type
func (ty *FuncType) Results() []*ValType {
	ptr := wasm_functype_results(uintptr(ty.ptr()))
	return ty.convertTypeList((*wasm_valtype_vec_t)(unsafe.Pointer(ptr)))
}

func (ty *FuncType) convertTypeList(list *wasm_valtype_vec_t) []*ValType {
	ret := make([]*ValType, list.size)

	base := unsafe.Pointer(list.data)
	var ptr *wasm_valtype_t
	for i := 0; i < int(list.size); i++ {
		ptr := *(**wasm_valtype_t)(unsafe.Pointer(uintptr(base) + unsafe.Sizeof(ptr)*uintptr(i)))
		ty := mkValType(uintptr(unsafe.Pointer(ptr)), ty.owner())
		ret[i] = ty
	}
	return ret
}

// AsExternType converts this type to an instance of `ExternType`
func (ty *FuncType) AsExternType() *ExternType {
	ptr := wasm_functype_as_externtype_const(ty.ptr())
	return mkExternType(ptr, ty.owner())
}
