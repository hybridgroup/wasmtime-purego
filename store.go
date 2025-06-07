package wasmtime

import (
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

// Store is a general group of wasm instances, and many objects
// must all be created with and reference the same `Store`
type Store struct {
	_ptr unsafe.Pointer // *C.wasmtime_store_t

	// The `Engine` that this store uses for compilation and environment
	// settings.
	Engine *Engine
}

// Storelike represents types that can be used to contextually reference a
// `Store`.
//
// This interface is implemented by `*Store` and `*Caller` and is pervasively
// used throughout this library. You'll want to pass one of those two objects
// into functions that take a `Storelike`.
type Storelike interface {
	// Returns the wasmtime context pointer this store is attached to.
	Context() uintptr // *C.wasmtime_context_t
}

var gStoreLock sync.Mutex
var gStoreMap = make(map[int]*storeData)
var gStoreSlab slab

// State associated with a `Store`, currently used to propagate panic
// information through invocations as well as store Go closures that have been
// added to the store.
type storeData struct {
	engine    *Engine
	funcNew   []funcNewEntry
	funcWrap  []funcWrapEntry
	lastPanic interface{}
}

type funcNewEntry struct {
	callback func(*Caller, []Val) ([]Val, *Trap)
	results  []*ValType
}

type funcWrapEntry struct {
	callback reflect.Value
}

// NewStore creates a new `Store` from the configuration provided in `engine`
func NewStore(engine *Engine) *Store {
	// Allocate an index for this store and allocate some internal data to go with
	// the store.
	gStoreLock.Lock()
	idx := gStoreSlab.allocate()
	gStoreMap[idx] = &storeData{engine: engine}
	gStoreLock.Unlock()

	ptr := wasmtime_store_new(uintptr(engine.ptr()), idx) // C.go_store_new(engine.ptr(), C.size_t(idx))
	store := &Store{
		_ptr:   unsafe.Pointer(ptr),
		Engine: engine,
	}
	runtime.SetFinalizer(store, func(store *Store) {
		store.Close()
	})
	return store
}

func (store *Store) ptr() uintptr {
	ret := store._ptr
	if ret == nil {
		panic("object has been closed already")
	}
	//maybeGC()
	return uintptr(ret)
}

// Close will deallocate this store's state explicitly.
//
// For more information see the documentation for engine.Close()
func (store *Store) Close() {
	if store._ptr == nil {
		return
	}
	runtime.SetFinalizer(store, nil)
	wasmtime_store_delete(uintptr(store._ptr))
	store._ptr = nil
}

// Implementation of the `Storelike` interface
func (store *Store) Context() uintptr {
	ret := wasmtime_store_context(store.ptr())
	//maybeGC()
	runtime.KeepAlive(store)
	return ret
}

//export goFinalizeStore
func goFinalizeStore(env unsafe.Pointer) {
	// When a store is finalized this is used as the finalization callback for the
	// custom data within the store, and our finalization here will delete the
	// store's data from the global map and deallocate its index to get reused by
	// a future store.
	idx := int(uintptr(env))
	gStoreLock.Lock()
	defer gStoreLock.Unlock()
	delete(gStoreMap, idx)
	gStoreSlab.deallocate(idx)
}

// Returns the underlying `*storeData` that this store references in Go, used
// for inserting functions or storing panic data.
func getDataInStore(store Storelike) *storeData {
	data := uintptr(wasmtime_context_get_data(uintptr(store.Context())))
	gStoreLock.Lock()
	defer gStoreLock.Unlock()
	return gStoreMap[int(data)]
}

var gEngineFuncLock sync.Mutex
var gEngineFuncNew = make(map[int]*funcNewEntry)
var gEngineFuncNewSlab slab
var gEngineFuncWrap = make(map[int]*funcWrapEntry)
var gEngineFuncWrapSlab slab

func insertFuncNew(data *storeData, ty *FuncType, callback func(*Caller, []Val) ([]Val, *Trap)) int {
	var idx int
	entry := funcNewEntry{
		callback: callback,
		results:  ty.Results(),
	}
	if data == nil {
		gEngineFuncLock.Lock()
		defer gEngineFuncLock.Unlock()
		idx = gEngineFuncNewSlab.allocate()
		gEngineFuncNew[idx] = &entry
		idx = (idx << 1)
	} else {
		idx = len(data.funcNew)
		data.funcNew = append(data.funcNew, entry)
		idx = (idx << 1) | 1
	}
	return idx
}

func (data *storeData) getFuncNew(idx int) *funcNewEntry {
	if idx&1 == 0 {
		gEngineFuncLock.Lock()
		defer gEngineFuncLock.Unlock()
		return gEngineFuncNew[idx>>1]
	} else {
		return &data.funcNew[idx>>1]
	}
}

func insertFuncWrap(data *storeData, callback reflect.Value) int {
	var idx int
	entry := funcWrapEntry{callback}
	if data == nil {
		gEngineFuncLock.Lock()
		defer gEngineFuncLock.Unlock()
		idx = gEngineFuncWrapSlab.allocate()
		gEngineFuncWrap[idx] = &entry
		idx = (idx << 1)
	} else {
		idx = len(data.funcWrap)
		data.funcWrap = append(data.funcWrap, entry)
		idx = (idx << 1) | 1
	}
	return idx

}

func (data *storeData) getFuncWrap(idx int) *funcWrapEntry {
	if idx&1 == 0 {
		gEngineFuncLock.Lock()
		defer gEngineFuncLock.Unlock()
		return gEngineFuncWrap[idx>>1]
	} else {
		return &data.funcWrap[idx>>1]
	}
}
