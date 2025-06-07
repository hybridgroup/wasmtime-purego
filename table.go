package wasmtime

import (
	"errors"
	"runtime"
	"unsafe"
)

type wasmtime_table_t struct {
	_ struct {
		/// Internal identifier of what store this belongs to, never zero.
		store_id uint64
		/// Private field for Wasmtime.
		__private1 uint32
	}
	/// Private field for Wasmtime.
	__private2 uint32
}

// Table is a table instance, which is the runtime representation of a table.
//
// It holds a vector of reference types and an optional maximum size, if one was
// specified in the table type at the table’s definition site.
// Read more in [spec](https://webassembly.github.io/spec/core/exec/runtime.html#table-instances)
type Table struct {
	val uintptr // C.wasmtime_table_t
}

// NewTable creates a new `Table` in the given `Store` with the specified `ty`.
//
// The `ty` must be a reference type (`funref` or `externref`) and `init`
// is the initial value for all table slots and must have the type specified by
// `ty`.
func NewTable(store Storelike, ty *TableType, init Val) (*Table, error) {
	var ret wasmtime_table_t
	var raw_val wasmtime_val_t
	init.initialize(store, &raw_val)
	err := wasmtime_table_new(uintptr(store.Context()), ty.ptr(), &raw_val, &ret)
	wasmtime_val_unroot(uintptr(store.Context()), uintptr(unsafe.Pointer(&raw_val)))
	runtime.KeepAlive(store)
	runtime.KeepAlive(ty)
	if err != uintptr(0) {
		return nil, mkError(err)
	}
	return mkTable(uintptr(unsafe.Pointer(&ret))), nil
}

func mkTable(val uintptr) *Table {
	return &Table{val}
}

// Size returns the size of this table in units of elements.
func (t *Table) Size(store Storelike) uint64 {
	ret := wasmtime_table_size(uintptr(store.Context()), t.val)
	runtime.KeepAlive(store)
	return ret
}

// Grow grows this table by the number of units specified, using the
// specified initializer value for new slots.
//
// Returns an error if the table failed to grow, or the previous size of the
// table if growth was successful.
func (t *Table) Grow(store Storelike, delta uint64, init Val) (uint64, error) {
	var prev uint64
	var raw_val wasmtime_val_t
	init.initialize(store, &raw_val)
	err := wasmtime_table_grow(uintptr(store.Context()), t.val, delta, &raw_val, &prev)
	wasmtime_val_unroot(uintptr(store.Context()), uintptr(unsafe.Pointer(&raw_val)))
	runtime.KeepAlive(store)
	if err != uintptr(0) {
		return 0, mkError(err)
	}

	return prev, nil
}

// Get gets an item from this table from the specified index.
//
// Returns an error if the index is out of bounds, or returns a value (which
// may be internally null) if the index is in bounds corresponding to the entry
// at the specified index.
func (t *Table) Get(store Storelike, idx uint64) (Val, error) {
	var val wasmtime_val_t
	ok := wasmtime_table_get(uintptr(store.Context()), t.val, idx, &val)
	runtime.KeepAlive(store)
	if !ok {
		return Val{}, errors.New("index out of bounds")
	}
	return takeVal(store, &val), nil
}

// Set sets an item in this table at the specified index.
//
// Returns an error if the index is out of bounds.
func (t *Table) Set(store Storelike, idx uint64, val Val) error {
	var raw_val wasmtime_val_t
	val.initialize(store, &raw_val)
	err := wasmtime_table_set(uintptr(store.Context()), t.val, idx, &raw_val)
	wasmtime_val_unroot(uintptr(store.Context()), uintptr(unsafe.Pointer(&raw_val)))
	runtime.KeepAlive(store)
	if err != uintptr(0) {
		return mkError(err)
	}
	return nil
}

// Type returns the underlying type of this table
func (t *Table) Type(store Storelike) *TableType {
	ptr := wasmtime_table_type(uintptr(store.Context()), t.val)
	runtime.KeepAlive(store)
	return mkTableType(ptr, nil)
}

func (t *Table) AsExtern() wasmtime_extern_t {
	ret := wasmtime_extern_t{kind: WASMTIME_EXTERN_TABLE}
	go_wasmtime_extern_table_set(&ret, t.val)
	return ret
}
