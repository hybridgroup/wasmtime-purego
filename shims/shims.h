#include <wasm.h>
#include <wasmtime.h>

#define EACH_UNION_ACCESSOR(name) \
  UNION_ACCESSOR(wasmtime_val, i32, int32_t) \
  UNION_ACCESSOR(wasmtime_val, i64, int64_t) \
  UNION_ACCESSOR(wasmtime_val, f32, float) \
  UNION_ACCESSOR(wasmtime_val, f64, double) \
  UNION_ACCESSOR(wasmtime_val, externref, wasmtime_externref_t) \
  UNION_ACCESSOR(wasmtime_val, funcref, wasmtime_func_t) \
  \
  UNION_ACCESSOR(wasmtime_extern, func, wasmtime_func_t) \
  UNION_ACCESSOR(wasmtime_extern, memory, wasmtime_memory_t) \
  UNION_ACCESSOR(wasmtime_extern, table, wasmtime_table_t) \
  UNION_ACCESSOR(wasmtime_extern, global, wasmtime_global_t)

#define UNION_ACCESSOR(name, field, ty) \
  ty go_##name##_##field##_get(const name##_t *val); \
  void go_##name##_##field##_set(name##_t *val, ty i);

EACH_UNION_ACCESSOR(UNION_ACCESSOR)

#undef UNION_ACCESSOR
