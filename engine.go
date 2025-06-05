package wasmtime

import (
	"runtime"
	"unsafe"
)

// Engine is an instance of a wasmtime engine which is used to create a `Store`.
//
// Engines are a form of global configuration for wasm compilations and modules
// and such.
type Engine struct {
	_ptr unsafe.Pointer //*C.wasm_engine_t
}

// NewEngine creates a new `Engine` with default configuration.
func NewEngine() *Engine {
	engine := &Engine{_ptr: unsafe.Pointer(wasm_engine_new())}
	runtime.SetFinalizer(engine, func(engine *Engine) {
		engine.Close()
	})
	return engine
}

// Close will deallocate this engine's state explicitly.
//
// By default state is cleaned up automatically when an engine is garbage
// collected but the Go GC. The Go GC, however, does not provide strict
// guarantees about finalizers especially in terms of timing. Additionally the
// Go GC is not aware of the full weight of an engine because it holds onto
// allocations in Wasmtime not tracked by the Go GC. For these reasons, it's
// recommended to where possible explicitly call this method and deallocate an
// engine to avoid relying on the Go GC.
//
// This method will deallocate Wasmtime-owned state. Future use of the engine
// will panic because the Wasmtime state is no longer there.
//
// Close can be called multiple times without error. Only the first time will
// deallocate resources.
func (engine *Engine) Close() {
	if engine._ptr == nil {
		return
	}
	runtime.SetFinalizer(engine, nil)
	wasm_engine_delete(uintptr(engine.ptr()))
	engine._ptr = nil

}

func (engine *Engine) ptr() unsafe.Pointer {
	ret := engine._ptr
	if ret == nil {
		panic("object has been closed already")
	}
	//maybeGC()
	return ret
}
