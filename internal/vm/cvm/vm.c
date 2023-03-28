#include <stdarg.h>
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

#include "vm.h"
#include "opcode.h"
#include "obj.h"
#include "_cgo_export.h"

#define read_uint8(b) ((b)[0])
#define read_uint16(b) (((b)[0] << 8) | (b)[1])
#define read_uint32(b) (((b)[0] << 24) | ((b)[1] << 16) | ((b)[2] << 8) | (b)[3])

#define vm_current_frame(vm) (&vm->frames[vm->frame_idx])
#define vm_push_frame(vm, frame) vm->frames[++vm->frame_idx] = frame
#define vm_pop_frame(vm) (&vm->frames[vm->frame_idx--])

#define vm_stack_push(vm, obj) vm->stack[vm->sp++] = obj
#define vm_stack_pop(vm) (vm->stack[--vm->sp])
#define vm_stack_pop_ignore(vm) vm->sp--
#define vm_stack_peek(vm) (vm->stack[vm->sp-1])

#define DISPATCH() goto *jump_table[*frame->ip++]
#define UNHANDLED() printf("unhandled opcode %s\n", opcode_str(*(frame->ip-1))); return -1

#define ASSERT(obj, t) (obj->type == t)
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

struct state new_state() {
	return (struct state) {
		.consts = calloc(0, sizeof(struct object)),
		.nconsts = 0,
		.globals = {null_obj}
	};
}

struct vm *new_vm(char *file, struct bytecode bc) {
	struct vm *vm = calloc(1, sizeof(struct vm));
	vm->file = file;
	vm->state.consts = bc.consts;
	vm->state.ndefs = bc.ndefs;

	struct object fn = new_function_obj(bc.insts, bc.len, 0, 0, bc.bookmarks, bc.bklen);
	struct object cl = new_closure_obj(fn.data.fn, NULL, 0);
	vm->frames[0] = new_frame(cl, 0);

	return vm;
}

struct vm *new_vm_with_state(char *file, struct bytecode bc, struct state state) {
	struct vm *vm = new_vm(file, bc);
	vm->state = state;

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
		for (int i = 0; i < blen; i++) {
			struct bookmark b = bookmarks[i];
			if (offset <= b.offset) {
				return &bookmarks[i];
			}
		}
	}
	return NULL;
}

void vm_errorf(struct vm * restrict vm, const char *fmt, ...) {
	struct bookmark *b = vm_get_bookmark(vm);

	if (b == NULL) {
		va_list args;
		va_start(args, fmt);
		vprintf(fmt, args);
		va_end(args);
		exit(1);
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

	printf(
		"error in file %s at line %d:\n    %s\n    %s\n%s\n",
		vm->file,
		b->lineno,
		b->line,
		arrow,
		msg
	);
	exit(1);
}

void go_vm_errorf(struct vm * restrict vm, const char *fmt) {
	vm_errorf(vm, fmt);
}

static inline __attribute__((always_inline))
struct object *unwrap(struct object *o) {
	if (o->type == obj_getsetter) {
		struct getsetter *gs = o->data.gs;
		*o = gs->get(gs);
		free(gs);
	}
	return o;
}

static inline __attribute__((always_inline))
struct object unwraps(struct object o) {
	if (o.type == obj_getsetter) {
		struct getsetter *gs = o.data.gs;
		o = gs->get(gs);
		free(gs);
	}
	return o;
}

static inline void vm_exec_dot(struct vm * restrict vm) {
	struct object *right = &vm_stack_pop(vm);
	struct object *left = unwrap(&vm_stack_pop(vm));

	if (!ASSERT(left, obj_object) || !ASSERT(right, obj_string)) {
		vm_errorf(vm, "%s object has no attribute %s", otype_str(left->type), object_str(*right));
	}
	vm_stack_push(vm, new_getsetter_obj(*left, *right, object_getsetter_get, object_getsetter_set));
}

static inline void vm_exec_define(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = &vm_stack_pop(vm);

	if (!ASSERT(left, obj_getsetter)) {
		vm_errorf(vm, "cannot assign to type \"%s\"", otype_str(left->type));
	}
	struct getsetter *gs = left->data.gs;
	vm_stack_push(vm, gs->set(gs, *right));
	free(gs);
}

static inline void vm_push_closure(struct vm * restrict vm, uint32_t const_idx, uint32_t num_free) {
	struct object fn = vm->state.consts[const_idx];

	if (fn.type != obj_function) {
		vm_errorf(vm, "not a function %s", object_str(fn));
	}
	
	struct object *free = malloc(sizeof(struct object) * num_free);
	for (int i = 0; i < num_free; i++) {
		free[i] = vm->stack[vm->sp-num_free+i];
	}

	struct object cl = new_closure_obj(fn.data.fn, free, num_free);
	vm->sp -= num_free;
	vm_stack_push(vm, cl);
}

static inline void vm_push_list(struct vm * restrict vm, uint32_t start, uint32_t end) {
	uint32_t len = end - start;
	struct object *list = malloc(sizeof(struct object) * len);

	for (uint32_t i = start; i < end; i++) {
		list[i-start] = vm->stack[i];
	}
	vm->sp -= len;
	vm_stack_push(vm, new_list_obj(list, len));
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
}

static inline void vm_push_interpolated(struct vm * restrict vm, uint32_t str_idx, uint32_t num_args) {
	struct object o = vm->state.consts[str_idx];
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

	uint32_t len = fmt_len + sub_len - num_args + 1;
	char *ret = malloc(sizeof(char) * len);
	ret[len-1] = '\0';
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

	vm_stack_push(vm, new_string_obj(ret, len));
}

static inline double to_double(struct object * restrict o) {
	if (ASSERT(o, obj_integer)) {
		return o->data.i;
	}
	return o->data.f;
}

static inline uint32_t is_truthy(struct object * restrict o) {
	switch (o->type) {
	case obj_boolean:
		return o->data.i == 1;
	case obj_integer:
		return o->data.i != 0;
	case obj_float:
		return o->data.f != 0;
	case obj_null:
		return 0;
	default:
		return 1;
	}
}

static inline void unsupported_operator_error(struct vm * restrict vm, char *op, struct object *l, struct object *r) {
	vm_errorf(vm, "unsupported operator '%s' for types %s and %s", op, otype_str(l->type), otype_str(r->type));
}

static inline void unsupported_prefix_operator_error(struct vm * restrict vm, char *op, struct object *o) {
	vm_errorf(vm, "unsupported operator '%s' for type %s", op, otype_str(o->type));
}

static inline void vm_exec_add(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

	if (M_ASSERT(left, right, obj_integer)) {
		left->data.i += right->data.i;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		double l = to_double(left);
		double r = to_double(right);
		left->data.f = l + r;
		left->type = obj_float;
	} else if (M_ASSERT(left, right, obj_string)) {
		size_t slen = left->data.str->len + right->data.str->len;
		char *str = malloc(sizeof(char) * (slen + 1));
		char *start = strcpy(str, left->data.str->str);
		strcpy(start, right->data.str->str);
		vm_stack_pop_ignore(vm);
		vm_stack_push(vm, new_string_obj(str, slen));
	} else {
		unsupported_operator_error(vm, "+", left, right);
	}
}

static inline void vm_exec_sub(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

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
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

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
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

	if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		double l = to_double(left);
		double r = to_double(right);
		left->data.f = l / r;
		left->type = obj_float;
	} else {
		unsupported_operator_error(vm, "/", left, right);
	}
}

static inline void vm_exec_mod(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

	if (!M_ASSERT(left, right, obj_integer)) {
		unsupported_operator_error(vm, "%", left, right);
	}
	left->data.i %= right->data.i;
}

static inline void vm_exec_and(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_pop(vm));

	vm_stack_push(vm, parse_bool(is_truthy(left) && is_truthy(right)));
}

static inline void vm_exec_or(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_pop(vm));

	vm_stack_push(vm, parse_bool(is_truthy(left) || is_truthy(right)));
}

static inline void vm_exec_bw_and(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

	if (!M_ASSERT(left, right, obj_integer)) {
		unsupported_operator_error(vm, "&", left, right);
	}
	left->data.i &= right->data.i;
}

static inline void vm_exec_bw_or(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

	if (!M_ASSERT(left, right, obj_integer)) {
		unsupported_operator_error(vm, "|", left, right);
	}
	left->data.i |= right->data.i;
}

static inline void vm_exec_bw_xor(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

	if (!M_ASSERT(left, right, obj_integer)) {
		unsupported_operator_error(vm, "^", left, right);
	}
	left->data.i ^= right->data.i;
}

static inline void vm_exec_bw_not(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_peek(vm));

	if (!ASSERT(right, obj_integer)) {
		unsupported_prefix_operator_error(vm, "~", right);
	}
	right->data.i = ~right->data.i;
}

static inline void vm_exec_bw_lshift(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

	if (!M_ASSERT(left, right, obj_integer)) {
		unsupported_operator_error(vm, "<<", left, right);
	}
	left->data.i <<= right->data.i;
}

static inline void vm_exec_bw_rshift(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

	if (!M_ASSERT(left, right, obj_integer)) {
		unsupported_operator_error(vm, ">>", left, right);
	}
	left->data.i >>= right->data.i;
}

static inline void vm_exec_eq(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

	if (M_ASSERT2(left, right, obj_boolean, obj_null)) {
		left->data.i = left->type == right->type && left->data.i == right->data.i;
		left->type = obj_boolean;
	} else if (M_ASSERT(left, right, obj_integer)) {
		left->data.i = left->data.i == right->data.i;
		left->type = obj_boolean;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		double l = to_double(left);
		double r = to_double(right);
		left->data.i = l == r;
		left->type = obj_boolean;
	} else if (M_ASSERT(left, right, obj_string)) {
		char *l = left->data.str->str;
		char *r = right->data.str->str;
		size_t lenl = left->data.str->len;
		size_t lenr = right->data.str->len;
		struct object res = (lenl == lenr) ? parse_bool(strcmp(l, r) == 0) : false_obj;
		vm_stack_pop_ignore(vm);
		vm_stack_push(vm, res);
	} else {
		vm_stack_push(vm, false_obj);
	}
}

static inline void vm_exec_not_eq(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

	if (M_ASSERT2(left, right, obj_boolean, obj_null)) {
		left->data.i = left->type != right->type && left->data.i != right->data.i;
		left->type = obj_boolean;
	} else if (M_ASSERT(left, right, obj_integer)) {
		left->data.i = left->data.i != right->data.i;
		left->type = obj_boolean;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		double l = to_double(left);
		double r = to_double(right);
		left->data.i = l != r;
		left->type = obj_boolean;
	} else if (M_ASSERT(left, right, obj_string)) {
		char *l = left->data.str->str;
		char *r = right->data.str->str;
		size_t lenl = left->data.str->len;
		size_t lenr = right->data.str->len;
		struct object res = (lenl == lenr) ? parse_bool(strcmp(l, r) != 0) : false_obj;
		vm_stack_pop_ignore(vm);
		vm_stack_push(vm, res);
	} else {
		vm_stack_push(vm, false_obj);
	}
}

static inline void vm_exec_greater_than(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

	if (M_ASSERT(left, right, obj_integer)) {
		left->data.i = left->data.i > right->data.i;
		left->type = obj_boolean;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		double l = to_double(left);
		double r = to_double(right);
		left->data.i = l > r;
		left->type = obj_boolean;
	} else if (M_ASSERT(left, right, obj_string)) {
		char *l = left->data.str->str;
		char *r = right->data.str->str;
		vm_stack_pop_ignore(vm);
		vm_stack_push(vm, parse_bool(strcmp(l, r) > 0));
	} else {
		unsupported_operator_error(vm, ">", left, right);
	}
}

static inline void vm_exec_greater_than_eq(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_peek(vm));

	if (M_ASSERT(left, right, obj_integer)) {
		left->data.i = left->data.i >= right->data.i;
		left->type = obj_boolean;
	} else if (M_ASSERT2(left, right, obj_integer, obj_float)) {
		double l = to_double(left);
		double r = to_double(right);
		left->data.i = l >= r;
		left->type = obj_boolean;
	} else if (M_ASSERT(left, right, obj_string)) {
		char *l = left->data.str->str;
		char *r = right->data.str->str;
		vm_stack_pop_ignore(vm);
		vm_stack_push(vm, parse_bool(strcmp(l, r) >= 0));
	} else {
		unsupported_operator_error(vm, ">", left, right);
	}
}

static inline void vm_exec_minus(struct vm * restrict vm) {
	struct object *right = unwrap(&vm_stack_peek(vm));

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
	struct object *right = unwrap(&vm_stack_pop(vm));

	switch (right->type) {
	case obj_boolean:
		vm_stack_push(vm, parse_bool(!right->data.i));
		break;
	case obj_null:
		vm_stack_push(vm, true_obj);
		break;
	default:
		vm_stack_push(vm, false_obj);
		break;
	}
}

// TODO: add support for bytes.
static inline void vm_exec_index(struct vm * restrict vm) {
	struct object *index = unwrap(&vm_stack_pop(vm));
	struct object *left = unwrap(&vm_stack_pop(vm));

	if (ASSERT(left, obj_list) && ASSERT(index, obj_integer)) {
		vm_stack_push(vm, new_getsetter_obj(*left, *index, list_getsetter_get, list_getsetter_set));
	} else if (ASSERT(left, obj_string) && ASSERT(index, obj_integer)) {
		char *str = left->data.str->str;
		size_t len = left->data.str->len;
		int64_t idx = index->data.i;

		if (idx < 0 || idx > len) {
			vm_errorf(vm, "index out of range");
		}
		char *new_str = malloc(sizeof(char) * 2);
		new_str[0] = str[idx];
		new_str[1] = '\0';
		vm_stack_push(vm, new_string_obj(new_str, 2));
	} else if (ASSERT(left, obj_map) && ASSERT4(index, obj_integer, obj_float, obj_string, obj_boolean)) {
		vm_stack_push(vm, new_getsetter_obj(*left, *index, map_getsetter_get, map_getsetter_set));
	} else {
		vm_errorf(vm, "invalid index operator for types %s and %s", otype_str(left->type), otype_str(index->type));
	}
}

static inline void vm_call_closure(struct vm * restrict vm, struct object *cl, size_t numargs) {
	int num_params = cl->data.cl->fn->num_params;

	if (num_params != numargs) {
		vm_errorf(vm, "wrong number of arguments: expected %d, got %lu", num_params, numargs);
	}

	struct frame frame = new_frame(*cl, vm->sp-numargs);
	vm_push_frame(vm, frame);
	vm->sp = frame.base_ptr + cl->data.cl->fn->num_locals;
}

static inline void vm_call_builtin(struct vm * restrict vm, builtin fn, size_t numargs) {
	struct object res = fn(&vm->stack[vm->sp-numargs], numargs);

	vm->sp -= numargs + 1;
	vm_stack_push(vm, res);
}

static inline void vm_exec_call(struct vm * restrict vm, size_t numargs) {
	struct object *o = unwrap(&vm->stack[vm->sp-1-numargs]);

	switch (o->type) {
	case obj_closure:
		return vm_call_closure(vm, o, numargs);
	case obj_builtin:
		return vm_call_builtin(vm, o->data.builtin, numargs);
	default:
		vm_errorf(vm, "calling non-function");
	}
}

int vm_run(struct vm * restrict vm);

static inline void vm_exec_concurrent_call(struct vm * restrict vm, uint32_t num_args) {
	struct vm *tvm = calloc(1, sizeof(struct vm));
	tvm->file = vm->file;
	tvm->state.consts = vm->state.consts;
	tvm->state.nconsts = vm->state.nconsts;
	tvm->sp = vm->sp;
	tvm->frame_idx = 1;

	#pragma omp parallel default(none) shared(vm, tvm)
	#pragma omp single
	{
		#pragma omp task
		memcpy(tvm->stack, vm->stack, STACK_SIZE);
		#pragma omp task
		memcpy(tvm->state.globals, vm->state.globals, GLOBAL_SIZE);
		#pragma omp taskwait
	}
	vm_exec_call(tvm, num_args);

	#pragma omp single
	{
		vm_run(tvm);
		free(tvm);
	}
	vm_stack_push(vm, null_obj);
}

static inline void vm_exec_return(struct vm * restrict vm) {
	struct frame *frame = vm_pop_frame(vm);
	vm->sp = frame->base_ptr - 1;
	vm_stack_push(vm, null_obj);
}

static inline void vm_exec_return_value(struct vm * restrict vm) {
	struct object *o = unwrap(&vm_stack_pop(vm));
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

int vm_run(struct vm * restrict vm) {
#include "jump_table.h"

	register struct frame *frame = vm_current_frame(vm);
	DISPATCH();

	TARGET_CONST: {
		uint16_t idx = read_uint16(frame->ip);
		frame->ip += 2;
		vm_stack_push(vm, vm->state.consts[idx]);
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
		DISPATCH();
	}

	TARGET_JUMP_NOT_TRUTHY: {
		uint16_t pos = read_uint16(frame->ip);
		frame->ip += 2;

		struct object *cond = unwrap(&vm_stack_pop(vm));
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
		vm_stack_push(vm, vm->state.globals[global_idx]);
		DISPATCH();
	}

	TARGET_SET_GLOBAL: {
		uint32_t global_idx = read_uint16(frame->ip);
		frame->ip += 2;
		vm->state.globals[global_idx] = unwraps(vm_stack_peek(vm));
		DISPATCH();
	}

	TARGET_GET_LOCAL: {
		uint32_t local_idx = read_uint8(frame->ip++);
		vm_stack_push(vm, vm->stack[frame->base_ptr+local_idx]);
		DISPATCH();
	}

	TARGET_SET_LOCAL: {
		uint32_t local_idx = read_uint8(frame->ip++);
		vm->stack[frame->base_ptr+local_idx] = unwraps(vm_stack_peek(vm));
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
		vm_exec_load_module(vm, path.data.str->str);
		DISPATCH();
	}

	TARGET_INTERPOLATE: {
		uint32_t str_idx = read_uint16(frame->ip);
		uint32_t num_args = read_uint16(frame->ip+2);
		frame->ip += 4;
		vm_push_interpolated(vm, str_idx, num_args);
		DISPATCH();
	}

	TARGET_POP: {
		vm_stack_pop_ignore(vm);
		DISPATCH();
	}

	TARGET_HALT:
		return 0;
}
