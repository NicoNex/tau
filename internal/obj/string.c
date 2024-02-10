#include "object.h"
#include "../vm/gc.h"

#if defined(_WIN32) || defined(WIN32)
	char *GC_STRNDUP(char * restrict s, size_t len) {
		char *dup = GC_MALLOC(sizeof(char) * len + 1);
		dup[len] = '\0';
		memcpy(dup, s, sizeof(char) * len);

		return dup;
	}
#endif

char *string_str(struct object o) {
	return GC_STRNDUP(o.data.str->str, o.data.str->len);
}

struct object new_string_obj(char *str, size_t len) {
	struct string *s = GC_MALLOC(sizeof(struct string));
	s->str = str;
	s->len = len;

	return (struct object) {
		.data.str = s,
		.type = obj_string,
	};
}
