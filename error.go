package wasmtime

import (
	"runtime"
	"unsafe"
)

type Error struct {
	_ptr unsafe.Pointer //*C.wasmtime_error_t
}

func mkError(ptr unsafe.Pointer) *Error {
	err := &Error{_ptr: ptr}
	runtime.SetFinalizer(err, func(err *Error) {
		err.Close()
	})
	return err
}

// Close will deallocate this error's state explicitly.
//
// For more information see the documentation for engine.Close()
func (e *Error) Close() {
	if e._ptr == nil {
		return
	}
	runtime.SetFinalizer(e, nil)
	wasmtime_error_delete(uintptr(e._ptr))
	e._ptr = nil
}

func (e *Error) Error() string {
	return "need error message implementation"
}
