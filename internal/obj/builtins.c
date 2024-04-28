#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <errno.h>
#include <math.h>
#include "plugin.h"
#include "object.h"
#include "../vm/gc.h"

struct object new_builtin_obj(struct object (*builtin)(struct object *args, size_t len)) {
	return (struct object) {
		.data.builtin = builtin,
		.type = obj_builtin,
	};
}

static struct object len_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("len: wrong number of arguments, expected 1, got %lu", len);
	}

	struct object arg = args[0];

	switch (arg.type) {
	case obj_list:
		return new_integer_obj(arg.data.list->len);
	case obj_error:
	case obj_string:
		return new_integer_obj(arg.data.str->len);
	case obj_bytes:
		return new_integer_obj(arg.data.bytes->len);
	default:
		return errorf("len: object of type \"%s\" has no length", otype_str(arg.type));
	}
}

static struct object println_b(struct object *args, size_t len) {

	for (uint32_t i = 0; i < len; i++) {
		char *s = object_str(args[i]);
		fputs(s, stdout);
		if (i < len-1) putc(' ', stdout);
	}
	putc('\n', stdout);
	return null_obj;
}

static struct object print_b(struct object *args, size_t len) {

	for (uint32_t i = 0; i < len; i++) {
		char *s = object_str(args[i]);
		fputs(s, stdout);
		if (i < len-1) putc(' ', stdout);
	}
	return null_obj;
}

static struct object input_b(struct object *args, size_t len) {
	if (len == 1) {
		if (args[0].type != obj_string) {
			return errorf("input: argument must be a string, got %s", otype_str(args[0].type));
		}
		fputs(args[0].data.str->str, stdout);
	}

	char tmp;
	char *input = NULL;
	size_t ilen = 0;

	do {
		tmp = getchar();
		char *reinput = GC_REALLOC(input, ilen + 1);
		if (reinput == NULL) {
			return errorf("input: error allocating memory");
		}
		input = reinput;
		input[ilen++] = tmp;

	} while (tmp != '\n' && tmp != '\0');
	input[ilen - 1] = '\0';

	return new_string_obj(input, ilen);
}

static struct object string_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("string: wrong number of arguments, expected 1, got %lu", len);
	}

	switch (args[0].type) {
	case obj_native: {
		char *s = (char *) args[0].data.handle;
		return new_string_obj(s, strlen(s));
	}

	default: {
		char *s = object_str(args[0]);
		return new_string_obj(s, strlen(s));
	}
	}
}

static struct object error_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("error: wrong number of arguments, expected 1, got %lu", len);
	} else if (args[0].type != obj_string) {
		return errorf("error: argument must be a string, got %s", otype_str(args[0].type));
	}
	return new_error_obj(GC_STRDUP(args[0].data.str->str), args[0].data.str->len);
}

static struct object type_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("type: wrong number of arguments, expected 1, got %lu", len);
	}

	char *s = otype_str(args[0].type);
	return new_string_obj(GC_STRDUP(s), strlen(s));
}

static struct object int_b(struct object *args, size_t len) {
	if (len != 1 && len != 2) {
		return errorf("int: wrong number of arguments, expected 1 or 2, got %lu", len);
	}

	switch (args[0].type) {
	case obj_integer:
		return args[0];

	case obj_float:
		args[0].data.i = (int64_t) args[0].data.f;
		args[0].type = obj_integer;
		return args[0];

	case obj_string: {
		errno = 0;
		int64_t i = strtol(args[0].data.str->str, NULL, 10);
		if (errno != EINVAL && errno != ERANGE) {
			return new_integer_obj(i);
		}
	}

	case obj_native:
		if (len == 1) {
			return new_integer_obj(*(int32_t*)args[0].data.handle);
		} else if (args[1].type != obj_integer) {
			return errorf("int: second argument must be an integer");
		}

		switch (args[1].data.i) {
		case 0:  return new_integer_obj(*(int32_t*)args[0].data.handle);
		case 8:  return new_integer_obj(*(int8_t*)args[0].data.handle);
		case 16: return new_integer_obj(*(int16_t*)args[0].data.handle);
		case 32: return new_integer_obj(*(int32_t*)args[0].data.handle);
		case 64: return new_integer_obj(*(int64_t*)args[0].data.handle);
		default: return errorf("int: invalid bit size, must be a power of 2 and not exceed 64, got %lld", args[1].data.i);
		}

	default: {
		char *s = object_str(args[0]);
		struct object err = errorf("int: %s is not a number", s);
		return err;
	}
	}
}

static struct object float_b(struct object *args, size_t len) {
	if (len != 1 && len != 2) {
		return errorf("float: wrong number of arguments, expected 1 or 2, got %lu", len);
	}

	switch (args[0].type) {
	case obj_integer:
		args[0].data.f = (double) args[0].data.i;
		args[0].type = obj_float;
		return args[0];

	case obj_float:
		return args[0];

	case obj_string: {
		errno = 0;
		double f = strtod(args[0].data.str->str, NULL);
		if (errno != ERANGE) {
			return new_float_obj(f);
		}
	}

	case obj_native:
		if (len == 1) {
			return new_float_obj(*(double*)args[0].data.handle);
		} else if (args[1].type != obj_integer) {
			return errorf("float: second argument must be an integer");
		}

		switch (args[1].data.i) {
		case 0:  return new_float_obj(*(float*)args[0].data.handle);
		case 32: return new_float_obj(*(float*)args[0].data.handle);
		case 64: return new_float_obj(*(double*)args[0].data.handle);
		default: return errorf("float: invalid bit size, must be either 0, 32 or 64, got %lld", args[1].data.i);
		}

	default:
		return errorf("float: %s is not a number", object_str(args[0]));
	}
}

static struct object exit_b(struct object *args, size_t len) {

	switch (len) {
	case 0:
		exit(0);

	case 1:
		switch (args[0].type) {
		case obj_integer:
			exit(args[0].data.i);
		case obj_string:
		case obj_error:
			puts(args[0].data.str->str);
			exit(0);
		default:
			return errorf("exit: argument must be an integer, string or error");
		}

	case 2:
		if (args[0].type != obj_string) {
			return errorf("exit: first argument must be a string");
		}
		if (args[1].type != obj_integer) {
			return errorf("exit: second argument must be an int");
		}

		puts(args[0].data.str->str);
		exit(args[1].data.i);

	default:
		return errorf("exit: wrong number of arguments, max 2, got %lu", len);
	}
}

static struct object append_b(struct object *args, size_t len) {
	if (len < 2) {
		return errorf("append: wrong number of arguments, expected at least 2");
	}

	if (args[0].type != obj_list) {
		return errorf("append: first argument must be a list");
	}

	struct list *l = args[0].data.list;

	// If there isn't enough space in the list allocate the closest power of two
	// able to contain both the previous objects and the new ones.
	if (l->cap - l->len < len - 1) {
		l->cap = pow(2, ceil(log2(l->cap + (len - 1))));
		l->list = GC_REALLOC(l->list, l->cap * sizeof(struct object));
	}

	// Copy the objects in the argument to the list.
	for (size_t i = 1; i < len; i++) {
		l->list[l->len++] = args[i];
	}

	return args[0];
}

static struct object new_b(struct object *args, size_t len) {
	if (len != 0) {
		return errorf("new: wrong number of arguments, expected 0, got %lu", len);
	}
	return new_object();
}

static struct object failed_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("failed: wrong number of arguments, expected 1, got %lu", len);
	}

	return parse_bool(args[0].type == obj_error);
}

static struct object plugin_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("plugin: wrong number of arguments, expected 1, got %lu", len);
	}

	if (args[0].type != obj_string) {
		return errorf("plugin: first argument must be string, got %s instead", otype_str(args[0].type));
	}
	char *path = args[0].data.str->str;
	void *handle = dlopen(path, RTLD_LAZY);
	if (!handle) {
		return errorf("plugin: %s", dlerror());
	}

	return (struct object) {
		.data.handle = handle,
		.type = obj_native,
	};
}

static struct object pipe_b(struct object *args, size_t len) {
	switch (len) {
	case 0:
		return new_pipe();
	case 1:
		if (args[0].type != obj_integer) {
			return errorf("pipe: first argument must be int, got %s instead", otype_str(args[0].type));
		}
		if (args[0].data.i < 0) {
			return errorf("pipe: invalid argument: size %ld, must not be negative", args[0].data.i);
		}
		return new_buffered_pipe(args[0].data.i);
	default:
		return errorf("pipe: wrong number of arguments, expected 0 or 1, got %lu", len);
	}
}

static struct object send_b(struct object *args, size_t len) {
	if (len != 2) {
		return errorf("send: wrong number of arguments, expected 2, got %lu", len);
	}

	struct object pipe = args[0];
	struct object o = args[1];

	if (pipe.type != obj_pipe) {
		return errorf("send: first argument must be a pipe, got %s instead", otype_str(args[0].type));
	}
	if (!pipe_send(pipe, o)) {
		return errorf("send: closed pipe");
	}
	return o;
}

static struct object recv_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("recv: wrong number of arguments, expected 1, got %lu", len);
	}

	if (args[0].type != obj_pipe) {
		return errorf("recv: first argument must be a pipe, got %s instead", otype_str(args[0].type));
	}
	return pipe_recv(args[0]);
}

static struct object close_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("close: wrong number of arguments, expected 1, got %lu", len);
	}

	if (args[0].type != obj_pipe) {
		return errorf("close: first argument must be a pipe, got %s instead", otype_str(args[0].type));
	}
	if (!pipe_close(args[0])) {
		return errorf("close: pipe already closed");
	}
	return null_obj;
}

static struct object hex_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("hex: wrong number of arguments, expected 1, got %lu", len);
	}

	if (args[0].type != obj_integer) {
		return errorf("hex: first argument must be int, got %s instead", otype_str(args[0].type));
	}

	char *s = GC_CALLOC(30, sizeof(char));
#ifdef __unix__
	sprintf(s, "0x%lx", args[0].data.i);
#else
	sprintf(s, "0x%llx", args[0].data.i);
#endif
	return new_string_obj(s, strlen(s));
}

static struct object oct_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("oct: wrong number of arguments, expected 1, got %lu", len);
	}

	if (args[0].type != obj_integer) {
		return errorf("oct: first argument must be int, got %s instead", otype_str(args[0].type));
	}

	char *s = GC_CALLOC(30, sizeof(char));
#ifdef __unix__
	sprintf(s, "0o%lo", args[0].data.i);
#else
	sprintf(s, "0o%llo", args[0].data.i);
#endif
	return new_string_obj(s, strlen(s));
}

static struct object bin_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("bin: wrong number of arguments, expected 1, got %lu", len);
	}

	if (args[0].type != obj_integer) {
		return errorf("bin: first argument must be int, got %s instead", otype_str(args[0].type));
	}

	char *s = GC_CALLOC(67, sizeof(char));
	s[0] = '0';
	s[1] = 'b';
	int idx = 2;

	for (int64_t n = args[0].data.i; n; n >>= 1) {
		s[idx++] = n & 1 ? '1' : '0';
	}
	return new_string_obj(s, strlen(s));
}

static struct object slice_b(struct object *args, size_t len) {
	if (len != 3) {
		return errorf("slice: wrong number of arguments, expected 3, got %lu", len);
	}

	if (args[1].type != obj_integer) {
		return errorf("slice: second argument must be an int, got %s instead", otype_str(args[1].type));
	} else if (args[2].type != obj_integer) {
		return errorf("slice: third argument must be an int, got %s instead", otype_str(args[2].type));
	}

	int64_t start = args[1].data.i;
	int64_t end = args[2].data.i;
	if (start < 0 || end < 0) {
		return errorf("slice: invalid argument: index arguments must not be negative");
	} else if (end < start) {
		return errorf("slice: invalid slice indices: %ld < %ld", end, start);
	}

	switch (args[0].type) {
	case obj_list: {
		if (end > args[0].data.list->len) {
			return errorf("slice: list bounds out of range %d with capacity %lu", end, args[0].data.list->len);
		} else if (start == end) {
			return new_list_obj(NULL, 0);
		}
		return new_list_obj(&args[0].data.list->list[start], end-start);
	}

	case obj_string: {
		if (end > args[0].data.str->len) {
			return errorf("slice: string bounds out of range %d with capacity %lu", end, args[0].data.list->len);
		} else if (start == end) {
			return new_string_obj(GC_STRDUP(""), 0);
		}
		return new_string_obj(&args[0].data.str->str[start], end-start);
	}
	case obj_bytes: {
		if (end > args[0].data.bytes->len) {
			return errorf("slice: bytes bounds out of range %d with capacity %lu", end, args[0].data.list->len);
		} else if (start == end) {
			return new_bytes_obj(NULL, 0);
		}
		return new_bytes_slice(&args[0].data.bytes->bytes[start], end-start);
	}
	default:
		return errorf("slice: first argument must be a list or string, got %s instead", otype_str(args[0].type));
	}
}

static struct object keys_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("keys: wrong number of arguments, expected 1, got %lu", len);
	} else if (args[0].type != obj_map) {
		return errorf("keys: argument must be a map, got %s instead", otype_str(args[0].type));
	}
	return map_keys(args[0]);
}

static struct object delete_b(struct object *args, size_t len) {
	if (len != 2) {
		return errorf("delete: wrong number of arguments, expected 2, got %lu", len);
	} else if (args[0].type != obj_map) {
		return errorf("delete: first argument must be a map, got %s instead", otype_str(args[0].type));
	}

	switch (args[1].type) {
	case obj_boolean:
	case obj_integer:
	case obj_float:
	case obj_string:
	case obj_error:
		map_delete(args[0], args[1]);
		return null_obj;
	default:
		return errorf("delete: second argument must be one of boolean integer float string error, got %s instead", otype_str(args[1].type));
	}
}

// TODO: eventually add the obj_integer case like in Python.
static struct object bytes_b(struct object *args, size_t len) {
	if (len != 1) {
		return errorf("bytes: wrong number of arguments, expected 1, got %lu", len);
	}

	struct object arg = args[0];
	switch (arg.type) {
	case obj_string:
		return new_bytes_slice(arg.data.str->str, arg.data.str->len);
	case obj_list: {
		size_t len = arg.data.list->len;
		struct object *list = arg.data.list->list;
		uint8_t *b = GC_MALLOC(sizeof(uint8_t) * len);

		for (uint32_t i = 0; i < len; i++) {
			if (list[i].type != obj_integer) {
				return errorf("bytes: list cannot be converted to bytes");
			}
			b[i] = list[i].data.i;
		}
		return new_bytes_obj(b, len);
	}
	default:
		return errorf("bytes: %s cannot be converted to bytes", otype_str(arg.type));
	}
}

const builtin builtins[NUM_BUILTINS] = {
	len_b,
	println_b,
	print_b,
	input_b,
	string_b,
	error_b,
	type_b,
	int_b,
	float_b,
	exit_b,
	append_b,
	new_b,
	failed_b,
	plugin_b,
	pipe_b,
	send_b,
	recv_b,
	close_b,
	hex_b,
	oct_b,
	bin_b,
	slice_b,
	keys_b,
	delete_b,
	bytes_b
};
