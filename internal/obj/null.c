#include "object.h"

struct object null_obj = (struct object) {
	.data.i = 0,
	.type = obj_null,
};
