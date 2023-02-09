#include <stdio.h>
#include <stdlib.h>
#include "obj.h"

static void dispose_string_obj(struct object *o) {
	free(o->data.str);
	free(o);
}

static void print_string_obj(struct object *o) {
	puts(o->data.str);
}

struct object *new_string_obj(char *str, size_t len) {
	struct object *o = malloc(sizeof(struct object));
	o->data.str = str;
	o->len = len;
	o->type = obj_string;
	o->dispose = dispose_string_obj;
	o->print = print_string_obj;

	return o;
}
