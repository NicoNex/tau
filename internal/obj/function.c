#include <stdlib.h>
#include <stdio.h>
#include <string.h>
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
}

char *function_str(struct object o) {
	char *str = calloc(35, sizeof(char));
	sprintf(str, "closure[%p]", o.data.fn);

	return str;
}

// The instructions and the bookmarks are copied rather than pointed at: they
// come from the compiler, which is written in Go, and a Go allocation lives
// only as long as Go can see a reference to it. A pointer parked inside a C
// struct is not one, so the bytecode of every function was being freed under
// the VM as soon as the garbage collector of the other language got around
// to it, and the interpreter went on executing whatever landed there next.
//
// The line of a bookmark is already a C string, so copying the array keeps
// it as it was.
inline struct function *new_function(uint8_t *insts, size_t len, uint32_t num_locals, uint32_t num_params, struct bookmark *bmarks, uint32_t bklen) {
	struct function *fn = malloc(sizeof(struct function));

	fn->instructions = malloc(len > 0 ? len : 1);
	if (len > 0) {
		memcpy(fn->instructions, insts, len);
	}

	fn->bookmarks = NULL;
	if (bklen > 0) {
		fn->bookmarks = malloc(sizeof(struct bookmark) * bklen);
		memcpy(fn->bookmarks, bmarks, sizeof(struct bookmark) * bklen);
	}

	fn->len = len;
	fn->num_locals = num_locals;
	fn->num_params = num_params;
	fn->bklen = bklen;

	return fn;
}

inline struct object new_function_obj(uint8_t *insts, size_t len, uint32_t num_locals, uint32_t num_params, struct bookmark *bmarks, uint32_t bklen) {
	return (struct object) {
		.data.fn = new_function(insts, len, num_locals, num_params, bmarks, bklen),
		.type = obj_function,
		.gc = gc_header_alloc(),
	};
}
