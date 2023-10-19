#pragma once

#include <stdint.h>
#include <stddef.h>
#include <threads.h>
#include "../tauerr/bookmark.h"

#define NUM_BUILTINS 26

enum obj_type {
	obj_null,
	obj_boolean,
	obj_integer,
	obj_float,
	obj_builtin,
	obj_string,
	obj_error,
	obj_list,
	obj_map,
	obj_function,
	obj_closure,
	obj_object,
	obj_pipe,
	obj_bytes,
	obj_getsetter,
	obj_native
};

struct function {
	uint8_t *instructions;
	size_t len;
	uint32_t num_locals;
	uint32_t num_params;
	uint32_t bklen;
	struct bookmark *bookmarks;
};

struct closure {
	struct function *fn;
	struct object *free;
	size_t num_free;
};

struct map {
	struct map_node *root;
	size_t len;
};

struct list {
	struct object *list;
	size_t len;
	size_t cap;
};

struct string {
	char *str;
	size_t len;
};

struct pipe {
	struct object *buf;
	size_t cap;
	size_t len;
	uint32_t head;
	uint32_t tail;
	uint32_t is_buffered;
	uint32_t is_closed;
	mtx_t mu;
	cnd_t not_empty;
	cnd_t not_full;
};

union data {
	int64_t i;
	double f;
	void *handle;
	struct function *fn;
	struct closure *cl;
	struct string *str;
	struct list *list;
	struct map *map;
	struct object_node **obj;
	struct pipe *pipe;
	struct getsetter *gs;
	struct object (*builtin)(struct object *args, size_t len);
};

struct object {
	union data data;
	enum obj_type type;
};

struct getsetter {
	struct object l;
	struct object r;
	struct object (*get)(struct getsetter *gs);
	struct object (*set)(struct getsetter *gs, struct object o);
};

typedef struct object (*getfn)(struct getsetter *gs);
typedef struct object (*setfn)(struct getsetter *gs, struct object o);

struct key_hash {
	enum obj_type type;
	uint64_t val;
};

struct map_pair {
	struct object key;
	struct object val;
};

struct map_node {
	struct key_hash key;
	struct map_pair val;
	struct map_node *l;
	struct map_node *r;
};

struct object_node {
	char *name;
	uint64_t key;
	struct object val;
	struct object_node *l;
	struct object_node *r;
};

// Static objects.
extern struct object null_obj;
extern struct object true_obj;
extern struct object false_obj;

// Boolean object.
struct object new_boolean_obj(uint32_t b);
struct object parse_bool(uint32_t b);
char *boolean_str(struct object o);

// Integer object.
struct object new_integer_obj(int64_t val);
char *integer_str(struct object o);

// Float object.
struct object new_float_obj(double val);
char *float_str(struct object o);

// String object.
struct object new_string_obj(char *str, size_t len);
char *string_str(struct object o);
void mark_string_obj(struct object s);

// Error object.
struct object new_error_obj(char *msg, size_t len);
char *error_str(struct object o);

// List object.
struct list new_list(size_t cap);
void list_insert(struct list *l, struct object o, size_t idx);
struct list list_copy(struct list l);
struct object new_list_obj(struct object *list, size_t len);
struct object list_getsetter_get(struct getsetter *gs);
struct object list_getsetter_set(struct getsetter *gs, struct object val);
char *list_str(struct object o);
void mark_list_obj(struct object l);

// Pipe object.
struct object new_pipe();
struct object new_buffered_pipe(size_t size);
int pipe_send(struct object pipe, struct object o);
struct object pipe_recv(struct object pipe);
int pipe_close(struct object pipe);
void mark_pipe_obj(struct object pipe);

// Map object.
struct object new_map();
struct map_pair map_get(struct object map, struct object k);
struct map_pair map_set(struct object map, struct object k, struct object v);
struct object map_getsetter_get(struct getsetter *gs);
struct object map_getsetter_set(struct getsetter *gs, struct object val);
void mark_map_obj(struct object m);
char *map_str(struct object map);

// Object object.
struct object new_object();
struct object object_get(struct object obj, char *name);
struct object object_set(struct object obj, char *name, struct object val);
struct object object_getsetter_get(struct getsetter *gs);
struct object object_getsetter_set(struct getsetter *gs, struct object val);
struct object object_to_module(struct object o);
void mark_object_obj(struct object o);
char *object_obj_str(struct object obj);

// Function object.
struct object new_function_obj(uint8_t *insts, size_t len, uint32_t num_locals, uint32_t num_params, struct bookmark *bmarks, uint32_t num_bookmarks);
char *function_str(struct object o);

// Closure object.
struct object new_closure_obj(struct function *fn, struct object *free, size_t num_free);
char *closure_str(struct object o);
void mark_closure_obj(struct object c);

// Builtin object.
typedef struct object (*builtin)(struct object *args, size_t len);
extern const builtin builtins[NUM_BUILTINS];
struct object new_builtin_obj(struct object (*builtin)(struct object *args, size_t len));

// Getsetter object.
struct object new_getsetter_obj(struct object l, struct object r, getfn get, setfn set);
char *getsetter_str(struct object o);

// Native object.
struct object native_getsetter_get(struct getsetter *gs);
struct object native_getsetter_set(struct getsetter *gs, struct object val);

// Util functions.
char *otype_str(enum obj_type t);
char *object_str(struct object o);
void print_obj(struct object o);
struct object errorf(char *fmt, ...);
uint64_t fnv64a(char *s);
