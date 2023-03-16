#pragma once

#include <stdint.h>
#include "obj.h"

#define STACK_SIZE 2048
#define GLOBAL_SIZE 65536
#define MAX_FRAMES 1024

struct bytecode {
	uint8_t *insts;
	struct object *consts;
	uint32_t len;
	uint32_t nconsts;
	uint32_t bklen;
	struct bookmark *bookmarks;
	uint32_t ndefs;
};

struct frame {
	struct object cl;
	uint8_t *ip;
	uint8_t *start;
	uint32_t base_ptr;
};

struct state {
	struct object *consts;
	struct object globals[GLOBAL_SIZE];
	uint32_t nconsts;
	uint32_t ndefs;
};

struct vm {
	struct object stack[STACK_SIZE];
	struct frame frames[MAX_FRAMES];
	struct state state;
	uint32_t sp;
	uint32_t frame_idx;
	char *file;
};

struct state new_state();
struct vm *new_vm(char *file, struct bytecode bytecode);
struct vm *new_vm_with_state(char *file, struct bytecode bc, struct state state);
int vm_run(struct vm * restrict vm);
void vm_errorf(struct vm * restrict vm, const char *fmt, ...);
void go_vm_errorf(struct vm * restrict vm, const char *fmt);
struct object vm_last_popped_stack_elem(struct vm * restrict vm);
void vm_dispose(struct vm *vm);
