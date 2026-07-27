#include <stdlib.h>
#include <stdio.h>
#include "object.h"

void dispose_closure_obj(struct object o) {
	// The function belongs to the constants pool and is shared by every
	// closure built from it, only the closure itself is freed here.
	free(o.data.cl->free);
	free(o.data.cl);
}

char *closure_str(struct object o) {
	char *str = calloc(35, sizeof(char));
	sprintf(str, "closure[%p]", o.data.cl->fn);

	return str;
}

void mark_closure_obj(struct object c) {
	c.gc->mark |= GC_MARK;
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
		.gc = gc_header_alloc(),
	};
}
