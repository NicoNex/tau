#pragma once

#include <stdlib.h>
#include <stdint.h>
#include <stddef.h>
#include "../vm/thrd.h"
#include "../tauerr/bookmark.h"

// Layout of gc_header.mark:
//
//   bit  0    GC_MARK: reachable, set by the mark phase, cleared by the sweep.
//   bit  1    GC_TRACKED: already in the heap, prevents adding it twice.
//   bits 2..  epoch of the last visit of the mark phase.
//
// The mark bit alone can't say whether an object was already traversed: a
// slice marks the header of its owner before the owner itself is visited.
// The epoch is what stops the traversal of cycles, and unlike a "visited"
// bit it needs no cleanup for the objects that aren't in the heap.
#define GC_MARK        1
#define GC_TRACKED     2
#define GC_EPOCH_SHIFT 2

extern uint32_t gc_epoch;

struct gc_header;

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
	obj_native,
	// A native function with a declared signature. It is added at the end so
	// that the values of the ones before it, which the bytecode carries, stay
	// where they were.
	obj_native_fn
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
	// Set when this list is a slice: the header of the object that owns the
	// underlying array, kept alive as long as the slice is.
	struct gc_header *owner;
};

struct string {
	char *str;
	size_t len;
	struct gc_header *owner;
};

struct bytes {
	uint8_t *bytes;
	size_t len;
	struct gc_header *owner;
};

struct pipe {
	struct object *buf;
	size_t cap;
	size_t len;
	uint32_t head;
	uint32_t tail;
	uint32_t is_buffered;
	uint32_t is_closed;
	// How many values went in and how many came out since the pipe was made.
	// An unbuffered send waits for the count of received values to pass the
	// one it got when it put its own value in: that, and not the room left in
	// the buffer, is what tells a sender its value has been taken.
	uint64_t sent;
	uint64_t recvd;
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
	struct bytes *bytes;
	struct list *list;
	struct map *map;
	struct object_node **obj;
	struct pipe *pipe;
	struct object (*builtin)(struct object *args, size_t len);
};

struct object {
	union data data;
	enum obj_type type;
};

// Every collectable object owns one of these: it holds the state the
// collector keeps about the object, doubles as its node in the heap, and is
// allocated in one block with what `data` points at, which follows it.
//
// That last part is why an object carries no pointer to its header: the
// header sits at a known offset before the payload, so `data` finds it. An
// object is two words instead of three, and it costs one allocation instead
// of two.
struct gc_header {
	uint32_t mark;
	// Bytes of payload after the header, needed to put the block back in the
	// right free list.
	uint32_t size;
	struct gc_header *next;
	struct object obj;
};

// The payload of a header, and the header of a payload.
#define GC_PAYLOAD(h) ((void *) ((struct gc_header *) (h) + 1))
#define GC_OF(p)      ((struct gc_header *) (p) - 1)

// A header and the `size` bytes of payload after it, in one allocation.
struct gc_header *gc_alloc(size_t size);

// The header of an object, NULL for the ones the collector doesn't look
// after. Those are the values with no payload of their own (integers,
// booleans, builtins) and the native handles, which point at memory this
// program didn't allocate and doesn't free.
static inline struct gc_header *obj_gc(struct object o) {
	if (o.type <= obj_builtin || o.type == obj_native) {
		return NULL;
	}
	return GC_OF(o.data.handle);
}

struct key_hash {
	uint64_t type;
	uint64_t val;
} __attribute__((packed));

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
struct object new_string_slice(char *str, size_t len, struct gc_header *owner);
char *string_str(struct object o);

// A tau string is a pointer and a length, never a C string: a slice points
// into the middle of somebody else's buffer and has no terminator of its own,
// so reading one as a char * runs into whatever follows. Everything that has
// to hand a string to C goes through cstr, which copies only when it must,
// and gives back what it made through cstr_free.
char *cstr(struct string *s);
void cstr_free(struct string *s, char *c);
void mark_string_obj(struct object s);
void dispose_string_obj(struct object o);

// Bytes object.
struct object new_bytes_obj(uint8_t *bytes, size_t len);
struct object new_bytes_slice(uint8_t *bytes, size_t len, struct gc_header *owner);
char *bytes_str(struct object o);
void mark_bytes_obj(struct object o);
void dispose_bytes_obj(struct object o);

// Error object.
struct object new_error_obj(char *msg, size_t len);
struct object errorf(char *fmt, ...);
char *error_str(struct object o);
void dispose_error_obj(struct object o);

// List object.
struct object make_list(size_t cap);
struct object new_list_obj(struct object *list, size_t len);
struct object new_list_obj_data(struct object *list, size_t len, size_t cap);
struct object new_list_slice(struct object *list, size_t len, struct gc_header *owner);
char *list_str(struct object o);
void mark_list_obj(struct object l);
void dispose_list_obj(struct object o);
struct list list_copy(struct list l);

// Pipe object.
struct object new_pipe();
struct object new_buffered_pipe(size_t size);
int pipe_send(struct object pipe, struct object o);
struct object pipe_recv(struct object pipe);
int pipe_close(struct object pipe);
void mark_pipe_obj(struct object pipe);
void dispose_pipe_obj(struct object pipe);

// Map object.
struct object new_map();
struct map_pair map_get(struct object map, struct object k);
struct map_pair map_set(struct object map, struct object k, struct object v);
void mark_map_obj(struct object m);
char *map_str(struct object map);
void map_delete(struct object map, struct object key);
void dispose_map_obj(struct object map);
struct object map_keys(struct object map);

// Object object.
struct object new_object();
struct object object_get(struct object obj, char *name);
struct object object_set(struct object obj, char *name, struct object val);
struct object object_to_module(struct object o);
struct object object_keys(struct object o);
void mark_object_obj(struct object o);
char *object_obj_str(struct object obj);
void dispose_object_obj(struct object obj);

// Function object.
struct function *new_function(
	uint8_t *insts,
	size_t len,
	uint32_t num_locals,
	uint32_t num_params,
	struct bookmark *bmarks,
	uint32_t num_bookmarks
);
struct object new_function_obj(
	uint8_t *insts,
	size_t len,
	uint32_t num_locals,
	uint32_t num_params,
	struct bookmark *bmarks,
	uint32_t num_bookmarks
);
char *function_str(struct object o);
void dispose_function_obj(struct object o);
void dispose_function_data(struct function *fn);

// Closure object.
struct object new_closure_obj(struct function *fn, struct object *free, size_t num_free);
char *closure_str(struct object o);
void dispose_closure_obj(struct object o);
void mark_closure_obj(struct object c);

// Native function object: a C function plus the signature that says how to
// talk to it, see ffi.c.
struct object new_native_obj(void *fn, int64_t ret, const int64_t *args, size_t nargs);
struct object new_cexport_obj(struct object fn, int64_t ret, const int64_t *args, size_t nargs);
struct object native_call(struct object f, struct object *args, size_t nargs);
void dispose_native_obj(struct object o);
char *native_str(struct object o);

// Builtin object.
typedef struct object (*builtin)(struct object *args, size_t len);
extern const builtin builtins[];
struct object new_builtin_obj(struct object (*builtin)(struct object *args, size_t len));

// Util functions.
char *otype_str(enum obj_type t);
char *object_str(struct object o);
void print_obj(struct object o);
void mark_obj(struct object o);
// The owner of the buffer a slice points into. It goes through mark_obj and
// not through the mark bit alone, so that an owner which is itself a slice
// marks the owner underneath it. Slicing flattens the chain and there is
// normally nothing under the first owner: this is what makes forgetting to
// flatten cost a hop instead of the buffer.
void mark_owner(struct gc_header *h);
void free_obj(struct object o);

// The header to hand to a slice cut out of `o`, which holds the buffer alive
// for as long as the slice is around. A slice of a slice takes the owner of
// the buffer and not the slice it was cut from, so that the chain is one hop
// deep whatever it is sliced out of.
static inline struct gc_header *slice_owner(struct object o) {
	switch (o.type) {
	case obj_string:
		return o.data.str->owner != NULL ? o.data.str->owner : obj_gc(o);
	case obj_bytes:
		return o.data.bytes->owner != NULL ? o.data.bytes->owner : obj_gc(o);
	case obj_list:
		return o.data.list->owner != NULL ? o.data.list->owner : obj_gc(o);
	default:
		return NULL;
	}
}

// Garbage collector hooks, implemented in ../vm/heap.c.
// Park before blocking so the collector doesn't wait for this thread.
void gc_park(void);
void gc_unpark(void);

// The VM of this thread and a call into it, for a C function that calls back
// into tau. Both live in internal/vm; declared here because the FFI is the
// only thing in obj that needs them.
struct vm;
struct vm *gc_current_vm(void);
struct object vm_call_tau(struct vm *vm, struct object cl, struct object *args, size_t nargs);
void *gc_add_roots(struct object *objs, size_t len);
uint64_t fnv64a(const char *s, size_t len);
uint32_t is_truthy(struct object * restrict o);
