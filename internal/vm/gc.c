#include <stdio.h>
#include <stdlib.h>
#include "vm.h"
#include "thrd.h"

/*
 * Global stop-the-world mark & sweep collector.
 *
 * There is a single heap shared by every VM because objects travel across
 * tau routines: the stack of a tau routine is copied from the caller, pipes
 * move objects between threads and the globals pool is shared. A per-VM heap
 * cannot know when the other VMs are done with an object.
 *
 * Every VM registers itself as a set of roots. When a VM decides to collect
 * it raises gc_wanted and waits until all the other VMs are parked, that is,
 * either sitting on a safepoint or blocked on something that doesn't touch
 * objects (pipes, native calls, terminal input).
 */

struct vm_node {
	struct vm *vm;
	int parked;
	struct vm_node *next;
};

struct root_node {
	struct object *objs;
	size_t len;
	struct root_node *next;
};

// The heap is cut into one segment per thread. A thread only ever pushes
// onto its own, so allocating takes no lock at all; the collector, which
// runs with the world stopped, is the only one that walks them all.
static struct heap *segments = NULL;
static __thread struct heap *segment = NULL;

// How many objects live in the heap, and how many may be allocated before
// the next collection. Both are only written with the world stopped.
static size_t heap_len = 0;
static int64_t heap_treshold = HEAP_TRESHOLD;

// How many objects this thread may still allocate before it is worth asking
// for a collection. Counting down a thread local number is all the fast path
// of an allocation can afford: everything exact needs the lock, and the lock
// is what a collection is for.
static __thread int64_t budget = HEAP_TRESHOLD;
static mtx_t mu;
static cnd_t cnd;
static struct vm_node *vms = NULL;
static struct root_node *roots = NULL;
static volatile int nvms = 0;
static int nparked = 0;
static int initialised = 0;

volatile int gc_wanted = 0;
// Bumped at every collection, tells apart the objects visited by the current
// mark phase from the ones visited by the previous ones.
uint32_t gc_epoch = 1;

// The node of the VM the current thread is running, NULL for threads that
// don't run a VM (e.g. `tau somebuiltin()`).
static __thread struct vm_node *self = NULL;

// Recycled headers, kept per thread so that allocating an object needs no
// lock. A thread that ends gives back what it holds in gc_flush_headers.
static __thread struct gc_header *free_headers = NULL;
static __thread int nfree_headers = 0;
// Build with -DMAX_FREE_HEADERS=0 to always go through malloc, so that a
// sanitizer can see the use-after-frees the reuse would hide.
#ifndef MAX_FREE_HEADERS
	#define MAX_FREE_HEADERS 1024
#endif

// Header of a new object: the state of the collector and the node of the
// heap in a single allocation.
struct gc_header *gc_header_alloc(void) {
	struct gc_header *h = free_headers;

	if (h != NULL) {
		free_headers = h->next;
		nfree_headers--;
	} else {
		h = malloc(sizeof(struct gc_header));
	}
	h->mark = 0;

	return h;
}

static void gc_mark_recycle(struct gc_header *h) {
	if (nfree_headers >= MAX_FREE_HEADERS) {
		free(h);
		return;
	}
	h->next = free_headers;
	free_headers = h;
	nfree_headers++;
}

// Called by a thread that is about to end, its headers are of no use to it.
void gc_flush_headers(void) {
	for (struct gc_header *h = free_headers; h != NULL;) {
		struct gc_header *next = h->next;
		free(h);
		h = next;
	}
	free_headers = NULL;
	nfree_headers = 0;
}

// ponytail: no lock here, gc_init runs from new_vm before any tau routine exists.
void gc_init(void) {
	if (initialised) return;
	initialised = 1;
	mtx_init(&mu, mtx_plain);
	cnd_init(&cnd);
}

// Locks the heap mutex joining any collection that is pending or in progress,
// so that a thread waiting for the mutex can never stall the collector.
static void gc_lock(void) {
	mtx_lock(&mu);
	while (gc_wanted) {
		if (self != NULL && !self->parked) {
			self->parked = 1;
			nparked++;
			cnd_broadcast(&cnd);
		}
		cnd_wait(&cnd, &mu);
	}
	if (self != NULL && self->parked) {
		self->parked = 0;
		nparked--;
	}
}

void gc_register(struct vm *vm) {
	struct vm_node *n = malloc(sizeof(struct vm_node));
	n->vm = vm;
	// A registered VM starts parked: its roots are already reachable but no
	// thread is mutating them yet.
	n->parked = 1;

	mtx_lock(&mu);
	n->next = vms;
	vms = n;
	nvms++;
	nparked++;
	cnd_broadcast(&cnd);
	mtx_unlock(&mu);

	vm->gc_node = n;
}

void gc_unregister(struct vm *vm) {
	struct vm_node *n = vm->gc_node;
	if (n == NULL) return;

	mtx_lock(&mu);
	for (struct vm_node **p = &vms; *p != NULL; p = &(*p)->next) {
		if (*p == n) {
			*p = n->next;
			break;
		}
	}
	nvms--;
	if (n->parked) nparked--;
	cnd_broadcast(&cnd);
	mtx_unlock(&mu);

	if (self == n) self = NULL;
	vm->gc_node = NULL;
	free(n);
}

void *gc_activate(struct vm *vm) {
	struct vm_node *prev = self;
	self = vm->gc_node;
	gc_unpark();
	return prev;
}

// Restores the VM that was running before a nested vm_run (module import).
void gc_restore(void *prev) {
	self = prev;
}

// Roots that don't belong to any VM, e.g. the arguments of a builtin called
// with `tau`.
void *gc_add_roots(struct object *objs, size_t len) {
	struct root_node *n = malloc(sizeof(struct root_node));
	n->objs = objs;
	n->len = len;

	gc_lock();
	n->next = roots;
	roots = n;
	mtx_unlock(&mu);

	return n;
}

void gc_remove_roots(void *handle) {
	if (handle == NULL) return;

	gc_lock();
	for (struct root_node **p = &roots; *p != NULL; p = &(*p)->next) {
		if (*p == handle) {
			*p = ((struct root_node *) handle)->next;
			break;
		}
	}
	mtx_unlock(&mu);
	free(handle);
}

void gc_park(void) {
	if (self == NULL) return;

	mtx_lock(&mu);
	if (!self->parked) {
		self->parked = 1;
		nparked++;
		cnd_broadcast(&cnd);
	}
	mtx_unlock(&mu);
}

void gc_unpark(void) {
	if (self == NULL) return;
	gc_lock();
	mtx_unlock(&mu);
}

void gc_safepoint(void) {
	if (!gc_wanted || self == NULL) return;
	gc_park();
	gc_unpark();
}

// The segment of this thread, made on first use. Only the list of segments
// is shared, and only while a segment is being added to it.
static struct heap *heap_segment(void) {
	if (segment != NULL) return segment;

	struct heap *s = malloc(sizeof(struct heap));
	s->root = NULL;
	s->len = 0;
	s->treshold = 0;

	gc_lock();
	s->next = segments;
	segments = s;
	mtx_unlock(&mu);

	segment = s;
	return s;
}

// Takes ownership of the object. Objects that are already tracked (e.g. one
// received from a pipe and returned by a builtin) are ignored.
void heap_add(struct object obj) {
	if (obj.gc == NULL || (obj.gc->mark & GC_TRACKED)) return;
	obj.gc->mark |= GC_TRACKED;

	// The header is the node, it was allocated with the object.
	struct gc_header *node = obj.gc;
	node->obj = obj;

	struct heap *s = heap_segment();
	node->next = s->root;
	s->root = node;
	s->len++;
	budget--;
}

static void mark_vm(struct vm *vm) {
	// Only up to sp: the slots above it hold stale copies of objects that may
	// have been collected already.
	for (uint32_t i = 0; i < vm->sp && i < STACK_SIZE; i++) {
		mark_obj(vm->stack[i]);
	}
	for (uint32_t i = 0; i <= vm->frame_idx; i++) {
		mark_obj(vm->frames[i].cl);
	}

	struct pool *globals = vm->state.globals;
	if (globals != NULL) {
		for (size_t i = 0; i < globals->len; i++) {
			mark_obj(globals->list[i]);
		}
	}
	for (size_t i = 0; i < vm->state.consts.len; i++) {
		mark_obj(vm->state.consts.list[i]);
	}
}

// Frees the unmarked objects of every segment and clears the mark of the
// surviving ones. Must be called with the world stopped, which is what makes
// it safe to walk segments other threads own.
static void sweep(void) {
	heap_len = 0;

	for (struct heap *s = segments; s != NULL; s = s->next) {
		struct gc_header **prev = &s->root;
		struct gc_header *n = s->root;

		while (n != NULL) {
			struct gc_header *next = n->next;

			if (n->mark & GC_MARK) {
				n->mark &= ~GC_MARK;
				prev = &n->next;
			} else {
				*prev = next;
				s->len--;
				free_obj(n->obj);
				gc_mark_recycle(n);
			}
			n = next;
		}

		heap_len += s->len;
	}
}

// The share of the room left that one thread may fill on its own before it
// asks whether a collection is due.
static int64_t gc_budget(void) {
	int nthreads = nvms > 0 ? nvms : 1;
	int64_t left = (heap_treshold - (int64_t) heap_len) / nthreads;

	return left > HEAP_TRESHOLD ? left : HEAP_TRESHOLD;
}

void gc(void) {
	if (gc_wanted) {
		gc_safepoint();
		return;
	}

	if (budget > 0) return;

	gc_lock();

	// The real size of the heap, garbage included: every segment knows how
	// much it holds, and there are as many segments as there are threads.
	size_t total = 0;
	for (struct heap *s = segments; s != NULL; s = s->next) {
		total += s->len;
	}

	if ((int64_t) total < heap_treshold) {
		budget = gc_budget();
		mtx_unlock(&mu);
		return;
	}

#ifdef GC_DEBUG
	printf("heap size before: %lu\n", heap_len);
#endif

	gc_wanted = 1;
	while (nparked < nvms - 1) {
		cnd_wait(&cnd, &mu);
	}

	// ponytail: the epoch wraps after 2^30 collections, an object visited
	// exactly that many collections ago would be skipped once.
	if (++gc_epoch > (UINT32_MAX >> GC_EPOCH_SHIFT)) gc_epoch = 1;

	for (struct vm_node *n = vms; n != NULL; n = n->next) {
		mark_vm(n->vm);
	}
	for (struct root_node *r = roots; r != NULL; r = r->next) {
		for (size_t i = 0; i < r->len; i++) {
			mark_obj(r->objs[i]);
		}
	}
	sweep();

	// Grows with the live set and always leaves room for HEAP_TRESHOLD new
	// objects, otherwise a program with few live objects and lots of garbage
	// would collect at every single allocation burst.
	// The room left for new objects grows with the number of threads: every
	// one of them allocates, and every collection stops all of them, so the
	// more there are the less often it is worth stopping them.
	heap_treshold = heap_len * 2 + HEAP_TRESHOLD * (nvms > 0 ? nvms : 1);
	budget = gc_budget();
	gc_wanted = 0;
	cnd_broadcast(&cnd);
	mtx_unlock(&mu);

#ifdef GC_DEBUG
	printf("heap size after: %lu\n", heap_len);
#endif
}

void heap_dispose(void) {
	gc_lock();
	for (struct heap *s = segments; s != NULL;) {
		struct heap *nexts = s->next;

		for (struct gc_header *n = s->root; n != NULL;) {
			struct gc_header *tmp = n->next;
			free_obj(n->obj);
			free(n);
			n = tmp;
		}
		free(s);
		s = nexts;
	}

	segments = NULL;
	segment = NULL;
	heap_len = 0;
	heap_treshold = HEAP_TRESHOLD;
	budget = HEAP_TRESHOLD;
	mtx_unlock(&mu);
}
