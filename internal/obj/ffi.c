#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "object.h"
#include "libffi/include/ffi.h"

// A native function that was told what it looks like. Without a signature the
// VM has to guess the argument types from the tau values and can only ever
// bring back a machine word; with one every argument and the result travel as
// the C type they really are, so a double comes back as a float, an int32 as
// an int, and a void function as null.
//
// The cif is prepared once, here, and not at every call: it points at the
// types array, which is why that array lives in the same allocation.
struct native {
	void *fn;
	ffi_cif cif;
	char ret;
	char *args;
	uint8_t nargs;
	ffi_type *types[];
};

// The types a call can be made of. The numbers are what tau passes in, and
// stdlib/ffi.tau is where they are given names: this is one half of an
// agreement, the other half is the list of constants in that file.
//
// Internally each one is a letter, which is what the marshalling below
// switches on: lowercase signed, uppercase the unsigned of the same width.
enum {
	CODE_VOID, CODE_BOOL,
	CODE_INT8, CODE_UINT8,
	CODE_INT16, CODE_UINT16,
	CODE_INT32, CODE_UINT32,
	CODE_INT64, CODE_UINT64,
	CODE_FLOAT32, CODE_FLOAT64,
	CODE_POINTER, CODE_CSTRING,
	CODE_MAX
};

static const char code_letters[CODE_MAX] = {
	'v', 'b', 'c', 'C', 's', 'S', 'i', 'I', 'l', 'L', 'f', 'd', 'p', 'z'
};

// letter_of returns the letter of a type code, or 0 when there is no such
// type.
static char letter_of(int64_t code) {
	if (code < 0 || code >= CODE_MAX) {
		return 0;
	}
	return code_letters[code];
}

static ffi_type *type_of(char c) {
	switch (c) {
	case 'v': return &ffi_type_void;
	case 'b': return &ffi_type_uint8;
	case 'c': return &ffi_type_sint8;
	case 'C': return &ffi_type_uint8;
	case 's': return &ffi_type_sint16;
	case 'S': return &ffi_type_uint16;
	case 'i': return &ffi_type_sint32;
	case 'I': return &ffi_type_uint32;
	case 'l': return &ffi_type_sint64;
	case 'L': return &ffi_type_uint64;
	case 'f': return &ffi_type_float;
	case 'd': return &ffi_type_double;
	case 'p':
	case 'z': return &ffi_type_pointer;
	default:  return NULL;
	}
}

// new_native_obj prepares a call once and for all: the cif is built here and
// not at every call, which is the whole reason this object exists.
//
// What a C declaration looks like is none of its business - the types arrive
// as codes, and reading "double pow(double, double)" is the job of
// stdlib/ffi.tau.
struct object new_native_obj(void *fn, int64_t ret, const int64_t *args, size_t nargs) {
	char retc = letter_of(ret);

	if (retc == 0) {
		return errorf("cfunc: %ld is not a type", ret);
	}
	if (nargs > 255) {
		return errorf("cfunc: too many arguments, %lu", nargs);
	}

	// The prepared call is the payload of the header: variable in size,
	// because the argument types follow it.
	struct gc_header *h = gc_alloc(sizeof(struct native) + nargs * sizeof(ffi_type *));
	struct native *n = GC_PAYLOAD(h);
	char *codes = malloc(nargs + 1);

	for (size_t i = 0; i < nargs; i++) {
		char c = letter_of(args[i]);

		if (c == 0 || c == 'v') {
			free(codes);
			free(h);
			return errorf("cfunc: argument %lu: %ld is not a type an argument can have", i+1, args[i]);
		}
		codes[i] = c;
		n->types[i] = type_of(c);
	}
	codes[nargs] = '\0';

	n->fn = fn;
	n->ret = retc;
	n->nargs = nargs;
	n->args = codes;

	if (ffi_prep_cif(&n->cif, FFI_DEFAULT_ABI, nargs, type_of(retc), n->types) != FFI_OK) {
		free(n->args);
		free(h);
		return errorf("cfunc: cannot prepare a call with these types");
	}

	h->obj = (struct object) {
		.data.handle = n,
		.type = obj_native_fn,
	};
	return h->obj;
}

void dispose_native_obj(struct object o) {
	struct native *n = o.data.handle;
	free(n->args);
}

char *native_str(struct object o) {
	struct native *n = o.data.handle;
	char *s = malloc(n->nargs + 8);

	sprintf(s, "<%c(%s)>", n->ret, n->args);
	return s;
}

// The room one argument needs, whatever its C type is.
union word {
	int8_t i8;
	uint8_t u8;
	int16_t i16;
	uint16_t u16;
	int32_t i32;
	uint32_t u32;
	int64_t i64;
	uint64_t u64;
	float f;
	double d;
	void *p;
	ffi_arg word;
};

// Whatever a tau value is, an integer argument wants a number out of it.
static int64_t as_int(struct object o, int *ok) {
	switch (o.type) {
	case obj_integer:
	case obj_boolean:
	case obj_native:
		return o.data.i;
	case obj_float:
		return (int64_t) o.data.f;
	case obj_null:
		return 0;
	default:
		*ok = 0;
		return 0;
	}
}

static double as_float(struct object o, int *ok) {
	switch (o.type) {
	case obj_float:
		return o.data.f;
	case obj_integer:
	case obj_boolean:
		return (double) o.data.i;
	default:
		*ok = 0;
		return 0;
	}
}

// A pointer argument takes what already is a block of memory: a bytes buffer,
// a string, another pointer, or an integer holding an address.
static void *as_pointer(struct object o, char **copy, int *ok) {
	switch (o.type) {
	case obj_bytes:
		return o.data.bytes->bytes;
	case obj_string: {
		// A C function reads a string up to its NUL: a slice has none of its
		// own, so it travels as a copy that lives until the call returns.
		char *s = cstr(o.data.str);
		if (s != o.data.str->str) *copy = s;
		return s;
	}
	case obj_native:
		return o.data.handle;
	case obj_integer:
		return (void *) (intptr_t) o.data.i;
	case obj_null:
		return NULL;
	default:
		*ok = 0;
		return NULL;
	}
}

struct object native_call(struct object f, struct object *args, size_t nargs) {
	struct native *n = f.data.handle;

	if (nargs != n->nargs) {
		return errorf("native: wrong number of arguments, expected %u, got %lu", n->nargs, nargs);
	}

	union word vals[nargs > 0 ? nargs : 1];
	void *ptrs[nargs > 0 ? nargs : 1];
	// The copies a string argument may have needed, freed once the call is
	// over and the C function is done reading them.
	char *copies[nargs > 0 ? nargs : 1];
	size_t ncopies = 0;

	for (size_t i = 0; i < nargs; i++) {
		int ok = 1;
		char *copy = NULL;

		switch (n->args[i]) {
		case 'b': vals[i].u8 = is_truthy(&args[i]); break;
		case 'c': vals[i].i8 = as_int(args[i], &ok); break;
		case 'C': vals[i].u8 = as_int(args[i], &ok); break;
		case 's': vals[i].i16 = as_int(args[i], &ok); break;
		case 'S': vals[i].u16 = as_int(args[i], &ok); break;
		case 'i': vals[i].i32 = as_int(args[i], &ok); break;
		case 'I': vals[i].u32 = as_int(args[i], &ok); break;
		case 'l': vals[i].i64 = as_int(args[i], &ok); break;
		case 'L': vals[i].u64 = as_int(args[i], &ok); break;
		case 'f': vals[i].f = as_float(args[i], &ok); break;
		case 'd': vals[i].d = as_float(args[i], &ok); break;
		default:  vals[i].p = as_pointer(args[i], &copy, &ok); break;
		}
		if (copy != NULL) copies[ncopies++] = copy;

		if (!ok) {
			for (size_t j = 0; j < ncopies; j++) free(copies[j]);
			return errorf(
				"native: argument %lu is a %s, the signature says '%c'",
				i+1, otype_str(args[i].type), n->args[i]
			);
		}
		ptrs[i] = &vals[i];
	}

	// A native call can block for an arbitrary amount of time (sockets, IO):
	// park so that it doesn't hold back a collection. The arguments are still
	// on the stack of the caller, so a collection happening now marks them and
	// the buffers the C function is reading stay where they are.
	union word ret = {0};

	gc_park();
	ffi_call(&n->cif, n->fn, &ret, ptrs);
	gc_unpark();

	for (size_t i = 0; i < ncopies; i++) {
		free(copies[i]);
	}

	// Anything narrower than a word comes back widened to one, so it is read
	// out of the word and not out of the field of its own width.
	switch (n->ret) {
	case 'v': return null_obj;
	case 'b': return parse_bool((uint8_t) ret.word != 0);
	case 'c': return new_integer_obj((int8_t) ret.word);
	case 'C': return new_integer_obj((uint8_t) ret.word);
	case 's': return new_integer_obj((int16_t) ret.word);
	case 'S': return new_integer_obj((uint16_t) ret.word);
	case 'i': return new_integer_obj((int32_t) ret.word);
	case 'I': return new_integer_obj((uint32_t) ret.word);
	case 'l': return new_integer_obj((int64_t) ret.word);
	case 'L': return new_integer_obj((int64_t) (uint64_t) ret.word);
	case 'f': return new_float_obj(ret.f);
	case 'd': return new_float_obj(ret.d);
	case 'z': {
		// The string belongs to the library, tau takes a copy of it.
		char *s = ret.p;
		if (s == NULL) return null_obj;
		return new_string_obj(strdup(s), strlen(s));
	}
	default:
		// A pointer, kept as it is: nothing to free and nothing for the
		// collector to look after.
		return (struct object) {
			.data.handle = ret.p,
			.type = obj_native,
		};
	}
}

// ========== The other direction: a tau function C can call ==========

// A tau function C can call. libffi writes a trampoline at `code` that looks
// to C like an ordinary function pointer; calling it lands in cexport_entry
// below, with this structure as its user data.
//
// It is never freed, and neither is the tau function it holds, which stays a
// root for as long as the program runs. A callback is installed once and kept
// by whoever was given it - a widget, a sort, an event loop - and there is no
// moment when tau can know that side is done with it. Freeing a trampoline C
// still holds is the crash this deliberately cannot have.
struct cexport {
	ffi_closure *closure;
	void *code;
	ffi_cif cif;
	struct object fn;
	char ret;
	char *args;
	uint8_t nargs;
	ffi_type *types[];
};

// The handler every trampoline lands in. It converts what C passed into tau
// values, runs the function, and writes back whatever came out.
static void cexport_entry(ffi_cif *cif, void *ret, void **args, void *user) {
	struct cexport *cb = user;
	(void) cif;

	// The VM of this thread, which is the one that entered C in the first
	// place. A library that calls back from a thread of its own has no tau
	// running on it, and there is nothing here to answer with: the result is
	// left as zero rather than becoming a crash inside somebody else's loop.
	struct vm *vm = gc_current_vm();
	if (vm == NULL) {
		if (cb->ret != 'v') memset(ret, 0, sizeof(ffi_arg));
		return;
	}

	struct object objs[cb->nargs > 0 ? cb->nargs : 1];
	for (uint8_t i = 0; i < cb->nargs; i++) {
		void *p = args[i];

		switch (cb->args[i]) {
		case 'b': objs[i] = parse_bool(*(uint8_t *) p != 0); break;
		case 'c': objs[i] = new_integer_obj(*(int8_t *) p); break;
		case 'C': objs[i] = new_integer_obj(*(uint8_t *) p); break;
		case 's': objs[i] = new_integer_obj(*(int16_t *) p); break;
		case 'S': objs[i] = new_integer_obj(*(uint16_t *) p); break;
		case 'i': objs[i] = new_integer_obj(*(int32_t *) p); break;
		case 'I': objs[i] = new_integer_obj(*(uint32_t *) p); break;
		case 'l': objs[i] = new_integer_obj(*(int64_t *) p); break;
		case 'L': objs[i] = new_integer_obj((int64_t) *(uint64_t *) p); break;
		case 'f': objs[i] = new_float_obj(*(float *) p); break;
		case 'd': objs[i] = new_float_obj(*(double *) p); break;

		case 'z': {
			// A string C owns: tau takes a copy, since the one it was given
			// may be gone by the time the function looks at it again.
			char *s = *(char **) p;
			objs[i] = s == NULL ? null_obj : new_string_obj(strdup(s), strlen(s));
			break;
		}

		default:
			objs[i] = (struct object) {
				.data.handle = *(void **) p,
				.type = obj_native,
			};
			break;
		}
	}

	// Running tau again, so this thread is no longer parked: it was, for the
	// call into C that led here.
	gc_unpark();
	struct object res = vm_call_tau(vm, cb->fn, objs, cb->nargs);
	gc_park();

	if (cb->ret == 'v') {
		return;
	}

	// Anything narrower than a word is written as a word, which is what the
	// ABI expects of a result.
	int ok = 1;
	switch (cb->ret) {
	case 'b': *(ffi_arg *) ret = is_truthy(&res) ? 1 : 0; break;
	case 'c': *(ffi_arg *) ret = (int8_t) as_int(res, &ok); break;
	case 'C': *(ffi_arg *) ret = (uint8_t) as_int(res, &ok); break;
	case 's': *(ffi_arg *) ret = (int16_t) as_int(res, &ok); break;
	case 'S': *(ffi_arg *) ret = (uint16_t) as_int(res, &ok); break;
	case 'i': *(ffi_arg *) ret = (int32_t) as_int(res, &ok); break;
	case 'I': *(ffi_arg *) ret = (uint32_t) as_int(res, &ok); break;
	case 'l': *(int64_t *) ret = as_int(res, &ok); break;
	case 'L': *(uint64_t *) ret = (uint64_t) as_int(res, &ok); break;
	case 'f': *(float *) ret = (float) as_float(res, &ok); break;
	case 'd': *(double *) ret = as_float(res, &ok); break;
	default: {
		// A pointer or a string going back to C. A tau string would have to
		// outlive this call and nothing here owns it, so only an address is
		// taken: a bytes buffer, another pointer, or null.
		char *copy = NULL;
		*(void **) ret = as_pointer(res, &copy, &ok);
		break;
	}
	}

	// A function that answered with the wrong kind of value cannot be told
	// off from here: C is waiting for a number. Zero is the least surprising
	// thing to hand back.
	if (!ok) memset(ret, 0, sizeof(ffi_arg));
}

struct object new_cexport_obj(struct object fn, int64_t ret, const int64_t *args, size_t nargs) {
	char retc = letter_of(ret);

	if (fn.type != obj_closure) {
		return errorf("cexport: first argument must be a function, got %s", otype_str(fn.type));
	}
	if (retc == 0) {
		return errorf("cexport: %ld is not a type", ret);
	}
	if (nargs > 32) {
		return errorf("cexport: too many arguments, %lu", nargs);
	}

	struct cexport *cb = malloc(sizeof(struct cexport) + nargs * sizeof(ffi_type *));
	char *codes = malloc(nargs + 1);

	for (size_t i = 0; i < nargs; i++) {
		char c = letter_of(args[i]);

		if (c == 0 || c == 'v') {
			free(codes);
			free(cb);
			return errorf("cexport: argument %lu: %ld is not a type an argument can have", i+1, args[i]);
		}
		codes[i] = c;
		cb->types[i] = type_of(c);
	}
	codes[nargs] = '\0';

	cb->closure = ffi_closure_alloc(sizeof(ffi_closure), &cb->code);
	if (cb->closure == NULL) {
		free(codes);
		free(cb);
		return errorf("cexport: cannot allocate a trampoline");
	}

	cb->fn = fn;
	cb->ret = retc;
	cb->nargs = nargs;
	cb->args = codes;

	if (ffi_prep_cif(&cb->cif, FFI_DEFAULT_ABI, nargs, type_of(retc), cb->types) != FFI_OK) {
		ffi_closure_free(cb->closure);
		free(codes);
		free(cb);
		return errorf("cexport: cannot prepare a call with these types");
	}

	if (ffi_prep_closure_loc(cb->closure, &cb->cif, cexport_entry, cb, cb->code) != FFI_OK) {
		ffi_closure_free(cb->closure);
		free(codes);
		free(cb);
		return errorf("cexport: cannot prepare the trampoline");
	}

	// The function is reachable from C and from nowhere in tau, so the
	// collector is told about it here and never told otherwise.
	gc_add_roots(&cb->fn, 1);

	return (struct object) {
		.data.handle = cb->code,
		.type = obj_native,
	};
}
