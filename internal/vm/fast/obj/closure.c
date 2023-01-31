#include <stdio.h>
#include <stdlib.h>
#include "obj.h"

// TODO: eventually dispose the function too if it's the case.
static void dispose_closure_obj(struct object *o) {
	free(o->data.cl);
	free(o);
}

static void print_closure_obj(struct object *o) {
	printf("closure[%p]\n", o->data.cl);
}

struct object *new_closure_obj(struct function *fn, struct object **free, size_t num_free) {
	struct closure *cl = malloc(sizeof(struct closure));
	cl->fn = fn;
	cl->free = free;
	cl->num_free = num_free;

	struct object *obj = malloc(sizeof(struct object));
	obj->data.cl = cl;
	obj->type = obj_closure;
	obj->dispose = dispose_closure_obj;
	obj->print = print_closure_obj;

	return obj;
}
