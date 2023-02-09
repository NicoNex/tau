#ifndef VM_H_
#define VM_H_

#include <stdint.h>
#include "obj.h"

#define STACK_SIZE 2048
#define GLOBAL_SIZE 65536
#define MAX_FRAMES 1024

struct bytecode {
	uint8_t *insts;
	struct object **consts;
	size_t len;
	size_t nconsts;
};

struct frame {
	struct object *cl;
	uint8_t *ip;
	uint8_t *start;
	uint32_t base_ptr;
};

struct state {
	struct symbol_table *st;
	struct object **consts;
	size_t nconsts;
	struct object *globals[GLOBAL_SIZE];
};

struct vm {
	struct object *stack[STACK_SIZE];
	struct frame frames[MAX_FRAMES];
	struct state state;
	uint32_t sp;
	uint32_t frame_idx;
};

struct state new_state();
struct vm *new_vm(struct bytecode bytecode);
struct vm *new_vm_with_state(struct bytecode bytecode, struct state state);
int vm_run(struct vm * restrict vm);
struct object *vm_last_popped_stack_elem(struct vm * restrict vm);
void vm_dispose(struct vm *vm);

#endif
