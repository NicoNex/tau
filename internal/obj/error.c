#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include "object.h"

void dispose_error_obj(struct object o) {
	free(o.data.str->str);
}

char *error_str(struct object o) {
	return strndup(o.data.str->str, o.data.str->len);
}

struct object new_error_obj(char *str, size_t len) {
	struct gc_header *h = gc_alloc(sizeof(struct string));
	struct string *s = GC_PAYLOAD(h);
	s->str = str;
	s->len = len;
	s->owner = NULL;

	h->obj = (struct object) {
		.data.str = s,
		.type = obj_error,
	};
	return h->obj;
}

inline struct object errorf(char *fmt, ...) {
	char *msg = malloc(sizeof(char) * 256);
	msg[255] = '\0';

	va_list ap;
	va_start(ap, fmt);
	vsnprintf(msg, 256, fmt, ap);
	va_end(ap);

	return new_error_obj(msg, strlen(msg));
}
