package wasmtime

import (
	"runtime"
	"unsafe"
)

type wasmtime_memory_t struct {
	_ struct {
		/// Internal identifier of what store this belongs to, never zero.
		store_id uint64
		/// Private field for Wasmtime.
		__private1 uint32
	}
	/// Private field for Wasmtime.
	__private2 uint32
}

// Memory instance is the runtime representation of a linear memory.
// It holds a vector of bytes and an optional maximum size, if one was specified at the definition site of the memory.
// Read more in [spec](https://webassembly.github.io/spec/core/exec/runtime.html#memory-instances)
// In wasmtime-go, you can get the vector of bytes by the unsafe pointer of memory from `Memory.Data()`, or go style byte slice from `Memory.UnsafeData()`
type Memory struct {
	val uintptr // C.wasmtime_memory_t
}

// NewMemory creates a new `Memory` in the given `Store` with the specified `ty`.
func NewMemory(store Storelike, ty *MemoryType) (*Memory, error) {
	var ret wasmtime_memory_t
	err := wasmtime_memory_new(uintptr(store.Context()), ty.ptr(), &ret)
	runtime.KeepAlive(store)
	runtime.KeepAlive(ty)
	if err != uintptr(0) {
		return nil, mkError(err)
	}
	return mkMemory(uintptr(unsafe.Pointer(&ret))), nil
}

func mkMemory(val uintptr) *Memory {
	return &Memory{val}
}

// Type returns the type of this memory
func (mem *Memory) Type(store Storelike) *MemoryType {
	ptr := wasmtime_memory_type(uintptr(store.Context()), mem.val)
	runtime.KeepAlive(store)
	return mkMemoryType(ptr, nil)
}

// Data returns the raw pointer in memory of where this memory starts
func (mem *Memory) Data(store Storelike) unsafe.Pointer {
	ret := unsafe.Pointer(wasmtime_memory_data(uintptr(store.Context()), mem.val))
	runtime.KeepAlive(store)
	return ret
}

// UnsafeData returns the raw memory backed by this `Memory` as a byte slice (`[]byte`).
//
// This is not a safe method to call, hence the "unsafe" in the name. The byte
// slice returned from this function is not managed by the Go garbage collector.
// You need to ensure that `m`, the original `Memory`, lives longer than the
// `[]byte` returned.
//
// Note that you may need to use `runtime.KeepAlive` to keep the original memory
// `m` alive for long enough while you're using the `[]byte` slice. If the
// `[]byte` slice is used after `m` is GC'd then that is undefined behavior.
func (mem *Memory) UnsafeData(store Storelike) []byte {
	length := mem.DataSize(store)
	return unsafe.Slice((*byte)(mem.Data(store)), length)
}

// DataSize returns the size, in bytes, that `Data()` is valid for
func (mem *Memory) DataSize(store Storelike) uintptr {
	ret := uintptr(wasmtime_memory_data_size(uintptr(store.Context()), mem.val))
	runtime.KeepAlive(store)
	return ret
}

// Size returns the size, in wasm pages, of this memory
func (mem *Memory) Size(store Storelike) uint64 {
	ret := uint64(wasmtime_memory_size(uintptr(store.Context()), mem.val))
	runtime.KeepAlive(store)
	return ret
}

// Grow grows this memory by `delta` pages
func (mem *Memory) Grow(store Storelike, delta uint64) (uint64, error) {
	prev := uint64(0)
	err := wasmtime_memory_grow(uintptr(store.Context()), mem.val, delta, &prev)
	runtime.KeepAlive(store)
	if err != uintptr(0) {
		return 0, mkError(err)
	}
	return uint64(prev), nil
}

func (mem *Memory) AsExtern() wasmtime_extern_t {
	ret := wasmtime_extern_t{kind: WASMTIME_EXTERN_MEMORY}
	go_wasmtime_extern_memory_set(&ret, mem.val)
	return ret
}
