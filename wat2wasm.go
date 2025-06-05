package wasmtime

import (
	"runtime"
	"unsafe"
)

type wasm_byte_vec_t struct {
	size uint32 // C.size_t
	data *uint8 // *C.uint8_t
}

// Wat2Wasm converts the text format of WebAssembly to the binary format.
//
// Takes the text format in-memory as input, and returns either the binary
// encoding of the text format or an error if parsing fails.
func Wat2Wasm(wat string) ([]byte, error) {
	var retVec wasm_byte_vec_t
	err := wasmtime_wat2wasm(wat, len(wat), &retVec)
	runtime.KeepAlive(wat)

	if err == 0 {
		ret := make([]byte, retVec.size)
		copy(ret, unsafe.Slice(retVec.data, retVec.size)) // Convert the pointer to a slice
		wasm_byte_vec_delete(&retVec)
		return ret, nil
	}

	return nil, mkError(unsafe.Pointer(err))
}
