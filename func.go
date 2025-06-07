package wasmtime

import (
	"reflect"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Func is a function instance, which is the runtime representation of a function.
// It effectively is a closure of the original function over the runtime module instance of its originating module.
// The module instance is used to resolve references to other definitions during execution of the function.
// Read more in [spec](https://webassembly.github.io/spec/core/exec/runtime.html#function-instances)
type Func struct {
	val uintptr // C.wasmtime_func_t
}

// Caller is provided to host-defined functions when they're invoked from
// WebAssembly.
//
// A `Caller` can be used for `Storelike` arguments to allow recursive execution
// or creation of wasm objects. Additionally `Caller` can be used to learn about
// the exports of the calling instance.
type Caller struct {
	// Note that unlike other structures in these bindings this is named `ptr`
	// instead of `_ptr` because no finalizer is configured with `Caller` so it's
	// ok to access this raw value.
	ptr uintptr //*C.wasmtime_caller_t
}

type wasmtime_func_t struct {
	/// Internal identifier of what store this belongs to, never zero.
	store_id uint64
	/// Internal index within the store.
	index int
}

// NewFunc creates a new `Func` with the given `ty` which, when called, will call `f`
//
// The `ty` given is the wasm type signature of the `Func` to create. When called
// the `f` callback receives two arguments. The first is a `Caller` to learn
// information about the calling context and the second is a list of arguments
// represented as a `Val`. The parameters are guaranteed to match the parameters
// types specified in `ty`.
//
// The `f` callback is expected to produce one of two values. Results can be
// returned as an array of `[]Val`. The number and types of these results much
// match the `ty` given, otherwise the program will panic. The `f` callback can
// also produce a trap which will trigger trap unwinding in wasm, and the trap
// will be returned to the original caller.
//
// If the `f` callback panics then the panic will be propagated to the caller
// as well.
func NewFunc(
	store Storelike,
	ty *FuncType,
	f func(*Caller, []Val) ([]Val, *Trap),
) *Func {
	idx := insertFuncNew(getDataInStore(store), ty, f)

	var ret wasmtime_func_t
	wasmtime_func_new(
		store.Context(),
		ty.ptr(),
		purego.NewCallback(goTrampolineNew),
		idx,
		0, // this is `NewFunc`
		&ret,
	)

	runtime.KeepAlive(store)
	runtime.KeepAlive(ty)

	return mkFunc(uintptr(unsafe.Pointer(&ret)))
}

//export goTrampolineNew
func goTrampolineNew(
	callerPtr uintptr,
	env int,
	argsPtr uintptr,
	argsNum int,
	resultsPtr uintptr,
	resultsNum int) uintptr {
	caller := &Caller{ptr: callerPtr}
	defer func() { caller.ptr = uintptr(0) }()
	data := getDataInStore(caller)
	entry := data.getFuncNew(int(env))

	params := make([]Val, int(argsNum))
	var val wasmtime_val_t
	base := unsafe.Pointer(argsPtr)
	for i := 0; i < len(params); i++ {
		ptr := (*wasmtime_val_t)(unsafe.Pointer(uintptr(base) + uintptr(i)*unsafe.Sizeof(val)))
		params[i] = mkVal(caller, ptr)
	}

	var results []Val
	var trap *Trap
	var lastPanic interface{}
	func() {
		defer func() { lastPanic = recover() }()
		results, trap = entry.callback(caller, params)
		if trap != nil {
			if trap._ptr == uintptr(0) {
				panic("returned an already-returned trap")
			}
			return
		}
		if len(results) != len(entry.results) {
			panic("callback didn't produce the correct number of results")
		}
		for i, ty := range entry.results {
			if results[i].Kind() != ty.Kind() {
				panic("callback produced wrong type of result")
			}
		}
	}()
	if trap == nil && lastPanic != nil {
		data.lastPanic = lastPanic
		trap := NewTrap("go panicked")
		runtime.SetFinalizer(trap, nil)
		return uintptr(trap.ptr())
	}
	if trap != nil {
		runtime.SetFinalizer(trap, nil)
		ret := trap.ptr()
		trap._ptr = uintptr(0)
		return uintptr(ret)
	}

	base = unsafe.Pointer(resultsPtr)
	for i := 0; i < len(results); i++ {
		ptr := (*wasmtime_val_t)(unsafe.Pointer(uintptr(base) + uintptr(i)*unsafe.Sizeof(val)))
		results[i].initialize(caller, ptr)
	}
	runtime.KeepAlive(results)
	return uintptr(0)
}

// WrapFunc wraps a native Go function, `f`, as a wasm `Func`.
//
// This function differs from `NewFunc` in that it will determine the type
// signature of the wasm function given the input value `f`. The `f` value
// provided must be a Go function. It may take any number of the following
// types as arguments:
//
// `int32` - a wasm `i32`
//
// `int64` - a wasm `i64`
//
// `float32` - a wasm `f32`
//
// `float64` - a wasm `f64`
//
// `*Caller` - information about the caller's instance
//
// `*Func` - a wasm `funcref`
//
// anything else - a wasm `externref`
//
// The Go function may return any number of values. It can return any number of
// primitive wasm values (integers/floats), and the last return value may
// optionally be `*Trap`. If a `*Trap` returned is `nil` then the other values
// are returned from the wasm function. Otherwise the `*Trap` is returned and
// it's considered as if the host function trapped.
//
// If the function `f` panics then the panic will be propagated to the caller.
func WrapFunc(
	store Storelike,
	f interface{},
) *Func {
	val := reflect.ValueOf(f)
	wasmTy := inferFuncType(val)
	idx := insertFuncWrap(getDataInStore(store), val)

	var ret wasmtime_func_t
	wasmtime_func_new(
		store.Context(),
		wasmTy.ptr(),
		purego.NewCallback(goTrampolineWrap),
		idx,
		0, // this is `WrapFunc`, not `NewFunc`
		&ret,
	)
	runtime.KeepAlive(store)
	runtime.KeepAlive(wasmTy)
	return mkFunc(uintptr(unsafe.Pointer(&ret)))
}

func inferFuncType(val reflect.Value) *FuncType {
	// Make sure the `interface{}` passed in was indeed a function
	ty := val.Type()
	if ty.Kind() != reflect.Func {
		panic("callback provided must be a `func`")
	}

	// infer the parameter types, and `*Caller` type is special in the
	// parameters so be sure to case on that as well.
	params := make([]*ValType, 0, ty.NumIn())
	var caller *Caller
	for i := 0; i < ty.NumIn(); i++ {
		paramTy := ty.In(i)
		if paramTy != reflect.TypeOf(caller) {
			params = append(params, typeToValType(paramTy))
		}
	}

	// Then infer the result types, where a final `*Trap` result value is
	// also special.
	results := make([]*ValType, 0, ty.NumOut())
	var trap *Trap
	for i := 0; i < ty.NumOut(); i++ {
		resultTy := ty.Out(i)
		if i == ty.NumOut()-1 && resultTy == reflect.TypeOf(trap) {
			continue
		}
		results = append(results, typeToValType(resultTy))
	}
	return NewFuncType(params, results)
}

func typeToValType(ty reflect.Type) *ValType {
	var a int32
	if ty == reflect.TypeOf(a) {
		return NewValType(KindI32)
	}
	var b int64
	if ty == reflect.TypeOf(b) {
		return NewValType(KindI64)
	}
	var c float32
	if ty == reflect.TypeOf(c) {
		return NewValType(KindF32)
	}
	var d float64
	if ty == reflect.TypeOf(d) {
		return NewValType(KindF64)
	}
	var f *Func
	if ty == reflect.TypeOf(f) {
		return NewValType(KindFuncref)
	}
	return NewValType(KindExternref)
}

//export goTrampolineWrap
func goTrampolineWrap(
	callerPtr uintptr,
	env int,
	argsPtr uintptr,
	argsNum int,
	resultsPtr uintptr,
	resultsNum int) uintptr {
	// Convert all our parameters to `[]reflect.Value`, taking special care
	// for `*Caller` but otherwise reading everything through `Val`.
	caller := &Caller{ptr: callerPtr}
	defer func() { caller.ptr = uintptr(0) }()
	data := getDataInStore(caller)
	entry := data.getFuncWrap(int(env))

	ty := entry.callback.Type()
	params := make([]reflect.Value, ty.NumIn())
	base := unsafe.Pointer(argsPtr)
	var raw wasmtime_val_t
	for i := 0; i < len(params); i++ {
		if ty.In(i) == reflect.TypeOf(caller) {
			params[i] = reflect.ValueOf(caller)
		} else {
			ptr := (*wasmtime_val_t)(base)
			val := mkVal(caller, ptr)
			params[i] = reflect.ValueOf(val.Get())
			base = unsafe.Pointer(uintptr(base) + unsafe.Sizeof(raw))
		}
	}

	// Invoke the function, catching any panics to propagate later. Panics
	// result in immediately returning a trap.
	var results []reflect.Value
	var lastPanic interface{}
	func() {
		defer func() { lastPanic = recover() }()
		results = entry.callback.Call(params)
	}()
	if lastPanic != nil {
		data.lastPanic = lastPanic
		trap := NewTrap("go panicked")
		runtime.SetFinalizer(trap, nil)
		return uintptr(trap.ptr())
	}

	// And now we write all the results into memory depending on the type
	// of value that was returned.
	base = unsafe.Pointer(resultsPtr)
	for _, result := range results {
		ptr := (*wasmtime_val_t)(base)
		switch val := result.Interface().(type) {
		case int32:
			ValI32(val).initialize(caller, ptr)
		case int64:
			ValI64(val).initialize(caller, ptr)
		case float32:
			ValF32(val).initialize(caller, ptr)
		case float64:
			ValF64(val).initialize(caller, ptr)
		case *Func:
			ValFuncref(val).initialize(caller, ptr)
		case *Trap:
			if val != nil {
				runtime.SetFinalizer(val, nil)
				ret := val._ptr
				val._ptr = uintptr(0)
				if ret == uintptr(0) {
					data.lastPanic = "cannot return trap twice"
					return uintptr(0)
				} else {
					return uintptr(ret)
				}
			}
		default:
			ValExternref(val).initialize(caller, ptr)
		}
		base = unsafe.Pointer(uintptr(base) + unsafe.Sizeof(raw))
	}
	return uintptr(0)
}

func mkFunc(val uintptr) *Func {
	return &Func{val}
}

// Type returns the type of this func
func (f *Func) Type(store Storelike) *FuncType {
	ptr := wasmtime_func_type(store.Context(), f.val)
	runtime.KeepAlive(store)
	return mkFuncType(ptr, nil)
}

// Implementation of the `AsExtern` interface for `Func`
func (f *Func) AsExtern() wasmtime_extern_t {
	ret := wasmtime_extern_t{kind: WASMTIME_EXTERN_FUNC}
	go_wasmtime_extern_func_set(&ret, uintptr(f.val))
	return ret
}

// Implementation of the `Storelike` interface for `Caller`.
func (c *Caller) Context() uintptr {
	if c.ptr == uintptr(0) {
		panic("cannot use caller after host function returns")
	}
	return wasmtime_caller_context(c.ptr)
}

// Shim function that's expected to wrap any invocations of WebAssembly from Go
// itself.
//
// This is used to handle traps and error returns from any invocation of
// WebAssembly. This will also automatically propagate panics that happen within
// Go from one end back to this original invocation point.
//
// The `store` object is the context being used for the invocation, and `wasm`
// is the closure which will internally execute WebAssembly. A trap pointer is
// provided to the closure and it's expected that the closure returns an error.
func enterWasm(store Storelike, wasm func(**wasm_trap_t) uintptr) error {
	// Load the internal `storeData` that our `store` references, which is
	// used for handling panics which we are going to use here.
	data := getDataInStore(store)

	var trap *wasm_trap_t
	err := wasm(&trap)

	// Take ownership of any returned values to ensure we properly run
	// destructors for them.
	var wrappedTrap *Trap
	var wrappedError error
	if trap != nil {
		wrappedTrap = mkTrap(uintptr(unsafe.Pointer(trap)))
	}
	if err != uintptr(0) {
		wrappedError = mkError(err)
	}

	// Check to see if wasm panicked, and if it did then we need to
	// propagate that. Note that this happens after we take ownership of
	// return values to ensure they're cleaned up properly.
	if data.lastPanic != nil {
		lastPanic := data.lastPanic
		data.lastPanic = nil
		panic(lastPanic)
	}

	// If there wasn't a panic then we determine whether to return the trap
	// or the error.
	if wrappedTrap != nil {
		return wrappedTrap
	}
	return wrappedError
}
