#include <stdlib.h>
#include <string.h>
#include <stdio.h>
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
		return strdup("<corrupted>");
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
		case obj_bytes:
			mark_bytes_obj(o);
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
		dispose_bytes_obj(o);
		return;
	case obj_native:
		free(o.marked);
		free(o.data.handle);
		return;
	default:
		return;
	}
}
