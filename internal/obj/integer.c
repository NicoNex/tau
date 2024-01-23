#include <stdio.h>
#include <inttypes.h>
#include "object.h"
#include "../vm/gc.h"

char *integer_str(struct object o) {
	char *str = calloc(64, sizeof(char));
	sprintf(str, "%" PRId64, o.data.i);

	return str;
}

struct object new_integer_obj(int64_t val) {
	return (struct object) {
		.data.i = val,
		.type = obj_integer,
	};
}
