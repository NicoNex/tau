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

static struct heap heap = {.root = NULL, .len = 0, .treshold = HEAP_TRESHOLD};
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

// Takes ownership of the object. Objects that are already tracked (e.g. one
// received from a pipe and returned by a builtin) are ignored.
void heap_add(struct object obj) {
	if (obj.marked == NULL || (*obj.marked & GC_TRACKED)) return;
	*obj.marked |= GC_TRACKED;

	struct heap_node *node = malloc(sizeof(struct heap_node));
	node->obj = obj;

	// Fast path: with a single registered VM only its own thread allocates and
	// only it can spawn a second one, so there is nobody to synchronise with.
	// Threads running a `tau builtin()` never get here, they don't own a heap.
	if (nvms <= 1) {
		node->next = heap.root;
		heap.root = node;
		heap.len++;
		return;
	}

	gc_lock();
	node->next = heap.root;
	heap.root = node;
	heap.len++;
	mtx_unlock(&mu);
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

// Frees the unmarked objects and clears the mark of the surviving ones.
// Must be called with the world stopped.
static void sweep(void) {
	struct heap_node **prev = &heap.root;
	struct heap_node *n = heap.root;

	while (n != NULL) {
		struct heap_node *next = n->next;

		if (*n->obj.marked & GC_MARK) {
			*n->obj.marked &= ~GC_MARK;
			prev = &n->next;
		} else {
			*prev = next;
			heap.len--;
			free_obj(n->obj);
			free(n);
		}
		n = next;
	}
}

void gc(void) {
	if (gc_wanted) {
		gc_safepoint();
		return;
	}
	// Racy read, worst case the collection happens one allocation later.
	if (heap.len < heap.treshold) return;

	gc_lock();
	if (heap.len < heap.treshold) {
		mtx_unlock(&mu);
		return;
	}

#ifdef GC_DEBUG
	printf("heap size before: %lu\n", heap.len);
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
	heap.treshold = heap.len * 2 + HEAP_TRESHOLD;
	gc_wanted = 0;
	cnd_broadcast(&cnd);
	mtx_unlock(&mu);

#ifdef GC_DEBUG
	printf("heap size after: %lu\n", heap.len);
#endif
}

void heap_dispose(void) {
	gc_lock();
	for (struct heap_node *n = heap.root; n != NULL;) {
		struct heap_node *tmp = n->next;
		free_obj(n->obj);
		free(n);
		n = tmp;
	}
	heap.root = NULL;
	heap.len = 0;
	heap.treshold = HEAP_TRESHOLD;
	mtx_unlock(&mu);
}
