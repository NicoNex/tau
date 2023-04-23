#include <stdlib.h>
#include <stdio.h>
#include "object.h"

char *float_str(struct object o) {
	char *str = calloc(35, sizeof(char));
	sprintf(str, "%f", o.data.f);

	return str;
}

struct object new_float_obj(double val) {
	return (struct object) {
		.data.f = val,
		.type = obj_float,
		.marked = NULL,
	};
}
