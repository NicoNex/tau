#include "object.h"
#include "../vm/gc.h"

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

	return (struct object) {
		.data.gs = gs,
		.type = obj_getsetter,
	};
}
