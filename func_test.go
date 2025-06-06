package wasmtime

import (
	"testing"
)

func TestFunc(t *testing.T) {
	store := NewStore(NewEngine())
	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		return []Val{}, nil
	}
	NewFunc(store, NewFuncType([]*ValType{}, []*ValType{}), cb)
}

func TestWrapFunc(t *testing.T) {
	store := NewStore(NewEngine())
	WrapFunc(store, func() {
		var called bool
		for i := 0; i < 10; i++ {
			called = !called
		}
	})
}
