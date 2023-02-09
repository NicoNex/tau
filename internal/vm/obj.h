#ifndef OBJ_H_
#define OBJ_H_

#include <stdint.h>
#include <stddef.h>

enum obj_type {
	obj_null,
	obj_error,
	obj_integer,
	obj_float,
	obj_boolean,
	obj_string,
	obj_bytes,
	obj_object,
	obj_function,
	obj_closure,
	obj_builtin,
	obj_list,
	obj_map,
	obj_pipe,
	obj_getsetter,
};

struct function {
	uint8_t *instructions;
	size_t len;
	int num_locals;
	int num_params;
};

typedef struct object object;

struct closure {
	struct function *fn;
	struct object **free;
	size_t num_free;
};

union data {
	int64_t i;
	double f;
	char *str;
	struct object **list;
	struct function *fn;
	struct closure *cl;
};

struct object {
	union data data;
	enum obj_type type;
	size_t len;
	void (*dispose)(struct object *o);
	void (*print)(struct object *o);
};

struct object *new_function_obj(uint8_t *insts, size_t len, int num_locals, int num_params);
struct object *new_closure_obj(struct function *fn, struct object **free, size_t num_free);
struct object *new_boolean_obj(int b);
struct object *new_integer_obj(int64_t val);
struct object *new_error_obj(char *msg, size_t len);
struct object *new_string_obj(char *str, size_t len);
struct object *new_float_obj(double val);
struct object *parse_bool(int b);
char *otype_str(enum obj_type t);

void print_obj(struct object *o);
void print_boolean_obj(struct object *o);

extern struct object *true_obj;
extern struct object *false_obj;
extern struct object *null_obj;

#endif
