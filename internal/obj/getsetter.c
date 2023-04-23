#include <stdlib.h>
#include "object.h"

void dispose_getsetter_obj(struct object o) {
	free(o.data.gs);
}

char *getsetter_str(struct object o) {
	struct getsetter *gs = o.data.gs;
	return object_str(gs->get(gs));
}

struct object new_getsetter_obj(struct object l, struct object r, getfn get, setfn set) {
	struct getsetter *gs = malloc(sizeof(struct getsetter));
	gs->l = l;
	gs->r = r;
	gs->get = get;
	gs->set = set;

	// We shouldn't need the marked field here since the getsetter is freed
	// as soon as it's unwrapped.
	return (struct object) {
		.data.gs = gs,
		.type = obj_getsetter,
		.marked = NULL,
	};
}
