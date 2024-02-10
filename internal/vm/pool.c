#include <stdlib.h>
#include <string.h>
#include <math.h>
#include "vm.h"
#include "gc.h"

inline struct pool *new_pool(size_t cap) {
	struct pool *p = GC_MALLOC(sizeof(struct pool));
	p->list = GC_CALLOC(cap, sizeof(struct object));
	p->cap = cap;
	p->len = 0;

	return p;
}

inline struct pool *poolcpy(struct pool *p) {
	struct pool *ret = GC_MALLOC(sizeof(struct pool));
	ret->list = GC_MALLOC(sizeof(struct object) * p->cap);
	ret->cap = p->cap;
	ret->len = p->len;
	memcpy(ret->list, p->list, sizeof(struct object) * p->cap);

	return ret;
}

inline void pool_append(struct pool *p, struct object o) {
	if (p->len == p->cap) {
		p->cap = p->cap > 0 ? p->cap * 2 : 1;
		p->list = GC_REALLOC(p->list, p->cap * sizeof(struct object));
	}
	p->list[p->len++] = o;
}

inline void pool_insert(struct pool *p, size_t idx, struct object o) {
	if (idx >= p->cap) {
		p->cap = p->cap > 0 ? pow(2, ceil(log2(idx + 1))) : 1;
		p->list = GC_REALLOC(p->list, p->cap * sizeof(struct object));
	}
	p->list[idx] = o;
	if (idx >= p->len) p->len = idx + 1;
}
