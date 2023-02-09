#include <stdio.h>
#include <stdlib.h>
#include "obj.h"

static void dispose_error_obj(struct object *o) {
	free(o->data.str);
	free(o);
}

static void print_error_obj(struct object *o) {
	puts(o->data.str);
}

struct object *new_error_obj(char *str, size_t len) {
	struct object *o = malloc(sizeof(struct object));
	o->data.str = str;
	o->len = len;
	o->type = obj_error;
	o->dispose = dispose_error_obj;
	o->print = print_error_obj;

	return o;
}
