package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModule(t *testing.T) {
	_, err := NewModule(NewEngine(), []byte{})
	require.Error(t, err)
	_, err = NewModule(NewEngine(), []byte{1})
	require.Error(t, err)
}
