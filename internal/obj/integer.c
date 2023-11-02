#include <stdlib.h>
#include <stdio.h>
#include "object.h"

char *integer_str(struct object o) {
	char *str = calloc(64, sizeof(char));
	sprintf(str, "%lld", o.data.i);

	return str;
}

struct object new_integer_obj(int64_t val) {
	return (struct object) {
		.data.i = val,
		.type = obj_integer,
		.marked = NULL,
	};
}
