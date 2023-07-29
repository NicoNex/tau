#include <string.h>
#include <stdint.h>
#include <stdio.h>
#include "object.h"
#include "../vm/gc.h"

// Taken from: https://github.com/haipome/fnv/blob/master/fnv.c#L368
inline uint64_t fnv64a(char *s) {
	uint64_t hash = 0xcbf29ce484222325ULL;

	while (*s) {
		hash ^= (uint64_t)*s++;
		hash += (hash << 1) + (hash << 4) + (hash << 5) +
				(hash << 7) + (hash << 8) + (hash << 40);
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

	int32_t cmp = memcmp(&k, &n->key, sizeof(struct key_hash));
	if (cmp == 0) {
		return n->val;
	} else if (cmp < 0) {
		return _map_get(n->l, k);
	} else {
		return _map_get(n->r, k);
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

	int32_t cmp = memcmp(&k, &(*n)->key, sizeof(struct key_hash));
	if (cmp == 0) {
		(*n)->key = k;
		(*n)->val = v;
	} else if (cmp < 0) {
		_map_set(&(*n)->l, k, v);
	} else {
		_map_set(&(*n)->r, k, v);
	}
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
	};
}

struct object map_getsetter_get(struct getsetter *gs) {
	struct map_pair mp = map_get(gs->l, gs->r);
	return mp.val;
}

struct object map_getsetter_set(struct getsetter *gs, struct object val) {
	struct map_pair mp = map_set(gs->l, gs->r, val);
	return mp.val;
}
