#include <stdlib.h>
#include <string.h>
#include "object.h"

#if defined(_WIN32) || defined(WIN32)
	char *strndup(char * restrict s, size_t len) {
		char *dup = malloc(sizeof(char) * len + 1);
		dup[len] = '\0';
		memcpy(dup, s, sizeof(char) * len);

		return dup;
	}
#endif

void dispose_string_obj(struct object o) {
	// A slice doesn't own the buffer, its owner frees it.
	if (o.data.str->owner == NULL) {
		free(o.data.str->str);
	}
	free(o.data.str);
}

char *string_str(struct object o) {
	return strndup(o.data.str->str, o.data.str->len);
}

struct object new_string_obj(char *str, size_t len) {
	struct string *s = malloc(sizeof(struct string));
	s->str = str;
	s->len = len;
	s->owner = NULL;

	return (struct object) {
		.data.str = s,
		.type = obj_string,
		.gc = gc_header_alloc(),
	};
}

void mark_string_obj(struct object s) {
	s.gc->mark |= GC_MARK;
	if (s.data.str->owner != NULL) {
		s.data.str->owner->mark |= GC_MARK;
	}
}

struct object new_string_slice(char *str, size_t len, struct gc_header *owner) {
	struct string *s = malloc(sizeof(struct string));
	s->str = str;
	s->len = len;
	s->owner = owner;

	return (struct object) {
		.data.str = s,
		.type = obj_string,
		.gc = gc_header_alloc(),
	};
}
