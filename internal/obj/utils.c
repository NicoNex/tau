#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "object.h"
#include "plugin.h"

// Weak fallbacks for the collector hooks: the real ones live in ../vm/heap.c
// and take over whenever the VM is linked in. Without them a program that
// only links the obj package (the compiler tests) wouldn't build.
__attribute__((weak)) uint32_t gc_epoch = 1;
__attribute__((weak)) void gc_park(void) {}
__attribute__((weak)) void gc_unpark(void) {}

__attribute__((weak)) struct gc_header *gc_header_alloc(void) {
	struct gc_header *h = malloc(sizeof(struct gc_header));
	h->mark = 0;
	return h;
}

// The same for the three the FFI needs to call back into tau. Without a VM
// there is nobody to call and nothing to keep alive, which is exactly the
// truth in a program that links only this package.
__attribute__((weak)) struct vm *gc_current_vm(void) { return NULL; }
__attribute__((weak)) void *gc_add_roots(struct object *objs, size_t len) {
	(void) objs;
	(void) len;
	return NULL;
}

__attribute__((weak)) struct object vm_call_tau(struct vm *vm, struct object cl, struct object *args, size_t nargs) {
	(void) vm;
	(void) cl;
	(void) args;
	(void) nargs;
	return errorf("callback: there is no virtual machine here");
}

char *otype_str(enum obj_type t) {
	static char *strings[] = {
		"null",
		"bool",
		"int",
		"float",
		"builtin",
		"string",
		"error",
		"list",
		"map",
		"function",
		"closure",
		"object",
		"pipe",
		"bytes",
		"native",
		"native"
	};
	return t <= obj_native_fn ? strings[t] : "corrupted";
}

char *object_str(struct object o) {
	switch (o.type) {
	case obj_null:
		return strdup("null");
	case obj_boolean:
		return boolean_str(o);
	case obj_integer:
		return integer_str(o);
	case obj_float:
		return float_str(o);
	case obj_builtin:
		return strdup("<builtin function>");
	case obj_string:
		return string_str(o);
	case obj_error:
		return error_str(o);
	case obj_list:
		return list_str(o);
	case obj_map:
		return map_str(o);
	case obj_function:
		return function_str(o);
	case obj_closure:
		return closure_str(o);
	case obj_object:
		return object_obj_str(o);
	case obj_pipe:
		return strdup("<pipe>");
	case obj_bytes:
		return bytes_str(o);
	case obj_native:
		return strdup("<native>");
	case obj_native_fn:
		return native_str(o);
	default:
		return strdup("<corrupted>");
	}
}

void print_obj(struct object o) {
	char *str = object_str(o);
	puts(str);
	free(str);
}

inline void mark_obj(struct object o) {
	// Objects already visited in this cycle are skipped, otherwise a cycle
	// (an object holding a closure that captured the object itself, a list
	// containing itself...) would recur forever.
	if (o.type > obj_builtin && o.gc != NULL && (o.gc->mark >> GC_EPOCH_SHIFT) != gc_epoch) {
		o.gc->mark = (o.gc->mark & (GC_MARK | GC_TRACKED)) | (gc_epoch << GC_EPOCH_SHIFT);

		switch (o.type) {
		case obj_object:
			mark_object_obj(o);
			break;
		case obj_list:
			mark_list_obj(o);
			break;
		case obj_closure:
			mark_closure_obj(o);
			break;
		case obj_map:
			mark_map_obj(o);
			break;
		case obj_string:
			mark_string_obj(o);
			break;
		case obj_bytes:
			mark_bytes_obj(o);
			break;
		case obj_pipe:
			mark_pipe_obj(o);
			break;
		default:
			o.gc->mark |= GC_MARK;
			break;
		}
	}
}

void mark_owner(struct gc_header *h) {
	if (h != NULL) {
		mark_obj(h->obj);
	}
}

void free_obj(struct object o) {
	switch (o.type) {
	case obj_string:
		dispose_string_obj(o);
		return;
	case obj_error:
		dispose_error_obj(o);
		return;
	case obj_list:
		dispose_list_obj(o);
		return;
	case obj_map:
		dispose_map_obj(o);
		return;
	case obj_function:
		dispose_function_obj(o);
		return;
	case obj_closure:
		dispose_closure_obj(o);
		return;
	case obj_object:
		dispose_object_obj(o);
		return;
	case obj_pipe:
		dispose_pipe_obj(o);
		return;
	case obj_bytes:
		dispose_bytes_obj(o);
		return;
	case obj_native:
		// The handle of a shared object, which is not closed here. A function
		// prepared from one of its symbols keeps only the address, so
		// unloading the library under it turns the next call into a jump into
		// memory that is no longer code - and that is what happened, silently
		// and whenever a collection landed in the wrong place. A library
		// stays mapped for as long as the program runs; whoever really wants
		// it gone can call dlclose, which is a C function like any other.
		return;
	case obj_native_fn:
		dispose_native_obj(o);
		return;
	default:
		return;
	}
}

inline uint32_t is_truthy(struct object * o) {
	switch (o->type) {
	case obj_boolean:
		return o->data.i == 1;
	case obj_integer:
		return o->data.i != 0;
	case obj_float:
		return o->data.f != 0;
	case obj_string:
		return o->data.str->len != 0;
	case obj_null:
		return 0;
	default:
		return 1;
	}
}
