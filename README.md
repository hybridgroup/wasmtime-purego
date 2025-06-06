# wasmtime-purego

This package in a experiment to rewrite the Go wrappers for [Wasmtime](https://github.com/bytecodealliance/wasmtime) using [purego](https://github.com/ebitengine/purego).

Why?

- no more cgo, so projects like envoy and k8s can use wasmtime
- dynamic linking to wasmtime could reduce update fragility since the C interface is being used for integration anyhow.
- should be able to run on linux, macos, and windows making cross-compilation much easier without needing a builder container

## Installation

You must install the wasmtime dynamic libs before running code using this package.

## Usage

The idea is to be able to run code just like the `wasmtime-go` package, except without needing CGo.

The following example does not yet work...

```go
package main

import (
    "fmt"
    "github.com/hybridgroup/wasmtime"
)

func main() {
    // Almost all operations in wasmtime require a contextual `store`
    // argument to share, so create that first
    store := wasmtime.NewStore(wasmtime.NewEngine())

    // Compiling modules requires WebAssembly binary input, but the wasmtime
    // package also supports converting the WebAssembly text format to the
    // binary format.
    wasm, err := wasmtime.Wat2Wasm(`
      (module
        (import "" "hello" (func $hello))
        (func (export "run")
          (call $hello))
      )
    `)
    check(err)

    // Once we have our binary `wasm` we can compile that into a `*Module`
    // which represents compiled JIT code.
    module, err := wasmtime.NewModule(store.Engine, wasm)
    check(err)

    // Our `hello.wat` file imports one item, so we create that function
    // here.
    item := wasmtime.WrapFunc(store, func() {
        fmt.Println("Hello from Go!")
    })

    // Next up we instantiate a module which is where we link in all our
    // imports. We've got one import so we pass that in here.
    instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{item})
    check(err)

    // After we've instantiated we can lookup our `run` function and call
    // it.
    run := instance.GetFunc(store, "run")
    if run == nil {
        panic("not a function")
    }
    _, err = run.Call(store)
    check(err)
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}
```

## TODO to make the above example run:

- [X] `NewEngine()`
- [X] `NewStore()`
- [X] `Wat2Wasm()`
- [X] `NewModule()`
- [X] `WrapFunc()`
    - [X] `ValType`
    - [X] `ValKind`
    - [X] `Val`
    - [X] `Func`
    - [X] `FuncType`
    - [X] `Trap`
    - [X] `shims`
    - [X] `Caller`
    - [X] `Error`
- [ ] `NewInstance()`
    - [ ] `Extern`
    - [ ] `ImportType`
- [ ] `GetFunc()`
- [ ] `Call()`

## Credits

The code in this package is a combination of code modified from the https://github.com/bytecodealliance/wasmtime-go package, along with purego code taken from https://github.com/kelindar/search

Thanks everyone!
