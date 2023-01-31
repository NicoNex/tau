#include <stdio.h>
#include <stdlib.h>
#include "obj.h"

static void dispose_boolean_obj(struct object *o) {
	free(o);
}

void print_boolean_obj(struct object *o) {
	puts(o->data.i == 1 ? "true" : "false");
}

struct object *parse_bool(int b) {
	return b ? true_obj : false_obj;
}

struct object *new_boolean_obj(int b) {
	struct object *o = calloc(1, sizeof(struct object));
	o->data.i = b != 0;
	o->type = obj_boolean;
	o->dispose = dispose_boolean_obj;
	o->print = print_boolean_obj;

	return o;
}
