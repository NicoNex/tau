#include <string.h>
#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>
#include "object.h"

// Taken from: https://github.com/haipome/fnv/blob/master/fnv.c#L368
inline uint64_t fnv64a(char *s) {
	uint64_t hash = 0xcbf29ce484222325ULL;

	while (*s) {
		hash ^= (uint64_t)*s++;
	#if defined(NO_FNV_GCC_OPTIMIZATION)
		hash *= FNV_64_PRIME;
	#else
		hash += (hash << 1) + (hash << 4) + (hash << 5) +
				(hash << 7) + (hash << 8) + (hash << 40);
	#endif
	}
	return hash;
}

static inline uint64_t dtoi(double d) {
	uint64_t i;
	memcpy(&i, &d, sizeof(double));
	return i;
}

struct key_hash hash(struct object o) {
	switch (o.type) {
	case obj_integer:
	case obj_boolean:
		return (struct key_hash) {
			.type = o.type,
			.val = o.data.i
		};
	case obj_error:
	case obj_string:
		return (struct key_hash) {
			.type = o.type,
			.val = fnv64a(o.data.str->str)
		};
	case obj_float:
		return (struct key_hash) {
			.type = o.type,
			.val = dtoi(o.data.f)
		};
	default:
		return (struct key_hash) {0};
	}
}

static inline struct map_pair _map_get(struct map_node * restrict n, struct key_hash k) {
	if (n == NULL) {
		return (struct map_pair) {
			.key = null_obj,
			.val = null_obj
		};
	}

	int cmp = memcmp(&k, &n->key, sizeof(struct key_hash));
	if (cmp == 0) {
		return n->val;
	} else if (cmp < 0) {
		return _map_get(n->l, k);
	} else {
		return _map_get(n->r, k);
	}
}

static void mark_map_children(struct map_node * restrict n) {
	if (n != NULL) {
		mark_obj(n->val.key);
		mark_obj(n->val.val);
		mark_map_children(n->l);
		mark_map_children(n->r);
	}
}

static inline void _map_set(struct map_node **n, struct key_hash k, struct map_pair v) {
	if (*n == NULL) {
		struct map_node *tmp = malloc(sizeof(struct map_node));
		tmp->key = k;
		tmp->val = v;
		tmp->l = NULL;
		tmp->r = NULL;
		*n = tmp;
		return;
	}

	int cmp = memcmp(&k, &(*n)->key, sizeof(struct key_hash));
	if (cmp == 0) {
		(*n)->key = k;
		(*n)->val = v;
	} else if (cmp < 0) {
		_map_set(&(*n)->l, k, v);
	} else {
		_map_set(&(*n)->r, k, v);
	}
}

static inline void map_set_node(struct map_node **root, struct map_node **cur, struct map_node *n) {
	if (*cur == NULL) {
		*cur = n;
		return;
	}

	int cmp = memcmp(&(*cur)->key, &n->key, sizeof(struct key_hash));
	if (cmp == 0) {
		struct map_node *l = (*cur)->l;
		struct map_node *r = (*cur)->r;

		free(*cur);
		*cur = n;
		if (l != NULL) map_set_node(root, root, l);
		if (r != NULL) map_set_node(root, root, r);
	} else if (cmp < 0) {
		map_set_node(root, &(*cur)->l, n);
	} else {
		map_set_node(root, &(*cur)->r, n);
	}
}

static inline void _map_delete(struct map_node **root, struct map_node **n, struct key_hash k) {
	if (*n != NULL) {
		struct map_node *node = *n;
		int cmp = memcmp(&k, &node->key, sizeof(struct key_hash));

		if (cmp == 0) {
			*n = NULL;
			if (node->l) map_set_node(root, root, node->l);
			if (node->r) map_set_node(root, root, node->r);
			free(node);
		} else if (cmp < 0) {
			_map_delete(root, &(*n)->l, k);
		} else {
			_map_delete(root, &(*n)->r, k);
		}
	}
}

static inline void _map_dispose(struct map_node * restrict n) {
	if (n != NULL) {
		if (n->l != NULL) _map_dispose(n->l);
		if (n->r != NULL) _map_dispose(n->r);
		free(n);
	}
}

static inline void _map_keys(struct map_node * restrict n, struct list *list) {
	if (n != NULL) {
		list->list[list->len++] = n->val.key;
		_map_keys(n->l, list);
		_map_keys(n->r, list);
	}
}

struct object map_keys(struct object map) {
	struct object list = make_list(map.data.map->len);

	_map_keys(map.data.map->root, list.data.list);
	return list;
}

struct map_pair map_get(struct object map, struct object o) {
	return _map_get(map.data.map->root, hash(o));
}

struct map_pair map_set(struct object map, struct object k, struct object v) {
	struct map_pair p = (struct map_pair) {.key = k, .val = v};

	_map_set(&map.data.map->root, hash(k), p);
	map.data.map->len++;
	return p;
}

void map_delete(struct object map, struct object key) {
	_map_delete(&map.data.map->root, &map.data.map->root, hash(key));
	map.data.map->len--;
}

void dispose_map_obj(struct object map) {
	_map_dispose(map.data.map->root);
	free(map.data.map);
}

// TODO: actually return map content as string.
char *map_str(struct object map) {
	char *str = malloc(sizeof(char) * 64);
	str[63] = '\0';
	sprintf(str, "map[%p]", map.data.map->root);

	return str;
}

struct object new_map() {
	return (struct object) {
		.data.map = calloc(1, sizeof(struct map_node)),
		.type = obj_map,
		.gcdata = new_gcdata()
	};
}

void mark_map_obj(struct object m) {
	m.gcdata->marked = 1;
	mark_map_children(m.data.map->root);
}
