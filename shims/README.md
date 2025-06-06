# wasmtime-shims

Shared library to get around use of the `union` type in Go.

## Use

### Linux

Copy the file `libwasmtime-shims.so` to the `/usr/local/lib` directory.

### macOS

Coming soon...

### Windows

Coming soon...

## Build

If you need to rebuild the shims library, here is how:

```
cd shims
export WASMTIME_INCLUDE=~/Downloads/wasmtime-v33.0.0-x86_64-linux-c-api/include
gcc shims.c -c -Wall -Werror -fpic -I ${WASMTIME_INCLUDE} \
       /usr/local/lib/libwasmtime.a \
       -lpthread -ldl -lm \
       -o shims
gcc -shared -o libwasmtime-shims.so shims.o
```
