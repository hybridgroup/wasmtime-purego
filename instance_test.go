package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstanceBad(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(module (import "" "" (func)))`)
	require.NoError(t, err)
	module, err := NewModule(NewEngine(), wasm)
	require.NoError(t, err)

	// wrong number of imports
	instance, err := NewInstance(store, module, []AsExtern{})
	require.Nil(t, instance)
	require.Error(t, err)

	// wrong types of imports
	f := WrapFunc(store, func(a int32) {})
	instance, err = NewInstance(store, module, []AsExtern{f})
	require.Nil(t, instance)
	require.Error(t, err)
}
