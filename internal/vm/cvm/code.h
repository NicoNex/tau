#ifndef CODE_H_
#define CODE_H_

#include <stdint.h>
#include <stddef.h>
#include <stdarg.h>

#define NUM_OPCODES 46

enum opcode {
	op_constant,
	op_true,
	op_false,
	op_null,
	op_list,
	op_map,
	op_closure,
	op_current_closure,

	op_add,
	op_sub,
	op_mul,
	op_div,
	op_mod,

	op_bw_and,
	op_bw_or,
	op_bw_xor,
	op_bw_not,
	op_bw_lshift,
	op_bw_rshift,

	op_and,
	op_or,
	op_equal,
	op_not_equal,
	op_greater_than,
	op_greater_than_equal,

	op_minus,
	op_bang,
	op_index,

	op_call,
	op_concurrent_call,
	op_return,
	op_return_value,

	op_jump,
	op_jump_not_truthy,

	op_dot,
	op_define,
	op_get_global,
	op_set_global,
	op_get_local,
	op_set_local,
	op_get_builtin,
	op_get_free,
	op_load_module,
	op_interpolate,

	op_pop,
	op_halt
};

struct definition {
	char *name;
	int *opwidths;
	int noperands;
};

extern struct definition definitions[NUM_OPCODES];

int lookup_def(enum opcode op, struct definition *def);
size_t make_bcode(uint8_t **code, size_t code_len, enum opcode op, ...);
size_t vmake_bcode(uint8_t **code, size_t code_len, enum opcode op, va_list operands);
int read_operands(struct definition def, uint8_t *ins, int **operands);
char *opcode_str(enum opcode op);

uint8_t read_uint8(uint8_t *ins);
uint16_t read_uint16(uint8_t *ins);
uint32_t read_uint32(uint8_t *ins);

#endif
