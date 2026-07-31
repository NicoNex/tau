#pragma once

#include <stdint.h>
#include <stdlib.h>
#include <setjmp.h>
#include "../obj/object.h"
#include "../compiler/bytecode.h"

#define STACK_SIZE    2048
#define GLOBAL_SIZE   65536
#define MAX_FRAMES    16384
#define HEAP_TRESHOLD 1024

struct frame {
	struct object cl;
	uint8_t *ip;
	uint8_t *start;
	uint32_t base_ptr;
};

// One segment of the heap, owned by the thread that allocated into it. `next`
// links every segment there is, `free_next` only the ones no thread owns
// anymore and that the next thread to come may take over.
struct heap {
	struct gc_header *root;
	size_t len;
	int64_t treshold;
	struct heap *next;
	struct heap *free_next;
};

struct pool {
	struct object *list;
	size_t cap;
	size_t len;
};

// A module that has been imported, kept under the path it was found at so
// that importing it twice runs it once.
struct module {
	char *path;
	struct object mod;
};

// ponytail: a linear scan, a program imports tens of modules and not
// thousands. A map if that ever stops being true.
struct modtab {
	struct module *list;
	size_t len;
	size_t cap;
};

struct modtab *new_modtab(void);
int modtab_get(struct modtab *m, const char *path, struct object *out);
void modtab_put(struct modtab *m, const char *path, struct object mod);
void modtab_dispose(struct modtab *m);

struct state {
	struct pool *globals;
	// A pointer and not a value: every VM of a program shares this pool, and
	// an import appends to it. Two copies of the struct would mean one of them
	// keeps the address the other has just reallocated away.
	struct pool *consts;
	// The modules this program has already imported. Shared with the tau
	// routines like the rest of the state, and walked by the collector: a
	// module holds the objects a program reaches through it.
	struct modtab *mods;
	uint32_t ndefs;
};

struct vm {
	struct state state;
	struct object stack[STACK_SIZE];
	struct frame frames[MAX_FRAMES];
	uint32_t sp;
	uint32_t frame_idx;
	char *file;
	void *gc_node;
	jmp_buf env;
};

// Pool object.
struct pool *new_pool(size_t cap);
struct pool *poolcpy(struct pool *p);
void pool_append(struct pool *p, struct object o);
void pool_insert(struct pool *p, size_t idx, struct object o);
void pool_extend(struct pool *p, struct object *list, size_t len);
void pool_dispose(struct pool *p);

// VM object.
struct state new_state();
struct vm *new_vm(char *file, struct bytecode bytecode);
struct vm *new_vm_with_state(char *file, struct bytecode bc, struct state state);
int vm_run(struct vm * restrict vm);
void vm_errorf(struct vm * restrict vm, const char *fmt, ...);
void go_vm_errorf(struct vm * restrict vm, const char *fmt);
struct object vm_call_tau(struct vm * restrict vm, struct object cl, struct object *args, size_t nargs);
struct object vm_last_popped_stack_elem(struct vm * restrict vm);
void vm_dispose(struct vm *vm);
void state_dispose(struct state s);
void set_exit();

// Garbage collector.
// The heap is global and shared by every VM (the main one and the tau routines).
// Collection is stop-the-world: the collector waits for all the other VMs to
// reach a safepoint before marking and sweeping.
// Raised by the thread that is about to collect, read by every other one at
// its safepoints. Touched by more than one thread at a time, so it is read and
// written with the atomic builtins: `volatile` orders nothing and says nothing
// to the compiler about other threads. Relaxed is enough, what a reader needs
// to see is only whether it has to park, and the parking itself goes through
// the heap mutex, which is what orders everything else.
extern int gc_wanted;
#define gc_pending() __atomic_load_n(&gc_wanted, __ATOMIC_RELAXED)
#define gc_set_wanted(v) __atomic_store_n(&gc_wanted, (v), __ATOMIC_RELAXED)

// Small things the Go side reaches for. They live here and not in the preamble
// of one file because more than one file needs them, and a cgo preamble is
// only visible to the file it is written in.
static inline struct object get_global(struct pool *globals, size_t idx) {
	return globals->list[idx];
}

static inline void set_const(struct object *list, size_t idx, struct object o) {
	list[idx] = o;
}

extern char *tau_module_dir;
extern char *tau_exec_dir;

static inline void set_module_dir(char *dir) {
	free(tau_module_dir);
	tau_module_dir = dir;
}

// Where the interpreter itself is, so that a shared object shipped beside it
// is found from any working directory and under any prefix.
static inline void set_exec_dir(char *dir) {
	free(tau_exec_dir);
	tau_exec_dir = dir;
}

void gc_init(void);
void gc_register(struct vm *vm);   // Makes the VM a root, initially parked.
void gc_unregister(struct vm *vm);
void *gc_activate(struct vm *vm);  // Called by the thread that runs the VM, returns the previous one.
void gc_restore(void *prev);       // Restores the VM active before gc_activate.
struct vm *gc_current_vm(void);
void *gc_add_roots(struct object *objs, size_t len);
void gc_remove_roots(void *handle);
void gc_park(void);                // Before blocking (pipes, native calls, IO).
void gc_unpark(void);              // After blocking.
void gc_safepoint(void);           // Cheap check, parks only if a GC is pending.
void gc_flush_headers(void);       // Frees the object headers this thread kept for reuse.
void gc_release_segment(void);     // Gives this thread's heap segment to the next one that needs it.
void heap_add(struct object obj);
void gc(void);
