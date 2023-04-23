#include <stdlib.h>
#include <stdio.h>
#include "object.h"

void dispose_closure_obj(struct object o) {
	dispose_function_data(o.data.cl->fn);
	free(o.marked);
	free(o.data.cl->free);
	free(o.data.cl);
}

char *closure_str(struct object o) {
	char *str = calloc(35, sizeof(char));
	sprintf(str, "closure[%p]", o.data.cl->fn);

	return str;
}

void mark_closure_obj(struct object c) {
	*c.marked = 1;
	for (uint32_t i = 0; i < c.data.cl->num_free; i++) {
		mark_obj(c.data.cl->free[i]);
	}
}

struct object new_closure_obj(struct function *fn, struct object *free, size_t num_free) {
	struct closure *cl = malloc(sizeof(struct closure));
	cl->fn = fn;
	cl->free = free;
	cl->num_free = num_free;

	return (struct object) {
		.data.cl = cl,
		.type = obj_closure,
		.marked = MARKPTR(),
	};
}
