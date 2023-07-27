#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <stdarg.h>
#include "object.h"

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
		"getsetter",
		"native"
	};

	return strings[t];
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
		return strdup("<unimplemented bytes>");
	case obj_getsetter:
		return getsetter_str(o);
	case obj_native:
		return strdup("<native>");
	default:
		return strdup("<unimplemented>");
	}
}

void print_obj(struct object o) {
	char *str = object_str(o);
	puts(str);
	free(str);
}

inline void mark_obj(struct object o) {
	if (o.type > obj_builtin) {
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
		case obj_pipe:
			mark_pipe_obj(o);
			break;
		default:
			*o.marked = 1;
			break;
		}
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
		puts("no free function for bytes");
		return;
	case obj_getsetter:
		dispose_getsetter_obj(o);
		return;
	case obj_native:
		dispose_native_obj(o);
		return;
	default:
		return;
	}
}

inline struct object errorf(char *fmt, ...) {
	char *msg = malloc(sizeof(char) * 256);
	msg[255] = '\n';

	va_list ap;
	va_start(ap, fmt);
	vsnprintf(msg, 256, fmt, ap);
	va_end(ap);

	return new_error_obj(msg, strlen(msg));
}
