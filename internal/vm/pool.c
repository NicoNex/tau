#include <stdlib.h>
#include <string.h>
#include <math.h>
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
		p->cap = p->cap > 0 ? pow(2, ceil(log2(idx + 1))) : 1;
		p->list = realloc(p->list, p->cap * sizeof(struct object));
	}
	p->list[idx] = o;
	if (idx >= p->len) p->len = idx + 1;
}

// Appends a copy of len objects to the end of the pool. The array comes from
// the compiler, which is written in Go, and a Go allocation lives exactly as
// long as Go can see a reference to it: a pointer parked inside a malloc'd C
// struct is not one, so what the VM runs on has to be a copy of its own.
inline void pool_extend(struct pool *p, struct object *list, size_t len) {
	if (len == 0) return;

	if (p->len + len > p->cap) {
		p->cap = p->len + len;
		p->list = realloc(p->list, p->cap * sizeof(struct object));
	}
	memcpy(p->list + p->len, list, len * sizeof(struct object));
	p->len += len;
}

inline void pool_dispose(struct pool *p) {
	free(p->list);
	free(p);
}

struct modtab *new_modtab(void) {
	struct modtab *m = malloc(sizeof(struct modtab));

	m->list = NULL;
	m->len = 0;
	m->cap = 0;
	return m;
}

// Writes the module found under path into out and returns whether there was
// one. A miss leaves out alone.
int modtab_get(struct modtab *m, const char *path, struct object *out) {
	for (size_t i = 0; i < m->len; i++) {
		if (strcmp(m->list[i].path, path) == 0) {
			*out = m->list[i].mod;
			return 1;
		}
	}
	return 0;
}

// Takes a copy of path: it comes from Go, and what a C struct keeps has to be
// memory this side owns.
void modtab_put(struct modtab *m, const char *path, struct object mod) {
	if (m->len == m->cap) {
		m->cap = m->cap > 0 ? m->cap * 2 : 8;
		m->list = realloc(m->list, m->cap * sizeof(struct module));
	}
	m->list[m->len].path = strdup(path);
	m->list[m->len].mod = mod;
	m->len++;
}

void modtab_dispose(struct modtab *m) {
	for (size_t i = 0; i < m->len; i++) {
		free(m->list[i].path);
	}
	free(m->list);
	free(m);
}
