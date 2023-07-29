#include <stdio.h>
#include "object.h"
#include "../vm/gc.h"

char *integer_str(struct object o) {
	char *str = calloc(30, sizeof(char));

#ifdef __unix__
	sprintf(str, "%ld", o.data.i);
#else
	sprintf(str, "%lld", o.data.i);
#endif

	return str;
}

struct object new_integer_obj(int64_t val) {
	return (struct object) {
		.data.i = val,
		.type = obj_integer,
	};
}
