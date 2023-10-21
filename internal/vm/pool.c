#include <stdlib.h>
#include <string.h>
#include "vm.h"

inline struct pool *new_pool(size_t cap) {
	struct pool *p = malloc(sizeof(struct pool));
	p->list = calloc(cap, sizeof(struct object));
	p->cap = cap;
	p->len = 0;

	return p;
}

inline struct pool *poolcpy(struct pool *p) {
	struct pool *ret = malloc(sizeof(struct pool));
	ret->list = malloc(sizeof(struct object) * p->cap);
	ret->cap = p->cap;
	ret->len = p->len;
	memcpy(ret->list, p->list, sizeof(struct object) * p->cap);

	return ret;
}

inline void pool_append(struct pool *p, struct object o) {
	if (p->len == p->cap) {
		p->cap = p->cap > 0 ? p->cap * 2 : 1;
		p->list = realloc(p->list, p->cap * sizeof(struct object));
	}
	p->list[p->len++] = o;
}

inline void pool_insert(struct pool *p, size_t idx, struct object o) {
	if (idx >= p->cap) {
		p->cap = idx * 2;
		p->list = realloc(p->list, p->cap * sizeof(struct object));
	}
	p->list[idx] = o;
	p->len++;
}

inline void pool_dispose(struct pool *p) {
	free(p->list);
	free(p);
}
