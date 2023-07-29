#include <stdio.h>
#include <stdarg.h>
#include <string.h>
#include "object.h"
#include "../vm/gc.h"

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
		return "null";
	case obj_boolean:
		return boolean_str(o);
	case obj_integer:
		return integer_str(o);
	case obj_float:
		return float_str(o);
	case obj_builtin:
		return "<builtin function>";
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
		return "<pipe>";
	case obj_bytes:
		return "<unimplemented bytes>";
	case obj_getsetter:
		return getsetter_str(o);
	case obj_native:
		return "<native>";
	default:
		return "<unimplemented>";
	}
}

void print_obj(struct object o) {
	puts(object_str(o));
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
