#include <stdlib.h>
#include <stdio.h>
#include "obj.h"

/* ============================= BOOLEAN OBJECT ============================= */
static void dispose_boolean_obj(struct object *o) {
	free(o);
}

void print_boolean_obj(struct object *o) {
	puts(o->data.i == 1 ? "true" : "false");
}

struct object *parse_bool(int b) {
	return b ? true_obj : false_obj;
}

struct object *new_boolean_obj(int b) {
	struct object *o = calloc(1, sizeof(struct object));
	o->data.i = b != 0;
	o->type = obj_boolean;
	o->dispose = dispose_boolean_obj;
	o->print = print_boolean_obj;

	return o;
}

/* ============================= CLOSURE OBJECT ============================= */
// TODO: eventually dispose the function too if it's the case.
static void dispose_closure_obj(struct object *o) {
	free(o->data.cl);
	free(o);
}

static void print_closure_obj(struct object *o) {
	printf("closure[%p]\n", o->data.cl);
}

struct object *new_closure_obj(struct function *fn, struct object **free, size_t num_free) {
	struct closure *cl = malloc(sizeof(struct closure));
	cl->fn = fn;
	cl->free = free;
	cl->num_free = num_free;

	struct object *obj = malloc(sizeof(struct object));
	obj->data.cl = cl;
	obj->type = obj_closure;
	obj->dispose = dispose_closure_obj;
	obj->print = print_closure_obj;

	return obj;
}

/* ============================= ERROR OBJECT ============================= */
static void dispose_error_obj(struct object *o) {
	free(o->data.str);
	free(o);
}

static void print_error_obj(struct object *o) {
	puts(o->data.str);
}

struct object *new_error_obj(char *str, size_t len) {
	struct object *o = malloc(sizeof(struct object));
	o->data.str = str;
	o->len = len;
	o->type = obj_error;
	o->dispose = dispose_error_obj;
	o->print = print_error_obj;

	return o;
}

/* ============================= FLOAT OBJECT ============================= */
static void dispose_float_obj(struct object *o) {
	free(o);
}

static void print_float_obj(struct object *o) {
	double f = o->data.f;
	printf("%f\n", f);
}

struct object *new_float_obj(double val) {
	struct object *o = malloc(sizeof(struct object));
	o->data.f = val;
	o->type = obj_float;
	o->dispose = dispose_float_obj;
	o->print = print_float_obj;

	return o;
}

/* ============================= FUNCTION OBJECT ============================= */
static void dispose_function_obj(struct object *o) {
	free(o->data.fn);
	free(o);
}

static void print_function_obj(struct object *o) {
	printf("closure[%p]\n", o->data.fn);
}

struct object *new_function_obj(uint8_t *insts, size_t len, int num_params, int num_locals) {
	struct function *fn = malloc(sizeof(struct function));
	fn->instructions = insts;
	fn->len = len;
	fn->num_locals = num_locals;
	fn->num_params = num_params;

	struct object *o = calloc(1, sizeof(struct object));
	o->data.fn = fn;
	o->type = obj_function;
	o->dispose = dispose_function_obj;
	o->print = print_function_obj;

	return o;
}

/* ============================= INTEGER OBJECT ============================= */
static void dispose_integer_obj(struct object *o) {
	free(o);
}

static void print_integer_obj(struct object *o) {
	int64_t i = o->data.i;
#ifdef __APPLE__
	printf("%lld\n", i);
#else
	printf("%ld\n", i);
#endif
}

struct object *new_integer_obj(int64_t val) {
	struct object *o = malloc(sizeof(struct object));
	o->data.i = val;
	o->type = obj_integer;
	o->dispose = dispose_integer_obj;
	o->print = print_integer_obj;

	return o;
}

/* ============================= STRING OBJECT ============================= */
static void dispose_string_obj(struct object *o) {
	free(o->data.str);
	free(o);
}

static void print_string_obj(struct object *o) {
	puts(o->data.str);
}

struct object *new_string_obj(char *str, size_t len) {
	struct object *o = malloc(sizeof(struct object));
	o->data.str = str;
	o->len = len;
	o->type = obj_string;
	o->dispose = dispose_string_obj;
	o->print = print_string_obj;

	return o;
}

/* ============================= STATIC OBJECTS ============================= */
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

void print_obj(struct object *o) {
	o->print(o);
}
