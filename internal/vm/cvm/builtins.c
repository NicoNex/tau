#include <stdio.h>
#include <stdint.h>
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>

#include "obj.h"

#define BUILTIN(name) static struct object name(struct object *args, size_t len)

static struct object errorf(char *fmt, ...) {
	char *msg = malloc(sizeof(char) * 256);
	msg[255] = '\n';

	va_list ap;
	va_start(ap, fmt);
	vsnprintf(msg, 256, fmt, ap);
	va_end(ap);

	return new_error_obj(msg, strlen(msg));
}

BUILTIN(_len_b) {
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
	// case obj_bytes:
	// 	return new_integer_obj(arg.data.bytes->len);
	default:
		return errorf("len: object of type \"%s\" has no length", otype_str(arg.type));
	}
}

BUILTIN(_println_b) {
	for (uint32_t i = 0; i < len; i++) {
		char *s = object_str(args[i]);
		fputs(s, stdout);
		free(s);
		if (i < len-1) putc(' ', stdout);
	}
	putc('\n', stdout);
	return null_obj;
}

BUILTIN(_print_b) {
	for (uint32_t i = 0; i < len; i++) {
		char *s = object_str(args[i]);
		fputs(s, stdout);
		free(s);
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

	char *s = object_str(args[0]);
	return new_string_obj(s, strlen(s));
}

BUILTIN(_error_b) {
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

	char *s = otype_str(args[0].type);
	return new_string_obj(strdup(s), strlen(s));
}

BUILTIN(_int_b) {
	if (len != 1) {
		return errorf("int: wrong number of arguments, expected 1, got %lu", len);
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
	default: {
		char *s = object_str(args[0]);
		struct object err = errorf("int: %s is not a number", s);
		free(s);
		return err;
	}
	}
}

BUILTIN(_float_b) {
	if (len != 1) {
		return errorf("int: wrong number of arguments, expected 1, got %lu", len);
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
	default: {
		char *s = object_str(args[0]);
		struct object err = errorf("float: %s is not a number", s);
		free(s);
		return err;
	}
	}
}

BUILTIN(_exit_b) {
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

BUILTIN(_failed_b) {
	if (len != 1) {
		return errorf("failed: wrong number of arguments, expected 1, got %lu", len);
	}

	return parse_bool(args[0].type == obj_error);
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
	NULL, // append
	NULL, // push
	NULL, // range
	NULL, // new
	_failed_b,
	NULL, // plugin
	NULL, // pipe
	NULL, // send
	NULL, // recv
	NULL, // close
	NULL, // hex
	NULL, // oct
	NULL, // bin
	NULL, // slice
	NULL, // open
	NULL // bytes
};
