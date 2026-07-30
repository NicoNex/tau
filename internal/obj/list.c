#include <stdlib.h>
#include <string.h>
#include "object.h"

void dispose_list_obj(struct object o) {
	// A slice doesn't own the buffer, its owner frees it.
	if (o.data.list->owner == NULL) {
		free(o.data.list->list);
	}
	free(o.data.list);
}

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
		free(strings[i]);
	}
	strcat(str, "]");

	return str;
}

void mark_list_obj(struct object l) {
	l.gc->mark |= GC_MARK;
	mark_owner(l.data.list->owner);
	for (uint32_t i = 0; i < l.data.list->len; i++) {
		mark_obj(l.data.list->list[i]);
	}
}

struct object make_list(size_t cap) {
	struct list *l = malloc(sizeof(struct list));
	l->list = calloc(cap, sizeof(struct object));
	l->len = 0;
	l->cap = cap;
	l->owner = NULL;

	return (struct object) {
		.data.list = l,
		.type = obj_list,
		.gc = gc_header_alloc()
	};
}

struct object new_list_obj(struct object *list, size_t len) {
	struct list *l = malloc(sizeof(struct list));
	l->list = list;
	l->len = len;
	l->cap = len;
	l->owner = NULL;

	return (struct object) {
		.data.list = l,
		.type = obj_list,
		.gc = gc_header_alloc()
	};
}

struct object new_list_obj_data(struct object *list, size_t len, size_t cap) {
	struct list *l = malloc(sizeof(struct list));
	l->list = list;
	l->len = len;
	l->cap = cap;
	l->owner = NULL;

	return (struct object) {
		.data.list = l,
		.type = obj_list,
		.gc = gc_header_alloc()
	};
}

struct object new_list_slice(struct object *list, size_t len, struct gc_header *owner) {
	struct list *l = malloc(sizeof(struct list));
	l->list = list;
	l->len = len;
	l->cap = len;
	l->owner = owner;

	return (struct object) {
		.data.list = l,
		.type = obj_list,
		.gc = gc_header_alloc(),
	};
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
