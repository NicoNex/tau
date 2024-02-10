#include <string.h>
#include "object.h"
#include "../vm/gc.h"

struct object true_obj = (struct object) {
	.data.i = 1,
	.type = obj_boolean,
};

struct object false_obj = (struct object) {
	.data.i = 0,
	.type = obj_boolean,
};

inline __attribute__((always_inline))
struct object parse_bool(uint32_t b) {
	return b ? true_obj : false_obj;
}

char *boolean_str(struct object o) {
	return o.data.i ? GC_STRDUP("true") : GC_STRDUP("false");
}
