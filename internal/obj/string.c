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
	// Free everything if it's not a slice (marked parent bit is set to NULL).
	if (o.data.str->m_parent == NULL) {
		free(o.marked);
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
	s->m_parent = NULL;

	return (struct object) {
		.data.str = s,
		.type = obj_string,
		.marked = MARKPTR(),
	};
}

void mark_string_obj(struct object s) {
	*s.marked = 1;
	if (s.data.str->m_parent != NULL) {
		*s.data.str->m_parent = 1;
	}
}

struct object new_string_slice(char *str, size_t len, uint32_t *m_parent) {
	struct string *s = malloc(sizeof(struct string));
	s->str = str;
	s->len = len;
	s->m_parent = m_parent;

	return (struct object) {
		.data.str = s,
		.type = obj_string,
		.marked = MARKPTR(),
	};
}
