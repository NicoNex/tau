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

// The letters of a signature. Lowercase is signed, uppercase is the unsigned
// of the same width, which is the only pun in the whole notation.
//
//   v void (result only)   f float
//   b bool                 d double
//   c/C int8 uint8         p pointer, as a native value
//   s/S int16 uint16       z char *, as a tau string
//   i/I int32 uint32
//   l/L int64 uint64
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

// new_native_obj reads a signature written as "ret(args)", e.g. "d(dd)" for
// double f(double, double) or "v()" for void f(void), and prepares the call
// once and for all.
struct object new_native_obj(void *fn, char *sig) {
	char *open = strchr(sig, '(');
	size_t len = strlen(sig);

	if (open == NULL || open != sig+1 || sig[len-1] != ')') {
		return errorf("native: signature must be written as \"ret(args)\", got \"%s\"", sig);
	}
	if (type_of(sig[0]) == NULL) {
		return errorf("native: unknown type '%c' in signature \"%s\"", sig[0], sig);
	}

	char *argsig = open + 1;
	size_t nargs = len - 3;

	if (nargs > 255) {
		return errorf("native: too many arguments in signature \"%s\"", sig);
	}

	struct native *n = malloc(sizeof(struct native) + nargs * sizeof(ffi_type *));

	for (size_t i = 0; i < nargs; i++) {
		if (argsig[i] == 'v' || (n->types[i] = type_of(argsig[i])) == NULL) {
			free(n);
			return errorf("native: unknown type '%c' in signature \"%s\"", argsig[i], sig);
		}
	}

	n->fn = fn;
	n->ret = sig[0];
	n->nargs = nargs;
	n->args = strndup(argsig, nargs);

	if (ffi_prep_cif(&n->cif, FFI_DEFAULT_ABI, nargs, type_of(sig[0]), n->types) != FFI_OK) {
		free(n->args);
		free(n);
		return errorf("native: cannot prepare a call with signature \"%s\"", sig);
	}

	return (struct object) {
		.data.handle = n,
		.type = obj_native_fn,
		.gc = gc_header_alloc()
	};
}

void dispose_native_obj(struct object o) {
	struct native *n = o.data.handle;
	free(n->args);
	free(n);
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
			.gc = NULL
		};
	}
}
