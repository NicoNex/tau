#pragma once

#include <stdint.h>
#include "obj.h"

#ifdef __GLIBC__
#include <bits/stdint-uintn.h>
#define uint32_t uint_fast32_t
#endif


#define STACK_SIZE 2048
#define GLOBAL_SIZE 65536
#define MAX_FRAMES 1024

struct bytecode {
	uint8_t *insts;
	struct object *consts;
	size_t len;
	size_t nconsts;
	size_t bklen;
	struct bookmark *bookmarks;
};

struct frame {
	struct object cl;
	uint8_t *ip;
	uint8_t *start;
	uint32_t base_ptr;
};

struct state {
	struct object *consts;
	uint32_t nconsts;
	struct object globals[GLOBAL_SIZE];
};

struct vm {
	struct object stack[STACK_SIZE];
	struct frame frames[MAX_FRAMES];
	struct state state;
	uint32_t sp;
	uint32_t frame_idx;
	char *file;
	uint32_t locals[GLOBAL_SIZE];
};

struct state new_state();
struct vm *new_vm(char *file, struct bytecode bytecode);
struct vm *new_vm_with_state(char *file, struct bytecode bc, struct state state);
int vm_run(struct vm * restrict vm);
void vm_errorf(struct vm * restrict vm, const char *fmt, ...);
void go_vm_errorf(struct vm * restrict vm, const char *fmt);
struct object vm_last_popped_stack_elem(struct vm * restrict vm);
void vm_dispose(struct vm *vm);
