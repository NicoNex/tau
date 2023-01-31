#include <stdio.h>
#include <stdlib.h>
#include "obj.h"

static void dispose_integer_obj(struct object *o) {
	free(o);
}

static void print_integer_obj(struct object *o) {
	int64_t i = o->data.i;
#ifdef __APPLE__
	printf("%lld\n", i);
#else
	printf("%ld\n", i);
#endif
}

struct object *new_integer_obj(int64_t val) {
	struct object *o = malloc(sizeof(struct object));
	o->data.i = val;
	o->type = obj_integer;
	o->dispose = dispose_integer_obj;
	o->print = print_integer_obj;

	return o;
}
