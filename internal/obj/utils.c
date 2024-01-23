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
		"native"
	};
	return t <= obj_native ? strings[t] : "corrupted";
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
		return bytes_str(o);
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
