#include <string.h>
#include "object.h"
#include "../vm/gc.h"

// TODO: optimise this.
char *list_str(struct object o) {
	size_t len = o.data.list->len;
	struct object *list = o.data.list->list;
	char *strings[len];
	size_t string_len = 3;

	for (int i = 0; i < len; i++) {
		char *s = object_str(list[i]);
		strings[i] = s;
		string_len += i < len-1 ? strlen(s) + 2 : strlen(s);
	}

	char *str = calloc(string_len, sizeof(char));
	str[0] = '[';

	for (int i = 0; i < len; i++) {
		strcat(str, strings[i]);
		if (i < len-1) strcat(str, ", ");
	}
	strcat(str, "]");
	return str;
}

struct object new_list_obj(struct object *list, size_t len) {
	struct list *l = malloc(sizeof(struct list));
	l->list = list;
	l->len = len;
	l->cap = len;

	return (struct object) {
		.data.list = l,
		.type = obj_list,
	};
}

struct object list_getsetter_get(struct getsetter *gs) {
	struct object *list = gs->l.data.list->list;
	size_t listlen = gs->l.data.list->len;
	int64_t idx = gs->r.data.i;

	if (idx < 0 || idx >= listlen) {
		return new_error_obj(strdup("index out of range"), 18);
	}
	return list[idx];
}

struct object list_getsetter_set(struct getsetter *gs, struct object val) {
	struct object *list = gs->l.data.list->list;
	size_t listlen = gs->l.data.list->len;
	int64_t idx = gs->r.data.i;

	if (idx < 0 || idx >= listlen) {
		return new_error_obj(strdup("index out of range"), 18);
	}
	list[idx] = val;
	return val;
}

inline struct list new_list(size_t cap) {
	return (struct list) {
		.list = malloc(sizeof(struct object) * cap),
		.cap = cap,
		.len = 0
	};
}

inline void list_insert(struct list *l, struct object o, size_t idx) {
	if (idx >= l->cap) {
		if (l->cap == 0) l->cap = 1;
		while (l->cap <= idx) l->cap *= 2;
		l->list = realloc(l->list, sizeof(struct object) * l->cap);
	}
	l->list[idx] = o;
	l->len++;
}

inline struct list list_copy(struct list l) {
	struct list ret = {
		.list = malloc(sizeof(struct object) * l.cap),
		.cap = l.cap,
		.len = l.len
	};
	memcpy(ret.list, l.list, l.cap);

	return ret;
}
