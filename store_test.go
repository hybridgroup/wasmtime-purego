package wasmtime

import (
	"testing"
)

func TestStore(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()
	store := NewStore(engine)
	defer store.Close()
}
