#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "object.h"

// What a function holds, whoever holds the function itself.
static void dispose_function_parts(struct function *fn) {
	for (int i = 0; i < fn->bklen; i++) {
		free(fn->bookmarks[i].line);
	}
	free(fn->bookmarks);
	free(fn->instructions);
}

inline void dispose_function_data(struct function *fn) {
	dispose_function_parts(fn);
	free(fn);
}

// The struct itself is the payload of the header, so only the parts go.
void dispose_function_obj(struct object o) {
	dispose_function_parts(o.data.fn);
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
static void function_init(struct function *fn, uint8_t *insts, size_t len, uint32_t num_locals, uint32_t num_params, struct bookmark *bmarks, uint32_t bklen) {
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
}

// A function nobody collects: the one a VM runs, which outlives every object.
inline struct function *new_function(uint8_t *insts, size_t len, uint32_t num_locals, uint32_t num_params, struct bookmark *bmarks, uint32_t bklen) {
	struct function *fn = malloc(sizeof(struct function));

	function_init(fn, insts, len, num_locals, num_params, bmarks, bklen);
	return fn;
}

inline struct object new_function_obj(uint8_t *insts, size_t len, uint32_t num_locals, uint32_t num_params, struct bookmark *bmarks, uint32_t bklen) {
	struct gc_header *h = gc_alloc(sizeof(struct function));
	struct function *fn = GC_PAYLOAD(h);

	function_init(fn, insts, len, num_locals, num_params, bmarks, bklen);
	h->obj = (struct object) {
		.data.fn = fn,
		.type = obj_function,
	};
	return h->obj;
}
