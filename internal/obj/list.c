#include <stdlib.h>
#include <string.h>
#include "object.h"

void dispose_list_obj(struct object o) {
	// A slice doesn't own the buffer, its owner frees it.
	if (o.data.list->owner == NULL) {
		free(o.data.list->list);
	}
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
	obj_gc(l)->mark |= GC_MARK;
	mark_owner(l.data.list->owner);
	for (uint32_t i = 0; i < l.data.list->len; i++) {
		mark_obj(l.data.list->list[i]);
	}
}

// The one place a list object is made: the header and the descriptor after
// it in a single block, with the array of elements left where it is.
static struct object list_new(struct object *list, size_t len, size_t cap, struct gc_header *owner) {
	struct gc_header *h = gc_alloc(sizeof(struct list));
	struct list *l = GC_PAYLOAD(h);
	l->list = list;
	l->len = len;
	l->cap = cap;
	l->owner = owner;

	h->obj = (struct object) {
		.data.list = l,
		.type = obj_list,
	};
	return h->obj;
}

struct object make_list(size_t cap) {
	return list_new(calloc(cap, sizeof(struct object)), 0, cap, NULL);
}

struct object new_list_obj(struct object *list, size_t len) {
	return list_new(list, len, len, NULL);
}

struct object new_list_obj_data(struct object *list, size_t len, size_t cap) {
	return list_new(list, len, cap, NULL);
}

struct object new_list_slice(struct object *list, size_t len, struct gc_header *owner) {
	return list_new(list, len, len, owner);
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
