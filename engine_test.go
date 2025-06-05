package wasmtime

import (
	"testing"
)

func TestEngine(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()
	//NewEngineWithConfig(NewConfig())
	//engine.IsPulley()
}
