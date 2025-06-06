#include "shims.h"

#define UNION_ACCESSOR(name, field, ty) \
  ty go_##name##_##field##_get(const name##_t *val) { return val->of.field; } \
  void go_##name##_##field##_set(name##_t *val, ty i) { val->of.field = i; }

EACH_UNION_ACCESSOR(UNION_ACCESSOR)
