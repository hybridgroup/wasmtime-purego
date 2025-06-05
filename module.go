package wasmtime

import (
	"runtime"
	"unsafe"
)

// Module is a module which collects definitions for types, functions, tables, memories, and globals.
// In addition, it can declare imports and exports and provide initialization logic in the form of data and element segments or a start function.
// Modules organized WebAssembly programs as the unit of deployment, loading, and compilation.
type Module struct {
	_ptr unsafe.Pointer //*C.wasmtime_module_t
}

// NewModule compiles a new `Module` from the `wasm` provided with the given configuration
// in `engine`.
func NewModule(engine *Engine, wasm []byte) (*Module, error) {
	var ptr uintptr //*C.wasmtime_module_t
	err := wasmtime_module_new(uintptr(engine.ptr()), wasm, len(wasm), &ptr)
	runtime.KeepAlive(engine)
	runtime.KeepAlive(wasm)

	if err != 0 {
		return nil, mkError(unsafe.Pointer(err))
	}

	return mkModule(unsafe.Pointer(ptr)), nil
}

func mkModule(ptr unsafe.Pointer) *Module {
	module := &Module{_ptr: ptr}
	runtime.SetFinalizer(module, func(module *Module) {
		module.Close()
	})
	return module
}

// Close will deallocate this module's state explicitly.
//
// For more information see the documentation for engine.Close()
func (m *Module) Close() {
	if m._ptr == nil {
		return
	}
	runtime.SetFinalizer(m, nil)
	wasmtime_module_delete(uintptr(m._ptr))
	m._ptr = nil
}
