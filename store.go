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
	_ptr unsafe.Pointer //*C.wasmtime_store_t

	// The `Engine` that this store uses for compilation and environment
	// settings.
	Engine *Engine
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
	//callback func(*Caller, []Val) ([]Val, *Trap)
	//results  []*ValType
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
