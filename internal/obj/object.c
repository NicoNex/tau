#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <ctype.h>
#include "object.h"

static inline struct object _object_get(struct object_node * restrict n, uint64_t key) {
	if (n == NULL) {
		return null_obj;
	}

	if (key == n->key) {
		return n->val;
	} else if (key < n->key) {
		return _object_get(n->l, key);
	} else {
		return _object_get(n->r, key);
	}
}

static void mark_object_children(struct object_node * restrict n) {
	if (n != NULL) {
		mark_obj(n->val);
		mark_object_children(n->l);
		mark_object_children(n->r);
	}
}

struct object object_to_module(struct object o);

static void _object_to_module(struct object mod, struct object_node * restrict n) {
	if (n != NULL) {
		if (isupper(*n->name)) {
			if (n->val.type == obj_object) {
				object_set(mod, n->name, object_to_module(n->val));
			} else {
				object_set(mod, n->name, n->val);
			}
		}
		_object_to_module(mod, n->l);
		_object_to_module(mod, n->r);
	}
}

static inline void _object_set(struct object_node **n, uint64_t key, char *name, struct object val) {
	if (*n == NULL) {
		*n = malloc(sizeof(struct object_node));
		(*n)->name = strdup(name);
		(*n)->key = key;
		(*n)->val = val;
		(*n)->l = NULL;
		(*n)->r = NULL;
		return;
	}

	uint64_t cur = (*n)->key;
	if (key == cur) {
		(*n)->val = val;
	} else if (key < cur) {
		_object_set(&(*n)->l, key, name, val);
	} else {
		_object_set(&(*n)->r, key, name, val);
	}
}

static inline void _object_dispose(struct object_node * restrict n) {
	if (n != NULL) {
		if (n->l != NULL) _object_dispose(n->l);
		if (n->r != NULL) _object_dispose(n->r);
		free(n->name);
		free(n);
	}
}

struct object object_get(struct object obj, char *name) {
	return _object_get(*obj.data.obj, fnv64a(name));
}

struct object object_set(struct object obj, char *name, struct object val) {
	_object_set(obj.data.obj, fnv64a(name), name, val);
	return val;
}

void dispose_object_obj(struct object obj) {
	_object_dispose(*obj.data.obj);
	free(obj.data.obj);
}

// Appends "name: value" for every field, in the order of the tree.
static void _object_obj_str(struct object_node * restrict n, char **buf, size_t *len, size_t *cap) {
	if (n == NULL) return;

	_object_obj_str(n->l, buf, len, cap);

	char *v = object_str(n->val);
	size_t need = strlen(n->name) + strlen(v) + 4;

	if (*len + need >= *cap) {
		*cap = (*len + need) * 2;
		*buf = realloc(*buf, *cap);
	}
	if (*len > 1) {
		memcpy(*buf + *len, ", ", 2);
		*len += 2;
	}
	*len += sprintf(*buf + *len, "%s: %s", n->name, v);
	free(v);

	_object_obj_str(n->r, buf, len, cap);
}

char *object_obj_str(struct object obj) {
	size_t cap = 64;
	size_t len = 1;
	char *str = malloc(cap);

	str[0] = '{';
	_object_obj_str(*obj.data.obj, &str, &len, &cap);

	if (len + 2 >= cap) str = realloc(str, len + 2);
	str[len++] = '}';
	str[len] = '\0';

	return str;
}

struct object new_object() {
	return (struct object) {
		.data.obj = calloc(1, sizeof(struct object_node *)),
		.type = obj_object,
		.gc = gc_header_alloc(),
	};
}

struct object object_to_module(struct object o) {
	struct object mod = new_object();

	_object_to_module(mod, *o.data.obj);
	return mod;
}

static void _object_keys(struct object_node * restrict n, struct list *list) {
	if (n != NULL) {
		char *name = strdup(n->name);
		list->list[list->len++] = new_string_obj(name, strlen(name));
		_object_keys(n->l, list);
		_object_keys(n->r, list);
	}
}

static size_t _object_len(struct object_node * restrict n) {
	return n == NULL ? 0 : 1 + _object_len(n->l) + _object_len(n->r);
}

struct object object_keys(struct object o) {
	struct object list = make_list(_object_len(*o.data.obj));

	_object_keys(*o.data.obj, list.data.list);
	return list;
}

void mark_object_obj(struct object o) {
	o.gc->mark |= GC_MARK;
	mark_object_children(*o.data.obj);
}
