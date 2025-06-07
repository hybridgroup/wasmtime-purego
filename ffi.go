package wasmtime

import (
	"runtime"
	"unsafe"
)

// Convert a Go string into an owned `wasm_byte_vec_t`
func stringToByteVec(s string) wasm_byte_vec_t {
	vec := wasm_byte_vec_t{}
	wasm_byte_vec_new_uninitialized(&vec, len(s))
	copy(unsafe.Slice(vec.data, vec.size), s) // Convert the pointer to a slice
	runtime.KeepAlive(s)
	return vec
}
