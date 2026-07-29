#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
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

// The codes a signature is kept in, one per type. They are internal: what a
// signature is written in is C.
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

// The code of an integer type of a given width, so that the C types whose
// width depends on the machine are read with sizeof rather than assumed.
static char int_code(size_t width, int is_unsigned) {
	switch (width) {
	case 1:  return is_unsigned ? 'C' : 'c';
	case 2:  return is_unsigned ? 'S' : 's';
	case 4:  return is_unsigned ? 'I' : 'i';
	default: return is_unsigned ? 'L' : 'l';
	}
}

// The type names a signature may use: the ones C writes, the exact width ones
// of stdint.h, and the same widths spelled the way tau spells its own types.
// Everything is a name here, so a signature can be pasted out of a header.
static char code_of_name(const char *t) {
	if (!strcmp(t, "void"))                     return 'v';
	if (!strcmp(t, "bool") || !strcmp(t, "_Bool")) return 'b';
	if (!strcmp(t, "float"))                    return 'f';
	if (!strcmp(t, "double"))                   return 'd';
	if (!strcmp(t, "string"))                   return 'z';
	if (!strcmp(t, "pointer"))                  return 'p';

	// The widths C leaves to the machine.
	if (!strcmp(t, "char"))                     return int_code(sizeof(char), 0);
	if (!strcmp(t, "signed char"))              return int_code(sizeof(char), 0);
	if (!strcmp(t, "unsigned char"))            return int_code(sizeof(char), 1);
	if (!strcmp(t, "short") || !strcmp(t, "short int"))
		return int_code(sizeof(short), 0);
	if (!strcmp(t, "unsigned short") || !strcmp(t, "unsigned short int"))
		return int_code(sizeof(short), 1);
	if (!strcmp(t, "int") || !strcmp(t, "signed") || !strcmp(t, "signed int"))
		return int_code(sizeof(int), 0);
	if (!strcmp(t, "unsigned") || !strcmp(t, "unsigned int"))
		return int_code(sizeof(int), 1);
	if (!strcmp(t, "long") || !strcmp(t, "long int"))
		return int_code(sizeof(long), 0);
	if (!strcmp(t, "unsigned long") || !strcmp(t, "unsigned long int"))
		return int_code(sizeof(long), 1);
	if (!strcmp(t, "long long") || !strcmp(t, "long long int"))
		return int_code(sizeof(long long), 0);
	if (!strcmp(t, "unsigned long long") || !strcmp(t, "unsigned long long int"))
		return int_code(sizeof(long long), 1);
	if (!strcmp(t, "size_t"))                   return int_code(sizeof(size_t), 1);
	if (!strcmp(t, "ssize_t"))                  return int_code(sizeof(size_t), 0);
	if (!strcmp(t, "intptr_t"))                 return int_code(sizeof(void *), 0);
	if (!strcmp(t, "uintptr_t"))                return int_code(sizeof(void *), 1);

	// The widths that are the width whatever the machine is.
	if (!strcmp(t, "int8_t")   || !strcmp(t, "int8"))   return 'c';
	if (!strcmp(t, "uint8_t")  || !strcmp(t, "uint8"))  return 'C';
	if (!strcmp(t, "int16_t")  || !strcmp(t, "int16"))  return 's';
	if (!strcmp(t, "uint16_t") || !strcmp(t, "uint16")) return 'S';
	if (!strcmp(t, "int32_t")  || !strcmp(t, "int32"))  return 'i';
	if (!strcmp(t, "uint32_t") || !strcmp(t, "uint32")) return 'I';
	if (!strcmp(t, "int64_t")  || !strcmp(t, "int64"))  return 'l';
	if (!strcmp(t, "uint64_t") || !strcmp(t, "uint64")) return 'L';
	if (!strcmp(t, "float32"))                          return 'f';
	if (!strcmp(t, "float64"))                          return 'd';

	return 0;
}

// A word of a declaration that says nothing about the type it is made of.
static int is_qualifier(const char *w) {
	return !strcmp(w, "const") || !strcmp(w, "volatile") || !strcmp(w, "restrict")
		|| !strcmp(w, "__restrict") || !strcmp(w, "extern") || !strcmp(w, "static");
}

// code_of_decl reads one declaration - a return type, or a parameter, with or
// without the name that follows it - and gives back the code of its type.
//
// The name is dropped: in "char *buf" the type is char * and buf is a word
// only a person needs. A star anywhere makes it a pointer, and a char * is
// the one pointer that travels as a tau string.
static char code_of_decl(const char *decl, size_t len, char *err, size_t errlen) {
	char words[8][32];
	int nwords = 0, stars = 0;
	size_t i = 0;

	while (i < len) {
		char c = decl[i];

		if (c == ' ' || c == '\t' || c == '\n') { i++; continue; }
		if (c == '*') { stars++; i++; continue; }
		if (c == '[') {
			// An array parameter is a pointer, and what follows the bracket
			// says nothing about the call.
			stars++;
			while (i < len && decl[i] != ']') i++;
			i++;
			continue;
		}

		if (!isalnum((unsigned char) c) && c != '_') {
			snprintf(err, errlen, "'%c' has no place in a type", c);
			return 0;
		}

		size_t start = i;
		while (i < len && (isalnum((unsigned char) decl[i]) || decl[i] == '_')) i++;

		size_t wlen = i - start;
		if (wlen >= sizeof(words[0])) {
			snprintf(err, errlen, "\"%.*s\" is not a type", (int) wlen, decl + start);
			return 0;
		}
		if (nwords == 8) {
			snprintf(err, errlen, "too many words for one type");
			return 0;
		}

		memcpy(words[nwords], decl + start, wlen);
		words[nwords][wlen] = '\0';
		if (!is_qualifier(words[nwords])) nwords++;
	}

	if (nwords == 0) {
		snprintf(err, errlen, "there is no type here");
		return 0;
	}

	// The last word is the name of the thing being declared unless it is
	// part of the type: "unsigned long" keeps both words, "unsigned long n"
	// drops the third.
	char joined[8 * 32];
	int used = nwords;
	for (;;) {
		joined[0] = '\0';
		for (int w = 0; w < used; w++) {
			if (w > 0) strcat(joined, " ");
			strcat(joined, words[w]);
		}

		if (code_of_name(joined) != 0) break;
		if (used == 1) {
			snprintf(err, errlen, "\"%s\" is not a type this understands", joined);
			return 0;
		}
		used--;
	}

	char code = code_of_name(joined);

	if (stars > 0) {
		// char * is a string, everything else is an address.
		if (stars == 1 && (!strcmp(joined, "char") || !strcmp(joined, "signed char"))) {
			return 'z';
		}
		return 'p';
	}
	if (code == 'z') return 'z';
	return code;
}

// new_native_obj reads the signature of a C function and prepares the call
// once and for all. It is written the way C writes it, and the name of the
// function and the names of the arguments may be there or not:
//
//	native(libm.pow, "double(double, double)")
//	native(libc.snprintf, "int snprintf(char *buf, size_t n, const char *fmt, double x)")
//	native(libc.fflush, "void(void *)")
//
// The exact width names of stdint.h work as well, spelled either way:
// uint64_t and uint64, float64 and double.
struct object new_native_obj(void *fn, char *sig) {
	char err[128];

	// Room around the declaration is room, here as it is in C.
	while (isspace((unsigned char) *sig)) sig++;
	size_t len = strlen(sig);
	while (len > 0 && isspace((unsigned char) sig[len-1])) len--;

	char *open = memchr(sig, '(', len);

	if (open == NULL || len == 0 || sig[len-1] != ')') {
		return errorf(
			"native: a signature is a C declaration, \"double(double, double)\", got \"%s\"",
			sig
		);
	}
	if (memchr(sig, '.', len) != NULL) {
		return errorf(
			"native: \"%.*s\" is variadic, say the types of the arguments this call passes instead",
			(int) len, sig
		);
	}

	char ret = code_of_decl(sig, open - sig, err, sizeof(err));
	if (ret == 0) {
		return errorf("native: in the result of \"%.*s\": %s", (int) len, sig, err);
	}

	// The arguments, one declaration per comma. "void" on its own is C for
	// none at all, and so is nothing.
	char *argsig = open + 1;
	size_t arglen = (sig + len - 1) - argsig;
	char codes[256];
	size_t nargs = 0;

	for (size_t i = 0, start = 0; i <= arglen; i++) {
		if (i < arglen && sig[argsig - sig + i] != ',') continue;

		size_t plen = i - start;
		char code = code_of_decl(argsig + start, plen, err, sizeof(err));

		// Nothing between the parentheses at all.
		if (code == 0 && nargs == 0 && i == arglen && plen == 0) break;

		if (code == 0) {
			return errorf("native: in argument %lu of \"%.*s\": %s", nargs+1, (int) len, sig, err);
		}
		if (code == 'v') {
			if (nargs == 0 && i == arglen) break; // f(void)
			return errorf("native: argument %lu of \"%.*s\" cannot be void", nargs+1, (int) len, sig);
		}
		if (nargs == 255) {
			return errorf("native: too many arguments in \"%.*s\"", (int) len, sig);
		}

		codes[nargs++] = code;
		start = i + 1;
	}

	struct native *n = malloc(sizeof(struct native) + nargs * sizeof(ffi_type *));

	for (size_t i = 0; i < nargs; i++) {
		n->types[i] = type_of(codes[i]);
	}

	n->fn = fn;
	n->ret = ret;
	n->nargs = nargs;
	n->args = strndup(codes, nargs);

	if (ffi_prep_cif(&n->cif, FFI_DEFAULT_ABI, nargs, type_of(ret), n->types) != FFI_OK) {
		free(n->args);
		free(n);
		return errorf("native: cannot prepare a call with signature \"%.*s\"", (int) len, sig);
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
