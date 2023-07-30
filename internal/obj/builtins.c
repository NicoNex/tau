#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <errno.h>
#include <dlfcn.h>
#include "object.h"
#include "../vm/gc.h"

#define BUILTIN(name) static struct object name(struct object *_args, size_t len)

#define UNWRAP_ARGS()                                    \
	struct object args[len];                             \
	for (uint32_t i = 0; i < len; i++) {                 \
		if (_args[i].type == obj_getsetter) {            \
			struct getsetter *gs = _args[i].data.gs;     \
			args[i] = gs->get(gs);                       \
			continue;                                    \
		}                                                \
		args[i] = _args[i];                              \
	}

struct object new_builtin_obj(struct object (*builtin)(struct object *args, size_t len)) {
	return (struct object) {
		.data.builtin = builtin,
		.type = obj_builtin,
	};
}

BUILTIN(_len_b) {
	if (len != 1) {
		return errorf("len: wrong number of arguments, expected 1, got %lu", len);
	}
	UNWRAP_ARGS();

	struct object arg = args[0];

	switch (arg.type) {
	case obj_list:
		return new_integer_obj(arg.data.list->len);
	case obj_error:
	case obj_string:
		return new_integer_obj(arg.data.str->len);
	// case obj_bytes:
	// 	return new_integer_obj(arg.data.bytes->len);
	default:
		return errorf("len: object of type \"%s\" has no length", otype_str(arg.type));
	}
}

BUILTIN(_println_b) {
	UNWRAP_ARGS();

	for (uint32_t i = 0; i < len; i++) {
		char *s = object_str(args[i]);
		fputs(s, stdout);
		if (i < len-1) putc(' ', stdout);
	}
	putc('\n', stdout);
	return null_obj;
}

BUILTIN(_print_b) {
	UNWRAP_ARGS();

	for (uint32_t i = 0; i < len; i++) {
		char *s = object_str(args[i]);
		fputs(s, stdout);
		if (i < len-1) putc(' ', stdout);
	}
	return null_obj;
}

BUILTIN(_input_b) {
	return errorf("input: unimplemented");
}

BUILTIN(_string_b) {
	if (len != 1) {
		return errorf("string: wrong number of arguments, expected 1, got %lu", len);
	}
	UNWRAP_ARGS();

	switch (args[0].type) {
	case obj_native:
		char *s = (char *) args[0].data.handle;
		return new_string_obj(s, strlen(s));

	default: {
		char *s = object_str(args[0]);
		return new_string_obj(s, strlen(s));
	}
	}
}

BUILTIN(_error_b) {
	UNWRAP_ARGS();
	if (len != 1) {
		return errorf("error: wrong number of arguments, expected 1, got %lu", len);
	} else if (args[0].type != obj_string) {
		return errorf("error: argument must be a string, got %s", otype_str(args[0].type));
	}
	return new_error_obj(strdup(args[0].data.str->str), args[0].data.str->len);
}

BUILTIN(_type_b) {
	if (len != 1) {
		return errorf("type: wrong number of arguments, expected 1, got %lu", len);
	}
	UNWRAP_ARGS();

	char *s = otype_str(args[0].type);
	return new_string_obj(strdup(s), strlen(s));
}

BUILTIN(_int_b) {
	if (len != 1) {
		return errorf("int: wrong number of arguments, expected 1, got %lu", len);
	}
	UNWRAP_ARGS();

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
		return new_integer_obj(*(int64_t*)args[0].data.handle);

	default: {
		char *s = object_str(args[0]);
		struct object err = errorf("int: %s is not a number", s);
		return err;
	}
	}
}

BUILTIN(_float_b) {
	if (len != 1) {
		return errorf("int: wrong number of arguments, expected 1, got %lu", len);
	}
	UNWRAP_ARGS();

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
		return new_float_obj(*(double*)args[0].data.handle);

	default:
		return errorf("float: %s is not a number", object_str(args[0]));
	}
}

BUILTIN(_exit_b) {
	UNWRAP_ARGS();

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

BUILTIN(_append_b) {
	if (len < 2) {
		return errorf("append: wrong number of arguments, expected at least 2");
	}
	UNWRAP_ARGS();

	if (args[0].type != obj_list) {
		return errorf("append: first argument must be a list");
	}

	struct list *l = args[0].data.list;
	if (l->cap == 0) {
		l->list = realloc(l->list, ++l->cap * sizeof(struct object));
	}
	for (uint32_t i = 1; i < len; i++) {
		if (l->len == l->cap) {
			l->cap *= 2;
			l->list = realloc(l->list, l->cap * sizeof(struct object));
		}
		l->list[l->len++] = args[i];
	}

	return args[0];
}

BUILTIN(_new_b) {
	if (len != 0) {
		return errorf("new: wrong number of arguments, expected 0, got %lu", len);
	}
	return new_object();
}

BUILTIN(_failed_b) {
	if (len != 1) {
		return errorf("failed: wrong number of arguments, expected 1, got %lu", len);
	}
	UNWRAP_ARGS();

	return parse_bool(args[0].type == obj_error);
}

BUILTIN(_plugin_b) {
	if (len != 1) {
		return errorf("plugin: wrong number of arguments, expected 1, got %lu", len);
	}
	UNWRAP_ARGS();

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

BUILTIN(_pipe_b) {
	UNWRAP_ARGS();
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

BUILTIN(_send_b) {
	if (len != 2) {
		return errorf("send: wrong number of arguments, expected 2, got %lu", len);
	}
	UNWRAP_ARGS();

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

BUILTIN(_recv_b) {
	if (len != 1) {
		return errorf("recv: wrong number of arguments, expected 1, got %lu", len);
	}
	UNWRAP_ARGS();

	if (args[0].type != obj_pipe) {
		return errorf("recv: first argument must be a pipe, got %s instead", otype_str(args[0].type));
	}
	return pipe_recv(args[0]);
}

BUILTIN(_close_b) {
	if (len != 1) {
		return errorf("close: wrong number of arguments, expected 1, got %lu", len);
	}
	UNWRAP_ARGS();

	if (args[0].type != obj_pipe) {
		return errorf("close: first argument must be a pipe, got %s instead", otype_str(args[0].type));
	}
	if (!pipe_close(args[0])) {
		return errorf("close: pipe already closed");
	}
	return null_obj;
}

BUILTIN(_hex_b) {
	if (len != 1) {
		return errorf("hex: wrong number of arguments, expected 1, got %lu", len);
	}
	UNWRAP_ARGS();

	if (args[0].type != obj_integer) {
		return errorf("hex: first argument must be int, got %s instead", otype_str(args[0].type));
	}

	char *s = calloc(30, sizeof(char));
#ifdef __unix__
	sprintf(s, "0x%lx", args[0].data.i);
#else
	sprintf(s, "0x%llx", args[0].data.i);
#endif
	return new_string_obj(s, strlen(s));
}

BUILTIN(_oct_b) {
	if (len != 1) {
		return errorf("oct: wrong number of arguments, expected 1, got %lu", len);
	}
	UNWRAP_ARGS();

	if (args[0].type != obj_integer) {
		return errorf("oct: first argument must be int, got %s instead", otype_str(args[0].type));
	}

	char *s = calloc(30, sizeof(char));
#ifdef __unix__
	sprintf(s, "0o%lo", args[0].data.i);
#else
	sprintf(s, "0o%llo", args[0].data.i);
#endif
	return new_string_obj(s, strlen(s));
}

BUILTIN(_bin_b) {
	if (len != 1) {
		return errorf("bin: wrong number of arguments, expected 1, got %lu", len);
	}
	UNWRAP_ARGS();

	if (args[0].type != obj_integer) {
		return errorf("bin: first argument must be int, got %s instead", otype_str(args[0].type));
	}

	char *s = calloc(67, sizeof(char));
	s[0] = '0';
	s[1] = 'b';
	int idx = 2;

	for (int64_t n = args[0].data.i; n; n >>= 1) {
		s[idx++] = n & 1 ? '1' : '0';
	}
	return new_string_obj(s, strlen(s));
}

BUILTIN(_slice_b) {
	if (len != 3) {
		return errorf("slice: wrong number of arguments, expected 3, got %lu", len);
	}
	UNWRAP_ARGS();

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
			return new_string_obj(strdup(""), 0);
		}
		return new_string_obj(&args[0].data.str->str[start], end-start);
	}
	// case obj_bytes:
	default:
		return errorf("slice: first argument must be a list or string, got %s instead", otype_str(args[0].type));
	}
}

const builtin builtins[NUM_BUILTINS] = {
	_len_b,
	_println_b,
	_print_b,
	_input_b,
	_string_b,
	_error_b,
	_type_b,
	_int_b,
	_float_b,
	_exit_b,
	_append_b,
	_new_b,
	_failed_b,
	_plugin_b,
	_pipe_b,
	_send_b,
	_recv_b,
	_close_b,
	_hex_b,
	_oct_b,
	_bin_b,
	_slice_b,
	NULL, // open
	NULL // bytes
};
