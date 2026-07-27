#pragma once

#include <stdint.h>
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

// One segment of the heap, owned by the thread that allocated into it.
struct heap {
	struct gc_header *root;
	size_t len;
	int64_t treshold;
	struct heap *next;
};

struct pool {
	struct object *list;
	size_t cap;
	size_t len;
};

struct state {
	struct pool *globals;
	struct pool consts;
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
void pool_dispose(struct pool *p);

// VM object.
struct state new_state();
struct vm *new_vm(char *file, struct bytecode bytecode);
struct vm *new_vm_with_state(char *file, struct bytecode bc, struct state state);
int vm_run(struct vm * restrict vm);
void vm_errorf(struct vm * restrict vm, const char *fmt, ...);
void go_vm_errorf(struct vm * restrict vm, const char *fmt);
struct object vm_last_popped_stack_elem(struct vm * restrict vm);
void vm_dispose(struct vm *vm);
void state_dispose(struct state s);
void set_exit();

// Garbage collector.
// The heap is global and shared by every VM (the main one and the tau routines).
// Collection is stop-the-world: the collector waits for all the other VMs to
// reach a safepoint before marking and sweeping.
extern volatile int gc_wanted;

void gc_init(void);
void gc_register(struct vm *vm);   // Makes the VM a root, initially parked.
void gc_unregister(struct vm *vm);
void *gc_activate(struct vm *vm);  // Called by the thread that runs the VM, returns the previous one.
void gc_restore(void *prev);       // Restores the VM active before gc_activate.
void *gc_add_roots(struct object *objs, size_t len);
void gc_remove_roots(void *handle);
void gc_park(void);                // Before blocking (pipes, native calls, IO).
void gc_unpark(void);              // After blocking.
void gc_safepoint(void);           // Cheap check, parks only if a GC is pending.
void gc_flush_headers(void);       // Frees the object headers this thread kept for reuse.
void heap_add(struct object obj);
void heap_dispose(void);
void gc(void);
