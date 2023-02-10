#pragma once

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

char *opcode_str(enum opcode op) {
	static char *strings[] = {
		"op_constant",
		"op_true",
		"op_false",
		"op_null",
		"op_list",
		"op_map",
		"op_closure",
		"op_current_closure",

		"op_add",
		"op_sub",
		"op_mul",
		"op_div",
		"op_mod",

		"op_bw_and",
		"op_bw_or",
		"op_bw_xor",
		"op_bw_not",
		"op_bw_lshift",
		"op_bw_rshift",

		"op_and",
		"op_or",
		"op_equal",
		"op_not_equal",
		"op_greater_than",
		"op_greater_than_equal",

		"op_minus",
		"op_bang",
		"op_index",

		"op_call",
		"op_concurrent_call",
		"op_return",
		"op_return_value",

		"op_jump",
		"op_jump_not_truthy",

		"op_dot",
		"op_define",
		"op_get_global",
		"op_set_global",
		"op_get_local",
		"op_set_local",
		"op_get_builtin",
		"op_get_free",
		"op_load_module",
		"op_interpolate",

		"op_pop",
		"op_halt"
	};

	return strings[op];
}
