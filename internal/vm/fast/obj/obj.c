#include <stdio.h>
#include "obj.h"

static void dummy_dispose(struct object *o) {}

static void print_null_obj(struct object *o) {
	puts("null");
}

object *true_obj = &(struct object) {
	.data.i = 1,
	.type = obj_boolean,
	.len = 0,
	.dispose = dummy_dispose,
	.print = print_boolean_obj
};

object *false_obj = &(struct object) {
	.data.i = 0,
	.type = obj_boolean,
	.len = 0,
	.dispose = dummy_dispose,
	.print = print_boolean_obj
};

object *null_obj = &(struct object) {
	.data.i = 0,
	.type = obj_null,
	.len = 0,
	.dispose = dummy_dispose,
	.print = print_null_obj
};

char *otype_str(enum obj_type t) {
	char *strings[] = {
		"null",
		"error",
		"integer",
		"float",
		"boolean",
		"string",
		"bytes",
		"object",
		"function",
		"closure",
		"builtin",
		"list",
		"map",
		"pipe",
		"class",
		"getsetter"
	};

	return strings[t];
}
