#include <stdarg.h>
#include <stdio.h>
#include <string.h>
#include <stddef.h>
#include <stdlib.h>
#include <setjmp.h>

#include "vm.h"
#include "opcode.h"
#include "thrd.h"
#ifdef TAU_RT
	// A runtime built to run one bundled program has no Go in it: the loader
	// of a module that came with the program is written in C, next to the
	// main that starts it.
	int vm_exec_load_module(struct vm *vm, char *path);
#else
	#include "_cgo_export.h"
#endif
#include "../obj/plugin.h"
#include "../obj/libffi/include/ffi.h"

#define read_uint8(b) ((b)[0])
#define read_uint16(b) (((b)[0] << 8) | (b)[1])
#define read_uint32(b) (((b)[0] << 24) | ((b)[1] << 16) | ((b)[2] << 8) | (b)[3])

#define vm_current_frame(vm) (&vm->frames[vm->frame_idx])
#define vm_push_frame(vm, frame) vm->frames[++vm->frame_idx] = frame
#define vm_pop_frame(vm) (&vm->frames[vm->frame_idx--])

#define vm_stack_push(vm, obj) vm->stack[vm->sp++] = (obj)
#define vm_stack_pop(vm) (vm->stack[--vm->sp])
#define vm_stack_pop_ignore(vm) vm->sp--
#define vm_stack_peek(vm) (vm->stack[vm->sp-1])

#ifndef GC_DEBUG
	#define vm_heap_add(vm, o) heap_add(o)
#else
	#define vm_heap_add(vm, o) printf("adding type %s to heap\n", otype_str(o.type)); heap_add(o)
#endif

#ifndef DEBUG
	#define DISPATCH() goto *jump_table[*frame->ip++]
#else
	#define DISPATCH() puts(opcode_str(*frame->ip)); goto *jump_table[*frame->ip++]
#endif

#define ASSERT(obj, t) ((obj)->type == t)
#define ASSERT2(obj, t1, t2) (ASSERT(obj, t1) || ASSERT(obj, t2))
#define ASSERT4(obj, t1, t2, t3, t4) (ASSERT(obj, t1) || ASSERT(obj, t2) || ASSERT(obj, t3) || ASSERT(obj, t4))
#define M_ASSERT(o1, o2, t) (ASSERT(o1, t) && ASSERT(o2, t))
#define M_ASSERT2(o1, o2, t1, t2) (ASSERT2(o1, t1, t2) && ASSERT2(o2, t1, t2))

static inline struct frame new_frame(struct object cl, uint32_t base_ptr) {
	return (struct frame) {
		.cl = cl,
		.base_ptr = base_ptr,
		.ip = cl.data.cl->fn->instructions,
		.start = cl.data.cl->fn->instructions
	};
}

inline struct state new_state() {
	return (struct state) {
		// The globals pool is shared with the tau routines, so it's allocated
		// once with its final size: a realloc would pull the list from under
		// the other threads.
		.globals = new_pool(GLOBAL_SIZE),
		.consts = new_pool(0),
		.mods = new_modtab(),
		.ndefs = 0
	};
}

inline void state_dispose(struct state s) {
	// The constants are not freed: they belong to the compiler, which is
	// written in Go, and freeing memory this side did not allocate is how a
	// heap gets corrupted.
	//
	// Neither is the heap: it is shared with the tau routines, which are never
	// joined and may well still be running. Freeing it here would pull the
	// objects out from under them. The process ending is what gives it back.
	pool_dispose(s.globals);
	pool_dispose(s.consts);
	modtab_dispose(s.mods);
}

struct vm *new_vm(char *file, struct bytecode bc) {
	gc_init();
	struct vm *vm = calloc(1, sizeof(struct vm));
	vm->file = file;
	vm->state = new_state();
	// The globals this program defines: an imported module gets its own
	// globals right after these ones.
	vm->state.ndefs = bc.ndefs;
	pool_extend(vm->state.consts, bc.consts, bc.nconsts);

	struct function *fn = new_function(bc.insts, bc.len, 0, 0, bc.bookmarks, bc.bklen);
	struct object cl = new_closure_obj(fn, NULL, 0);
	vm->frames[0] = new_frame(cl, 0);

	return vm;
}

struct vm *new_vm_with_state(char *file, struct bytecode bc, struct state state) {
	gc_init();
	struct vm *vm = calloc(1, sizeof(struct vm));
	vm->file = file;
	vm->state = state;
	vm->state.ndefs = bc.ndefs;
	// The constants of this unit go at the end of the pool the program already
	// has: an import and a REPL line are compiled knowing where their own
	// constants will land, so the indices in their bytecode are absolute.
	pool_extend(vm->state.consts, bc.consts, bc.nconsts);

	struct function *fn = new_function(bc.insts, bc.len, 0, 0, bc.bookmarks, bc.bklen);
	struct object cl = new_closure_obj(fn, NULL, 0);
	vm->frames[0] = new_frame(cl, 0);

	return vm;
}

void vm_dispose(struct vm *vm) {
	free(vm->file);
	free(vm);
}

static struct bookmark *vm_get_bookmark(struct vm * restrict vm) {
	struct frame *frame = vm_current_frame(vm);
	uint32_t offset = frame->ip - frame->start;
	size_t blen = frame->cl.data.cl->fn->bklen;
	struct bookmark *bookmarks = frame->cl.data.cl->fn->bookmarks;

	if (blen > 0) {
		for (size_t i = 0; i < blen; i++) {
			struct bookmark b = bookmarks[i];
			if (offset <= b.offset) {
				return &bookmarks[i];
			}
		}
	}
	return NULL;
}

inline void vm_errorf(struct vm * restrict vm, const char *fmt, ...) {
	struct bookmark *b = vm_get_bookmark(vm);

	if (b == NULL) {
		va_list args;
		va_start(args, fmt);
		vfprintf(stderr, fmt, args);
		va_end(args);
		longjmp(vm->env, 1);
	}

	char msg[512];
	va_list args;
	va_start(args, fmt);
	vsnprintf(msg, 512, fmt, args);
	va_end(args);

	char arrow[b->pos+2];
	memset(arrow, ' ', b->pos+2);
	arrow[b->pos] = '^';
	arrow[b->pos+1] = '\0';

	// What went wrong goes to standard error, so that the output of a program
	// stays what the program wrote.
	fflush(stdout);
	fprintf(
		stderr,
		"error in file %s at line %d:\n    %s\n    %s\n%s\n",
		vm->file,
		b->lineno,
		b->line,
		arrow,
		msg
	);

	longjmp(vm->env, 1);
}

void go_vm_errorf(struct vm * restrict vm, const char *fmt) {
	struct bookmark *b = vm_get_bookmark(vm);
	if (b == NULL) {
		fflush(stdout);
		fprintf(stderr, "%s\n", fmt);
		return;
	}

	char arrow[b->pos+2];
	memset(arrow, ' ', b->pos+2);
	arrow[b->pos] = '^';
	arrow[b->pos+1] = '\0';

	fflush(stdout);
	fprintf(
		stderr,
		"error in file %s at line %d:\n    %s\n    %s\n%s\n",
		vm->file,
		b->lineno,
		b->line,
		arrow,
		fmt
	);
}

static inline void vm_exec_dot(struct vm * restrict vm) {
	struct object right = vm_stack_pop(vm);
	struct object left = vm_stack_pop(vm);

	if (right.type != obj_string) {
		vm_errorf(vm, "%s object has no attribute %s", otype_str(left.type), object_str(right));
	}

	char *name = cstr(right.data.str);

	switch (left.type) {
	case obj_object:
		vm_stack_push(vm, object_get(left, name));
		cstr_free(right.data.str, name);
		return;

	case obj_native: {
		// Pointer to the native object.
		void *ptr = dlsym(left.data.handle, name);
		if (ptr == NULL) {
			vm_stack_push(vm, errorf("no object with name \"%s\" found", name));
			cstr_free(right.data.str, name);
			return;
		}
		struct object o = (struct object) {
			.data.handle = ptr,
			.type = obj_native,
		};
		vm_stack_push(vm, o);
		cstr_free(right.data.str, name);
		return;
	}

	default:
		cstr_free(right.data.str, name);
		vm_errorf(vm, "%s object has no attribute %s", otype_str(left.type), object_str(right));
	}
}

static inline void vm_exec_define(struct vm * restrict vm) {
	struct object val = vm_stack_pop(vm);
	struct object field = vm_stack_pop(vm);
	struct object target = vm_stack_pop(vm);

	switch (target.type) {
	case obj_object: {
		char *name = cstr(field.data.str);
		vm_stack_push(vm, object_set(target, name, val));
		cstr_free(field.data.str, name);
		return;
	}

	case obj_list: {
		struct object *list = target.data.list->list;
		size_t len = target.data.list->len;
		int64_t idx = field.data.i;

		if (idx < 0 || idx >= len) {
			vm_stack_push(vm, errorf("index out of range"));
			return;
		}
		list[idx] = val;
		vm_stack_push(vm, val);
		return;
	}

	case obj_map: {
		struct map_pair mp = map_set(target, field, val);
		vm_stack_push(vm, mp.val);
		return;
	}

	default:
		vm_errorf(vm, "cannot assign to type \"%s\"", otype_str(target.type));
	}
}

static inline void vm_push_closure(struct vm * restrict vm, uint32_t const_idx, uint32_t num_free) {
	struct object fn = vm->state.consts->list[const_idx];

	if (fn.type != obj_function) {
		vm_errorf(vm, "not a function %s", object_str(fn));
	}

	struct object *free = malloc(sizeof(struct object) * num_free);
	for (uint32_t i = 0; i < num_free; i++) {
		free[i] = vm->stack[vm->sp-num_free+i];
	}

	struct object cl = new_closure_obj(fn.data.fn, free, num_free);
	vm->sp -= num_free;
	vm_stack_push(vm, cl);

	vm_heap_add(vm, cl);
	gc();
}

static inline void vm_push_list(struct vm * restrict vm, uint32_t start, uint32_t end) {
	uint32_t len = end - start;
	struct object *list = malloc(sizeof(struct object) * len);

	for (uint32_t i = start; i < end; i++) {
		list[i-start] = vm->stack[i];
	}
	vm->sp -= len;
	struct object lst = new_list_obj(list, len);
	vm_stack_push(vm, lst);
	vm_heap_add(vm, lst);
	gc();
}

static inline void vm_push_map(struct vm * restrict vm, uint32_t start, uint32_t end) {
	struct object map = new_map();

	for (uint32_t i = start; i < end; i += 2) {
		struct object key = vm->stack[i];
		struct object val = vm->stack[i+1];

		switch (key.type) {
		case obj_integer:
		case obj_float:
		case obj_boolean:
		case obj_string:
		case obj_error:
			map_set(map, key, val);
			break;
		default:
			vm_errorf(vm, "invalid map key type %s", otype_str(key.type));
		}
	}

	vm->sp -= end - start;
	vm_stack_push(vm, map);
	vm_heap_add(vm, map);
	gc();
}

static inline void vm_push_interpolated(struct vm * restrict vm, uint32_t str_idx, uint32_t num_args) {
	struct object o = vm->state.consts->list[str_idx];
	char *str = o.data.str->str;
	size_t fmt_len = o.data.str->len;
	char *subs[num_args];
	uint32_t len_table[num_args];
	uint32_t sub_len = 0;

	for (int i = num_args-1; i >= 0; i--) {
		char *s = object_str(vm_stack_pop(vm));
		subs[i] = s;
		uint32_t len = strlen(s);
		len_table[i] = len;
		sub_len += len;
	}

	// One placeholder byte per substitution goes away, and one more byte
	// holds the NUL, which is not part of the string.
	uint32_t len = fmt_len + sub_len - num_args;
	char *ret = malloc(sizeof(char) * (len + 1));
	ret[len] = '\0';
	uint32_t retidx = 0;
	uint32_t subidx = 0;

	for (uint8_t *s = (uint8_t *) str; *s != '\0'; s++) {
		if (*s == 0xff) {
			strncpy(&ret[retidx], subs[subidx], len_table[subidx]);
			retidx += len_table[subidx];
			free(subs[subidx]);
			subidx++;
			continue;
		}
		ret[retidx++] = *s;
	}

	struct object res = new_string_obj(ret, len);
	vm_stack_push(vm, res);
	vm_heap_add(vm, res);
	gc();
}

static inline double to_double(struct object * restrict o) {
	if (o->type == obj_integer) {
		return o->data.i;
	}
	return o->data.f;
}

static inline void unsupported_operator_error(struct vm * restrict vm, char *op, struct object *l, struct object *r) {
	vm_errorf(vm, "unsupported operator '%s' for types %s and %s", op, otype_str(l->type), otype_str(r->type));
}

static inline void unsupported_prefix_operator_error(struct vm * restrict vm, char *op, struct object *o) {
	vm_errorf(vm, "unsupported operator '%s' for type %s", op, otype_str(o->type));
}

static inline void vm_exec_add(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (M_ASSERT(left, right, obj_integer)) {
		left->data.i += right->data.i;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		double l = to_double(left);
		double r = to_double(right);
		left->data.f = l + r;
		left->type = obj_float;
	} else if (M_ASSERT(left, right, obj_bytes)) {
		size_t llen = left->data.bytes->len;
		size_t rlen = right->data.bytes->len;
		uint8_t *b = malloc(llen + rlen);

		memcpy(b, left->data.bytes->bytes, llen);
		memcpy(b + llen, right->data.bytes->bytes, rlen);

		vm_stack_pop_ignore(vm);
		struct object res = new_bytes_obj(b, llen + rlen);
		vm_stack_push(vm, res);
		vm_heap_add(vm, res);
		gc();
	} else if (M_ASSERT(left, right, obj_string)) {
		// By length and not up to the NUL: a slice has none of its own, and
		// copying past it would both give the wrong result and overrun.
		size_t llen = left->data.str->len;
		size_t rlen = right->data.str->len;
		size_t slen = llen + rlen;
		char *str = malloc(sizeof(char) * (slen + 1));

		memcpy(str, left->data.str->str, llen);
		memcpy(str + llen, right->data.str->str, rlen);
		str[slen] = '\0';
		vm_stack_pop_ignore(vm);
		struct object res = new_string_obj(str, slen);
		vm_stack_push(vm, res);
		vm_heap_add(vm, res);
		gc();
	} else {
		unsupported_operator_error(vm, "+", left, right);
	}
}

static inline void vm_exec_sub(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (M_ASSERT(left, right, obj_integer)) {
		left->data.i -= right->data.i;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		double l = to_double(left);
		double r = to_double(right);
		left->data.f = l - r;
		left->type = obj_float;
	} else {
		unsupported_operator_error(vm, "-", left, right);
	}
}

static inline void vm_exec_mul(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (M_ASSERT(left, right, obj_integer)) {
		left->data.i *= right->data.i;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		double l = to_double(left);
		double r = to_double(right);
		left->data.f = l * r;
		left->type = obj_float;
	} else {
		unsupported_operator_error(vm, "*", left, right);
	}
}

static inline void vm_exec_div(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	// Two integers divide into an integer, the way they do in C and in Go:
	// the remainder is dropped rather than turned into a fraction nobody
	// asked for. A float on either side makes it a float division, as it
	// does for +, - and *. Write float(a) / float(b) for the fraction of two
	// integers.
	if (M_ASSERT(left, right, obj_integer)) {
		if (right->data.i == 0) {
			vm_errorf(vm, "can't divide by 0");
		}
		left->data.i /= right->data.i;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		double l = to_double(left);
		double r = to_double(right);
		left->data.f = l / r;
		left->type = obj_float;
	} else {
		unsupported_operator_error(vm, "/", left, right);
	}
}

static inline void vm_exec_mod(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (!M_ASSERT(left, right, obj_integer)) {
		unsupported_operator_error(vm, "%", left, right);
	}
	// Without this the machine raises SIGFPE and the whole process goes.
	if (right->data.i == 0) {
		vm_errorf(vm, "can't divide by 0");
	}
	left->data.i %= right->data.i;
}

static inline void vm_exec_and(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_pop(vm);

	vm_stack_push(vm, parse_bool(is_truthy(left) && is_truthy(right)));
}

static inline void vm_exec_or(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_pop(vm);

	vm_stack_push(vm, parse_bool(is_truthy(left) || is_truthy(right)));
}

static inline void vm_exec_bw_and(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (!M_ASSERT(left, right, obj_integer)) {
		unsupported_operator_error(vm, "&", left, right);
	}
	left->data.i &= right->data.i;
}

static inline void vm_exec_bw_or(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (!M_ASSERT(left, right, obj_integer)) {
		unsupported_operator_error(vm, "|", left, right);
	}
	left->data.i |= right->data.i;
}

static inline void vm_exec_bw_xor(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (!M_ASSERT(left, right, obj_integer)) {
		unsupported_operator_error(vm, "^", left, right);
	}
	left->data.i ^= right->data.i;
}

static inline void vm_exec_bw_not(struct vm * restrict vm) {
	struct object *right = &vm_stack_peek(vm);

	if (!ASSERT(right, obj_integer)) {
		unsupported_prefix_operator_error(vm, "~", right);
	}
	right->data.i = ~right->data.i;
}

static inline void vm_exec_bw_lshift(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (!M_ASSERT(left, right, obj_integer)) {
		unsupported_operator_error(vm, "<<", left, right);
	}
	left->data.i <<= right->data.i;
}

static inline void vm_exec_bw_rshift(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (!M_ASSERT(left, right, obj_integer)) {
		unsupported_operator_error(vm, ">>", left, right);
	}
	left->data.i >>= right->data.i;
}

// Orders two strings by their content, honouring the length instead of
// looking for a NUL that a slice doesn't have.
static inline int str_compare(struct string *l, struct string *r) {
	size_t min = l->len < r->len ? l->len : r->len;
	int cmp = memcmp(l->str, r->str, min);

	if (cmp != 0) return cmp;
	if (l->len == r->len) return 0;
	return l->len < r->len ? -1 : 1;
}

static inline void vm_exec_eq(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (M_ASSERT(left, right, obj_string)) {
		char *l = left->data.str->str;
		char *r = right->data.str->str;
		if (l == r) {
			*left = true_obj;
			return;
		}
		size_t lenl = left->data.str->len;
		size_t lenr = right->data.str->len;
		// memcmp and not strcmp: a slice ends where its length says, not at
		// the NUL of the string it was cut from.
		*left = (lenl == lenr) ? parse_bool(memcmp(l, r, lenl) == 0) : false_obj;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		*left = parse_bool(to_double(left) == to_double(right));
	} else if (left->type == right->type) {
		*left = parse_bool(memcmp(&left->data, &right->data, sizeof(union data)) == 0);
	} else {
		*left = false_obj;
	}
}

static inline void vm_exec_not_eq(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (M_ASSERT(left, right, obj_string)) {
		char *l = left->data.str->str;
		char *r = right->data.str->str;
		if (l == r) {
			*left = false_obj;
			return;
		}
		size_t lenl = left->data.str->len;
		size_t lenr = right->data.str->len;
		*left = (lenl == lenr) ? parse_bool(memcmp(l, r, lenl) != 0) : true_obj;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		*left = parse_bool(to_double(left) != to_double(right));
	} else if (left->type == right->type) {
		*left = parse_bool(memcmp(&left->data, &right->data, sizeof(union data)) != 0);
	} else {
		*left = true_obj;
	}
}

static inline void vm_exec_greater_than(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (M_ASSERT(left, right, obj_integer)) {
		left->data.i = left->data.i > right->data.i;
		left->type = obj_boolean;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		double l = to_double(left);
		double r = to_double(right);
		left->data.i = l > r;
		left->type = obj_boolean;
	} else if (M_ASSERT(left, right, obj_string)) {
		vm_stack_pop_ignore(vm);
		vm_stack_push(vm, parse_bool(str_compare(left->data.str, right->data.str) > 0));
	} else {
		unsupported_operator_error(vm, ">", left, right);
	}
}

static inline void vm_exec_greater_than_eq(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_peek(vm);

	if (M_ASSERT(left, right, obj_integer)) {
		left->data.i = left->data.i >= right->data.i;
		left->type = obj_boolean;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		double l = to_double(left);
		double r = to_double(right);
		left->data.i = l >= r;
		left->type = obj_boolean;
	} else if (M_ASSERT(left, right, obj_string)) {
		vm_stack_pop_ignore(vm);
		vm_stack_push(vm, parse_bool(str_compare(left->data.str, right->data.str) >= 0));
	} else {
		unsupported_operator_error(vm, ">=", left, right);
	}
}

static inline void vm_exec_minus(struct vm * restrict vm) {
	struct object *right = &vm_stack_peek(vm);

	switch (right->type) {
	case obj_integer:
		right->data.i = -right->data.i;
		break;
	case obj_float:
		right->data.f = -right->data.f;
		break;
	default:
		unsupported_prefix_operator_error(vm, "-", right);
		break;
	}
}

static inline void vm_exec_bang(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	vm_stack_push(vm, parse_bool(!is_truthy(right)));
}

// TODO: improve type assertion.
static inline void vm_exec_index(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = &vm_stack_pop(vm);

	if (ASSERT(left, obj_list) && ASSERT(right, obj_integer)) {
		struct object *list = left->data.list->list;
		size_t len = left->data.list->len;
		int64_t idx = right->data.i;

		if (idx < 0 || idx >= len) {
			vm_errorf(vm, "index out of range");
		}
		vm_stack_push(vm, list[idx]);
	} else if (ASSERT(left, obj_string) && ASSERT(right, obj_integer)) {
		char *str = left->data.str->str;
		size_t len = left->data.str->len;
		int64_t idx = right->data.i;

		if (idx < 0 || idx >= len) {
			vm_errorf(vm, "index out of range");
		}
		char *new_str = malloc(sizeof(char) * 2);
		new_str[0] = str[idx];
		new_str[1] = '\0';
		vm_stack_push(vm, new_string_obj(new_str, 1));
	} else if (ASSERT(left, obj_bytes) && ASSERT(right, obj_integer)) {
		uint8_t *b = left->data.bytes->bytes;
		size_t len = left->data.bytes->len;
		int64_t idx = right->data.i;

		if (idx < 0 || idx >= len) {
			vm_errorf(vm, "index out of range");
		}
		vm_stack_push(vm, new_integer_obj(b[idx]));
	} else if (ASSERT(left, obj_object) && ASSERT(right, obj_string)) {
		char *name = cstr(right->data.str);
		vm_stack_push(vm, object_get(*left, name));
		cstr_free(right->data.str, name);
	} else if (ASSERT(left, obj_native) && ASSERT(right, obj_string)) {
		// The dot on a shared object is dlsym with the name written in the
		// source; this is the same lookup with a name worked out while the
		// program runs, which is what taking a whole library at once needs.
		char *name = cstr(right->data.str);
		void *ptr = dlsym(left->data.handle, name);

		if (ptr == NULL) {
			vm_stack_push(vm, errorf("no object with name \"%s\" found", name));
			cstr_free(right->data.str, name);
			return;
		}
		cstr_free(right->data.str, name);

		struct object sym = {.data.handle = ptr, .type = obj_native};
		vm_stack_push(vm, sym);
	} else if (ASSERT(left, obj_map) && ASSERT4(right, obj_integer, obj_float, obj_string, obj_boolean)) {
		struct map_pair mp = map_get(*left, *right);
		vm_stack_push(vm, mp.val);
	} else {
		vm_errorf(vm, "invalid index operator for types %s and %s", otype_str(left->type), otype_str(right->type));
	}
}

static inline void vm_call_closure(struct vm * restrict vm, struct object *cl, size_t numargs) {
	size_t num_params = cl->data.cl->fn->num_params;

	if (num_params != numargs) {
		vm_errorf(vm, "wrong number of arguments: expected %d, got %lu", num_params, numargs);
	}

	struct frame frame = new_frame(*cl, vm->sp-numargs);
	vm_push_frame(vm, frame);

	// The locals that are not arguments start as null. Their slots hold
	// whatever the stack left there, so a read before the first assignment
	// would hand the program a stale object, and the collector would mark it
	// as live because it sits below sp.
	uint32_t num_locals = cl->data.cl->fn->num_locals;
	for (uint32_t i = numargs; i < num_locals; i++) {
		vm->stack[frame.base_ptr + i] = null_obj;
	}
	vm->sp = frame.base_ptr + num_locals;
}

static inline void vm_call_builtin(struct vm * restrict vm, builtin fn, size_t numargs) {
	struct object res = fn(&vm->stack[vm->sp-numargs], numargs);

	vm->sp -= numargs + 1;
	vm_stack_push(vm, res);
	if (res.type > obj_builtin) {
		vm_heap_add(vm, res);
		gc();
	}
}

// A native function that was given a signature knows how to marshal itself,
// so the VM only has to hand it the arguments where they lie and take back
// whatever tau value comes out, an error included.
static inline void vm_call_native_fn(struct vm * restrict vm, struct object *n, size_t numargs) {
	struct object res = native_call(*n, &vm->stack[vm->sp-numargs], numargs);

	vm->sp -= numargs + 1;
	vm_stack_push(vm, res);
	if (res.type > obj_builtin) {
		vm_heap_add(vm, res);
		gc();
	}
}

static inline void vm_call_native(struct vm * restrict vm, struct object *n, size_t numargs) {
	ffi_cif cif;
	ffi_type *arg_types[numargs];
	void *arg_values[numargs];
	// A C function reads a string up to its NUL: a slice doesn't have one, so
	// it travels as a copy that lives until the call returns.
	char *copies[numargs];
	char *strings[numargs];
	size_t ncopies = 0;

	// Convert Tau types to C types. The arguments are read where they are:
	// popping them here would leave them above sp, where the collector does
	// not look, and the call below parks this VM on purpose. A collection
	// while the C function runs would then free the very buffers it is
	// reading.
	struct object *args = &vm->stack[vm->sp - numargs];

	for (int64_t i = numargs - 1; i >= 0; i--) {
		struct object *o = &args[i];

		switch (o->type) {
		case obj_boolean:
		case obj_integer:
			arg_types[i] = &ffi_type_sint64;
			arg_values[i] = &o->data.i;
			break;

		case obj_float:
			arg_types[i] = &ffi_type_double;
			arg_values[i] = &o->data.f;
			break;

		case obj_string: {
			struct string *str = o->data.str;

			strings[i] = cstr(str);
			if (strings[i] != str->str) {
				copies[ncopies++] = strings[i];
			}

			arg_types[i] = &ffi_type_pointer;
			arg_values[i] = &strings[i];
			break;
		}

		case obj_bytes:
			arg_types[i] = &ffi_type_pointer;
			arg_values[i] = &o->data.bytes->bytes;
			break;

		case obj_native:
			arg_types[i] = &ffi_type_pointer;
			arg_values[i] = o->data.handle;
			break;

		case obj_null:
			arg_types[i] = &ffi_type_pointer;
			arg_values[i] = &o->data.i;
			break;

		// A value C has no idea what to do with is a mistake in the program,
		// not in the VM: it comes back as an error the caller can check.
		default:
			for (size_t j = 0; j < ncopies; j++) free(copies[j]);
			vm->sp -= numargs + 1;
			vm_stack_push(vm, errorf("unsupported argument type %s for native objects", otype_str(o->type)));
			vm_heap_add(vm, vm->stack[vm->sp-1]);
			return;
		}
	}

	if (ffi_prep_cif(&cif, FFI_DEFAULT_ABI, numargs, &ffi_type_pointer, arg_types) != FFI_OK) {
		for (size_t j = 0; j < ncopies; j++) free(copies[j]);
		vm->sp -= numargs + 1;
		vm_stack_push(vm, errorf("failed to prepare the native function"));
		vm_heap_add(vm, vm->stack[vm->sp-1]);
		return;
	}

	// A native call can block for an arbitrary amount of time (sockets, IO):
	// park so that it doesn't hold back a collection. The arguments are still
	// on the stack, so a collection happening now marks them and the buffers
	// the C function is reading stay where they are.
	ffi_arg return_value = 0;

	gc_park();
	ffi_call(&cif, n->data.handle, &return_value, arg_values);
	gc_unpark();

	// Now they can go: the arguments and the native object under them.
	vm->sp -= numargs + 1;

	for (size_t i = 0; i < ncopies; i++) {
		free(copies[i]);
	}

	// The result is the returned word itself, not a pointer to a buffer
	// holding it: there is nothing to free and nothing for the collector to
	// look after.
	struct object res = (struct object) {
		.data.i = (int64_t) return_value,
		.type = obj_native,
		.gc = NULL
	};
	vm_stack_push(vm, res);
}

static inline void vm_exec_call(struct vm * restrict vm, size_t numargs) {
	struct object *o = &vm->stack[vm->sp-1-numargs];

	switch (o->type) {
	case obj_closure:
		return vm_call_closure(vm, o, numargs);
	case obj_builtin:
		return vm_call_builtin(vm, o->data.builtin, numargs);
	case obj_native:
		return vm_call_native(vm, o, numargs);
	case obj_native_fn:
		return vm_call_native_fn(vm, o, numargs);
	default:
		vm_errorf(vm, "calling non-function: got type %s", otype_str(o->type));
	}
}

int vm_run(struct vm * restrict vm);

static int run_and_cleanup(void *vmptr) {
	struct vm *vm = vmptr;

	int ret = vm_run(vm);
	fflush(stdout);

	// The heap is global: whatever this VM allocated and is still reachable
	// from somewhere else (a pipe, the globals) survives.
	gc_unregister(vm);
	gc_release_segment();
	gc_flush_headers();
	vm_dispose(vm);
	return ret;
}

struct builtin_call_data {
	builtin fn;
	struct object *args;
	size_t numargs;
	void *roots;
};

static int call_builtin_and_cleanup(void *data) {
	struct builtin_call_data *d = data;

	d->fn(d->args, d->numargs);
	gc_remove_roots(d->roots);
	fflush(stdout);

	// A builtin allocates too, and this thread is about to end.
	gc_release_segment();
	gc_flush_headers();

	free(d->args);
	free(d);
	return 0;
}

static inline void vm_exec_concurrent_call(struct vm * restrict vm, uint32_t num_args) {
	thrd_t thread;
	struct object *o = &vm->stack[vm->sp-1-num_args];

	switch (o->type) {
	case obj_closure: {
		struct vm *tvm = calloc(1, sizeof(struct vm));
		tvm->file = strdup(vm->file);
		tvm->state.consts = vm->state.consts;    // The same pool, shared.
		tvm->state.mods = vm->state.mods;
		tvm->state.globals = vm->state.globals;  // Shared, never reallocated.

		// Only copy the closure and its arguments to the new VM's stack
		// Stack layout: [closure, arg0, arg1, ..., argN-1]
		memcpy(tvm->stack, &vm->stack[vm->sp-1-num_args], (num_args + 1) * sizeof(struct object));
		tvm->sp = num_args + 1;

		vm_call_closure(tvm, o, num_args);
		// Registered here and not in the new thread: until the thread starts
		// running, the objects on its stack are reachable only from here.
		gc_register(tvm);
		if (thrd_create(&thread, run_and_cleanup, tvm) != thrd_success) {
			gc_unregister(tvm);
			vm_errorf(vm, "failed to create thread");
		}
		// Nobody ever joins a tau routine, so the thread has to be told that
		// its remains are of no interest: a thread that is neither joined nor
		// detached keeps its stack and its descriptor until the process ends,
		// and a program that starts a few thousand of them runs out of them.
		thrd_detach(thread);
		break;
	}

	case obj_builtin: {
		struct builtin_call_data *d = malloc(sizeof(struct builtin_call_data));
		d->fn = o->data.builtin;
		d->args = malloc(sizeof(struct object) * num_args);
		d->numargs = num_args;
		memcpy(d->args, &vm->stack[vm->sp-num_args], num_args * sizeof(struct object));
		d->roots = gc_add_roots(d->args, num_args);

		if (thrd_create(&thread, call_builtin_and_cleanup, d) != thrd_success) {
			gc_remove_roots(d->roots);
			vm_errorf(vm, "failed to create thread");
		}
		thrd_detach(thread);
		break;
	}

	default:
		vm_errorf(vm, "calling non-function: got type %s", otype_str(o->type));
	}

	// Drop the closure and its arguments and leave null in their place: a call
	// is an expression, and the statement it belongs to pops exactly one value.
	// Without this every `tau f(x)` would leave its arguments behind, and a
	// loop that starts routines would walk the stack pointer off the end of the
	// stack. It is done only once the routine can be reached from its own VM,
	// so the objects are never unreachable in between.
	vm->sp -= num_args + 1;
	vm_stack_push(vm, null_obj);
}

static inline void vm_exec_return(struct vm * restrict vm) {
	struct frame *frame = vm_pop_frame(vm);

	vm->sp = frame->base_ptr - 1;
	vm_stack_push(vm, null_obj);
}

static inline void vm_exec_return_value(struct vm * restrict vm) {
	struct object *o = &vm_stack_pop(vm);
	struct frame *frame = vm_pop_frame(vm);

	vm->sp = frame->base_ptr - 1;
	vm_stack_push(vm, *o);
}

struct object vm_last_popped_stack_elem(struct vm * restrict vm) {
	return vm->stack[vm->sp];
}


/*
 * The following comment is taken from CPython's source:
 * https://github.com/python/cpython/blob/3.11/Python/ceval.c#L1243

 * Computed GOTOs, or
       the-optimization-commonly-but-improperly-known-as-"threaded code"
 * using gcc's labels-as-values extension
 * (http://gcc.gnu.org/onlinedocs/gcc/Labels-as-Values.html).

 * The traditional bytecode evaluation loop uses a "switch" statement, which
 * decent compilers will optimize as a single indirect branch instruction
 * combined with a lookup table of jump addresses. However, since the
 * indirect jump instruction is shared by all opcodes, the CPU will have a
 * hard time making the right prediction for where to jump next (actually,
 * it will be always wrong except in the uncommon case of a sequence of
 * several identical opcodes).

 * "Threaded code" in contrast, uses an explicit jump table and an explicit
 * indirect jump instruction at the end of each opcode. Since the jump
 * instruction is at a different address for each opcode, the CPU will make a
 * separate prediction for each of these instructions, which is equivalent to
 * predicting the second opcode of each opcode pair. These predictions have
 * a much better chance to turn out valid, especially in small bytecode loops.

 * A mispredicted branch on a modern CPU flushes the whole pipeline and
 * can cost several CPU cycles (depending on the pipeline depth),
 * and potentially many more instructions (depending on the pipeline width).
 * A correctly predicted branch, however, is nearly free.

 * At the time of this writing, the "threaded code" version is up to 15-20%
 * faster than the normal "switch" version, depending on the compiler and the
 * CPU architecture.

 * NOTE: care must be taken that the compiler doesn't try to "optimize" the
 * indirect jumps by sharing them between all opcodes. Such optimizations
 * can be disabled on gcc by using the -fno-gcse flag (or possibly
 * -fno-crossjumping).
 */

static int vm_loop(struct vm * restrict vm);

// Registers the VM as a GC root set and runs it.
int vm_run(struct vm * restrict vm) {
	// The VM of a tau routine is registered by whoever spawned it.
	int owned = vm->gc_node == NULL;
	if (owned) gc_register(vm);

	void *prev = gc_activate(vm);
	int ret = vm_loop(vm);

	if (owned) {
		gc_unregister(vm);
	} else {
		gc_park();
	}
	gc_restore(prev);

	return ret;
}

// TODO: maybe return a char *.
static int vm_loop(struct vm * restrict vm) {
#include "jump_table.h"

	// Used by vm_errorf to stop the execution of the VM without exiting.
	if (setjmp(vm->env) == 1) {
		return 1;
	}
	register struct frame *frame = vm_current_frame(vm);
	DISPATCH();

	TARGET_POP: {
		vm_stack_pop_ignore(vm);
		DISPATCH();
	}

	TARGET_CONST: {
		uint16_t idx = read_uint16(frame->ip);
		frame->ip += 2;
		vm_stack_push(vm, vm->state.consts->list[idx]);
		DISPATCH();
	}

	TARGET_TRUE: {
		vm_stack_push(vm, true_obj);
		DISPATCH();
	}

	TARGET_FALSE: {
		vm_stack_push(vm, false_obj);
		DISPATCH();
	}

	TARGET_NULL: {
		vm_stack_push(vm, null_obj);
		DISPATCH();
	}

	TARGET_LIST: {
		uint32_t len = read_uint16(frame->ip);
		frame->ip += 2;
		vm_push_list(vm, vm->sp-len, vm->sp);
		DISPATCH();
	}

	TARGET_MAP: {
		uint32_t len = read_uint16(frame->ip);
		frame->ip += 2;
		vm_push_map(vm, vm->sp-len, vm->sp);
		DISPATCH();
	}

	TARGET_CLOSURE: {
		uint16_t const_idx = read_uint16(frame->ip);
		uint8_t num_free = read_uint8(frame->ip+2);
		frame->ip += 3;
		vm_push_closure(vm, const_idx, num_free);
		DISPATCH();
	}

	TARGET_CURRENT_CLOSURE: {
		vm_stack_push(vm, frame->cl);
		DISPATCH();
	}

	TARGET_ADD: {
		vm_exec_add(vm);
		DISPATCH();
	}

	TARGET_SUB: {
		vm_exec_sub(vm);
		DISPATCH();
	}

	TARGET_MUL: {
		vm_exec_mul(vm);
		DISPATCH();
	}

	TARGET_DIV: {
		vm_exec_div(vm);
		DISPATCH();
	}

	TARGET_MOD: {
		vm_exec_mod(vm);
		DISPATCH();
	}

	TARGET_BW_AND: {
		vm_exec_bw_and(vm);
		DISPATCH();
	}

	TARGET_BW_OR: {
		vm_exec_bw_or(vm);
		DISPATCH();
	}

	TARGET_BW_XOR: {
		vm_exec_bw_xor(vm);
		DISPATCH();
	}

	TARGET_BW_NOT: {
		vm_exec_bw_not(vm);
		DISPATCH();
	}

	TARGET_BW_LSHIFT: {
		vm_exec_bw_lshift(vm);
		DISPATCH();
	}

	TARGET_BW_RSHIFT: {
		vm_exec_bw_rshift(vm);
		DISPATCH();
	}

	TARGET_AND: {
		vm_exec_and(vm);
		DISPATCH();
	}

	TARGET_OR: {
		vm_exec_or(vm);
		DISPATCH();
	}

	TARGET_EQUAL: {
		vm_exec_eq(vm);
		DISPATCH();
	}

	TARGET_NOT_EQUAL: {
		vm_exec_not_eq(vm);
		DISPATCH();
	}

	TARGET_GREATER_THAN: {
		vm_exec_greater_than(vm);
		DISPATCH();
	}

	TARGET_GREATER_THAN_EQUAL: {
		vm_exec_greater_than_eq(vm);
		DISPATCH();
	}

	TARGET_MINUS: {
		vm_exec_minus(vm);
		DISPATCH();
	}

	TARGET_BANG: {
		vm_exec_bang(vm);
		DISPATCH();
	}

	TARGET_INDEX: {
		vm_exec_index(vm);
		DISPATCH();
	}

	TARGET_CALL: {
		uint8_t num_args = read_uint8(frame->ip++);
		if (gc_pending()) gc_safepoint();
		vm_exec_call(vm, num_args);
		frame = vm_current_frame(vm);
		DISPATCH();
	}

	TARGET_CONCURRENT_CALL: {
		uint8_t num_args = read_uint8(frame->ip++);
		vm_exec_concurrent_call(vm, num_args);
		DISPATCH();
	}

	TARGET_RETURN: {
		vm_exec_return(vm);
		frame = vm_current_frame(vm);
		if (frame->ip == NULL) goto TARGET_HALT;
		DISPATCH();
	}

	TARGET_RETURN_VALUE: {
		vm_exec_return_value(vm);
		frame = vm_current_frame(vm);
		if (frame->ip == NULL) goto TARGET_HALT;
		DISPATCH();
	}

	TARGET_JUMP: {
		uint16_t pos = read_uint16(frame->ip);
		frame->ip = &frame->start[pos];
		// Safepoint: a loop that doesn't allocate would otherwise never let
		// another VM collect.
		if (gc_pending()) gc_safepoint();
		DISPATCH();
	}

	TARGET_JUMP_NOT_TRUTHY: {
		uint16_t pos = read_uint16(frame->ip);
		frame->ip += 2;
		if (gc_pending()) gc_safepoint();

		struct object *cond = &vm_stack_pop(vm);
		if (!is_truthy(cond)) {
			frame->ip = &frame->start[pos];
		}
		DISPATCH();
	}

	TARGET_DOT: {
		vm_exec_dot(vm);
		DISPATCH();
	}

	TARGET_DEFINE: {
		vm_exec_define(vm);
		DISPATCH();
	}

	TARGET_GET_GLOBAL: {
		uint32_t global_idx = read_uint16(frame->ip);
		frame->ip += 2;
		vm_stack_push(vm, vm->state.globals->list[global_idx]);
		DISPATCH();
	}

	TARGET_SET_GLOBAL: {
		uint32_t global_idx = read_uint16(frame->ip);
		frame->ip += 2;
		pool_insert(vm->state.globals, global_idx, vm_stack_peek(vm));
		DISPATCH();
	}

	TARGET_GET_LOCAL: {
		uint32_t local_idx = read_uint8(frame->ip++);
		vm_stack_push(vm, vm->stack[frame->base_ptr+local_idx]);
		DISPATCH();
	}

	TARGET_SET_LOCAL: {
		uint32_t local_idx = read_uint8(frame->ip++);
		vm->stack[frame->base_ptr+local_idx] = vm_stack_peek(vm);
		DISPATCH();
	}

	TARGET_GET_BUILTIN: {
		uint32_t idx = read_uint8(frame->ip++);
		vm_stack_push(vm, new_builtin_obj(builtins[idx]));
		DISPATCH();
	}

	TARGET_GET_FREE: {
		uint32_t free_idx = read_uint8(frame->ip++);
		struct object cl = frame->cl;
		vm_stack_push(vm, cl.data.cl->free[free_idx]);
		DISPATCH();
	}

	TARGET_LOAD_MODULE: {
		struct object path = vm_stack_pop(vm);
		if (path.type != obj_string) {
			vm_errorf(vm, "import: expected string, got %s", otype_str(path.type));
		}
		char *modpath = cstr(path.data.str);
		int failed = vm_exec_load_module(vm, modpath);
		cstr_free(path.data.str, modpath);
		if (failed) {
			return 1;
		}
		DISPATCH();
	}

	TARGET_INTERPOLATE: {
		uint32_t str_idx = read_uint16(frame->ip);
		uint32_t num_args = read_uint16(frame->ip+2);
		frame->ip += 4;
		vm_push_interpolated(vm, str_idx, num_args);
		DISPATCH();
	}

	TARGET_HALT:
		return 0;
}

#if !defined(_WIN32) && !defined(WIN32)
	#include <termios.h>
	#include <unistd.h>

// The terminal as it was before anything touched it. The REPL puts it in raw
// mode, and the exit builtin calls exit() straight from a tau program, which
// runs no Go deferred call and would leave the shell in raw mode. Saved and
// put back on this side, so that nothing of the terminal has to cross over.
static struct termios term_state;
static int term_saved = 0;

static void restore_term(void) {
	if (term_saved) tcsetattr(STDIN_FILENO, TCSANOW, &term_state);
}

void set_exit() {
	term_saved = tcgetattr(STDIN_FILENO, &term_state) == 0;
	atexit(restore_term);
}
#else
// ponytail: on Windows raw mode is a console mode and not a termios, and the
// REPL already puts it back on the way out. Only exit() from a tau program
// gets away with it there.
void set_exit() {}
#endif
