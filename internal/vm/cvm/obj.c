#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "obj.h"

static void dummy_dispose(struct object o) {}

// ============================= CLOSURE OBJECT =============================
static void dispose_closure_obj(struct object o) {
	free(o.data.cl->fn->instructions);
	free(o.data.cl->fn);
	free(o.data.cl);
}

static char *closure_str(struct object o) {
	char *str = calloc(35, sizeof(char));
	sprintf(str, "closure[%p]", o.data.cl->fn);

	return str;
}

struct object new_closure_obj(struct function *fn, struct object *free, size_t num_free) {
	struct closure *cl = malloc(sizeof(struct closure));
	cl->fn = fn;
	cl->free = free;
	cl->num_free = num_free;

	return (struct object) {
		.data.cl = cl,
		.type = obj_closure,
		.dispose = dispose_closure_obj,
		.string = closure_str
	};
}

// ============================= ERROR OBJECT =============================
static void dispose_error_obj(struct object o) {
	free(o.data.str);
}

static char *error_str(struct object o) {
	char *str = calloc(o.len+1, sizeof(char));
	strncpy(str, o.data.str, o.len);

	return str;
}

struct object new_error_obj(char *str, size_t len) {
	return (struct object) {
		.data.str = str,
		.len = len,
		.type = obj_error,
		.dispose = dispose_error_obj,
		.string = error_str
	};
}

// ============================= FLOAT OBJECT =============================
static char *float_str(struct object o) {
	char *str = calloc(35, sizeof(char));
	sprintf(str, "%f", o.data.f);

	return str;
}

struct object new_float_obj(double val) {
	return (struct object) {
		.data.f = val,
		.type = obj_float,
		.dispose = dummy_dispose,
		.string = float_str
	};
}

// ============================= FUNCTION OBJECT =============================
static void dispose_function_obj(struct object o) {
	free(o.data.fn->instructions);
	free(o.data.fn);
}

static char *function_str(struct object o) {
	char *str = calloc(35, sizeof(char));
	sprintf(str, "closure[%p]", o.data.fn);

	return str;
}

struct object new_function_obj(uint8_t *insts, size_t len, int num_params, int num_locals) {
	struct function *fn = malloc(sizeof(struct function));
	fn->instructions = insts;
	fn->len = len;
	fn->num_locals = num_locals;
	fn->num_params = num_params;

	return (struct object) {
		.data.fn = fn,
		.type = obj_function,
		.dispose = dispose_function_obj,
		.string = function_str
	};
}

// ============================= INTEGER OBJECT =============================
static char *integer_str(struct object o) {
	char *str = calloc(30, sizeof(char));

#ifdef __unix__
	sprintf(str, "%ld", o.data.i);
#else
	sprintf(str, "%lld", o.data.i);
#endif

	return str;
}

struct object new_integer_obj(int64_t val) {
	return (struct object) {
		.data.i = val,
		.type = obj_integer,
		.dispose = dummy_dispose,
		.string = integer_str
	};
}

// ============================= STRING OBJECT =============================
static void dispose_string_obj(struct object o) {
	free(o.data.str);
}

static char *string_str(struct object o) {
	char *str = calloc(o.len+1, sizeof(char));
	strncpy(str, o.data.str, o.len);

	return str;
}

struct object new_string_obj(char *str, size_t len) {
	return (struct object) {
		.data.str = str,
		.len = len,
		.type = obj_string,
		.dispose = dispose_string_obj,
		.string = string_str
	};
}

// ============================= STATIC OBJECTS =============================
static char *boolean_str(struct object o) {
	char *str = calloc(6, sizeof(char));
	strcpy(str, o.data.i == 1 ? "true" : "false");

	return str;
}

struct object parse_bool(int b) {
	return b ? true_obj : false_obj;
}

static char *null_str(struct object o) {
	char *str = calloc(5, sizeof(char));
	strcpy(str, "null");

	return str;
}

struct object true_obj = (struct object) {
	.data.i = 1,
	.type = obj_boolean,
	.len = 0,
	.dispose = dummy_dispose,
	.string = boolean_str
};

struct object false_obj = (struct object) {
	.data.i = 0,
	.type = obj_boolean,
	.len = 0,
	.dispose = dummy_dispose,
	.string = boolean_str
};

struct object null_obj = (struct object) {
	.data.i = 0,
	.type = obj_null,
	.len = 0,
	.dispose = dummy_dispose,
	.string = null_str
};

char *otype_str(enum obj_type t) {
	static char *strings[] = {
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

char *object_str(struct object o) {
	return o.string(o);
}

void print_obj(struct object o) {
	char *str = o.string(o);
	puts(str);
	free(str);
}
