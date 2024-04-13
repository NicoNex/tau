#pragma once

#include <stdint.h>
#include <setjmp.h>
#include "../obj/object.h"
#include "../compiler/bytecode.h"

#define STACK_SIZE  2048
#define GLOBAL_SIZE 65536
#define MAX_FRAMES  16384

struct frame {
	struct object cl;
	uint8_t *ip;
	uint8_t *start;
	uint32_t base_ptr;
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
	jmp_buf env;
};

// Pool object.
struct pool *new_pool(size_t cap);
struct pool *poolcpy(struct pool *p);
void pool_append(struct pool *p, struct object o);
void pool_insert(struct pool *p, size_t idx, struct object o);
void pool_dispose(struct pool *p);

// State object.
struct state new_state();
void state_free(struct state s);
void state_dispose(struct state s);

// VM object.
struct vm *new_vm(char *file, struct bytecode bytecode);
struct vm *new_vm_with_state(char *file, struct bytecode bc, struct state state);
int vm_run(struct vm * restrict vm);
void vm_free(struct vm * restrict vm);
void vm_errorf(struct vm * restrict vm, const char *fmt, ...);
void go_vm_errorf(struct vm * restrict vm, const char *fmt);
struct object vm_last_popped_stack_elem(struct vm * restrict vm);

void gc_init(void);
void set_exit();
