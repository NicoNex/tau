#include <stdio.h>
#include "object.h"
#include "../vm/gc.h"

char *function_str(struct object o) {
	char *str = GC_CALLOC(35, sizeof(char));
	sprintf(str, "closure[%p]", o.data.fn);

	return str;
}

inline struct function *new_function(uint8_t *insts, size_t len, uint32_t num_locals, uint32_t num_params, struct bookmark *bmarks, uint32_t bklen) {
	struct function *fn = GC_MALLOC(sizeof(struct function));
	fn->instructions = insts;
	fn->len = len;
	fn->num_locals = num_locals;
	fn->num_params = num_params;
	fn->bookmarks = bmarks;
	fn->bklen = bklen;

	return fn;
}

inline struct object new_function_obj(uint8_t *insts, size_t len, uint32_t num_locals, uint32_t num_params, struct bookmark *bmarks, uint32_t bklen) {
	return (struct object) {
		.data.fn = new_function(insts, len, num_locals, num_params, bmarks, bklen),
		.type = obj_function,
	};
}
