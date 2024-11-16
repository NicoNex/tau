#include <stdlib.h>
#include <string.h>
#include "object.h"

void dispose_list_obj(struct object o) {
	// Free everything if it's not a slice (marked parent bit is set to NULL).
	if (o.data.list->parent == NULL) {
		free(o.data.list->list);
	}
	free(o.data.list);
	free(o.gcdata);
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
	l.gcdata->marked = 1;
	if (l.data.list->parent != NULL) {
		mark_list_obj(*l.data.list->parent);
		// We return early so that the marking of the elements happen only for
		// the parent object.
		return;
	}
	#pragma omp parallel for
	for (uint32_t i = 0; i < l.data.list->len; i++) {
		mark_obj(l.data.list->list[i]);
	}
}

struct object make_list(size_t cap) {
	struct list *l = malloc(sizeof(struct list));
	l->list = calloc(cap, sizeof(struct object));
	l->len = 0;
	l->cap = cap;
	l->parent = NULL;

	return (struct object) {
		.data.list = l,
		.type = obj_list,
		.gcdata = new_gcdata()
	};
}

struct object new_list_obj(struct object *list, size_t len) {
	struct list *l = malloc(sizeof(struct list));
	l->list = list;
	l->len = len;
	l->cap = len;
	l->parent = NULL;

	return (struct object) {
		.data.list = l,
		.type = obj_list,
		.gcdata = new_gcdata()
	};
}

struct object new_list_obj_data(struct object *list, size_t len, size_t cap) {
	struct list *l = malloc(sizeof(struct list));
	l->list = list;
	l->len = len;
	l->cap = cap;
	l->parent = NULL;

	return (struct object) {
		.data.list = l,
		.type = obj_list,
		.gcdata = new_gcdata()
	};
}

struct object new_list_slice(struct object *list, size_t len, struct object *parent) {
	struct list *l = malloc(sizeof(struct list));
	l->list = list;
	l->len = len;
	l->cap = len;
	l->parent = parent;

	return (struct object) {
		.data.list = l,
		.type = obj_list,
		.gcdata = new_gcdata(),
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
