#include <stdio.h>
#include <stdlib.h>
#include "obj.h"

static void dispose_float_obj(struct object *o) {
	free(o);
}

static void print_float_obj(struct object *o) {
	double f = o->data.f;
	printf("%f\n", f);
}

struct object *new_float_obj(double val) {
	struct object *o = malloc(sizeof(struct object));
	o->data.f = val;
	o->type = obj_float;
	o->dispose = dispose_float_obj;
	o->print = print_float_obj;

	return o;
}
