#include <stdio.h>
#include <string.h>
#include <stdarg.h>
#include "object.h"
#include "../vm/gc.h"

char *error_str(struct object o) {
	return GC_STRDUP(o.data.str->str);
}

struct object new_error_obj(char *str, size_t len) {
	struct string *s = GC_MALLOC(sizeof(struct string));
	s->str = str;
	s->len = len;

	return (struct object) {
		.data.str = s,
		.type = obj_error,
	};
}

inline struct object errorf(char *fmt, ...) {
	char *msg = GC_MALLOC(sizeof(char) * 256);
	msg[255] = '\0';

	va_list ap;
	va_start(ap, fmt);
	vsnprintf(msg, 256, fmt, ap);
	va_end(ap);

	return new_error_obj(msg, strlen(msg));
}
