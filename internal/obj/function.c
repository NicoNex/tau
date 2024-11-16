#include <stdlib.h>
#include <stdio.h>
#include "object.h"

inline void dispose_function_data(struct function *fn) {
	for (int i = 0; i < fn->bklen; i++) {
		free(fn->bookmarks[i].line);
	}
	free(fn->bookmarks);
	free(fn->instructions);
	free(fn);
}

void dispose_function_obj(struct object o) {
	dispose_function_data(o.data.fn);
	free(o.gcdata);
}

char *function_str(struct object o) {
	char *str = calloc(35, sizeof(char));
	sprintf(str, "closure[%p]", o.data.fn);

	return str;
}

inline struct function *new_function(uint8_t *insts, size_t len, uint32_t num_locals, uint32_t num_params, struct bookmark *bmarks, uint32_t bklen) {
	struct function *fn = malloc(sizeof(struct function));
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
		.gcdata = new_gcdata(),
	};
}
