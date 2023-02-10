#include <stdlib.h>
#include <stdio.h>
#include "obj.h"

static void dummy_dispose(struct object *o) {}

/* ============================= BOOLEAN OBJECT ============================= */
void print_boolean_obj(struct object *o) {
	puts(o->data.i == 1 ? "true" : "false");
}

struct object parse_bool(int b) {
	return b ? true_obj : false_obj;
}

/* ============================= CLOSURE OBJECT ============================= */
// TODO: eventually dispose the function too if it's the case.
static void dispose_closure_obj(struct object *o) {
	free(o->data.cl);
}

static void print_closure_obj(struct object *o) {
	printf("closure[%p]\n", o->data.cl);
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
		.print = print_closure_obj
	};
}

/* ============================= ERROR OBJECT ============================= */
static void dispose_error_obj(struct object *o) {
	free(o->data.str);
}

static void print_error_obj(struct object *o) {
	puts(o->data.str);
}

struct object new_error_obj(char *str, size_t len) {
	return (struct object) {
		.data.str = str,
		.len = len,
		.type = obj_error,
		.dispose = dispose_error_obj,
		.print = print_error_obj
	};
}

/* ============================= FLOAT OBJECT ============================= */
static void print_float_obj(struct object *o) {
	double f = o->data.f;
	printf("%f\n", f);
}

struct object new_float_obj(double val) {
	return (struct object) {
		.data.f = val,
		.type = obj_float,
		.dispose = dummy_dispose,
		.print = print_float_obj
	};
}

/* ============================= FUNCTION OBJECT ============================= */
static void dispose_function_obj(struct object *o) {
	free(o->data.fn);
}

static void print_function_obj(struct object *o) {
	printf("closure[%p]\n", o->data.fn);
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
		.print = print_function_obj
	};
}

/* ============================= INTEGER OBJECT ============================= */
static void print_integer_obj(struct object *o) {
	int64_t i = o->data.i;
#ifdef __APPLE__
	printf("%lld\n", i);
#else
	printf("%ld\n", i);
#endif
}

struct object new_integer_obj(int64_t val) {
	return (struct object) {
		.data.i = val,
		.type = obj_integer,
		.dispose = dummy_dispose,
		.print = print_integer_obj
	};
}

/* ============================= STRING OBJECT ============================= */
static void dispose_string_obj(struct object *o) {
	free(o->data.str);
}

static void print_string_obj(struct object *o) {
	puts(o->data.str);
}

struct object new_string_obj(char *str, size_t len) {
	return (struct object) {
		.data.str = str,
		.len = len,
		.type = obj_string,
		.dispose = dispose_string_obj,
		.print = print_string_obj
	};
}

/* ============================= STATIC OBJECTS ============================= */
static void print_null_obj(struct object *o) {
	puts("null");
}

struct object true_obj = (struct object) {
	.data.i = 1,
	.type = obj_boolean,
	.len = 0,
	.dispose = dummy_dispose,
	.print = print_boolean_obj
};

struct object false_obj = (struct object) {
	.data.i = 0,
	.type = obj_boolean,
	.len = 0,
	.dispose = dummy_dispose,
	.print = print_boolean_obj
};

struct object null_obj = (struct object) {
	.data.i = 0,
	.type = obj_null,
	.len = 0,
	.dispose = dummy_dispose,
	.print = print_null_obj
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

void print_obj(struct object o) {
	o.print(&o);
}
