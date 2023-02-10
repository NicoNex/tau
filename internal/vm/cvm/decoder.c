#include <stdint.h>
#include <string.h>
#include <stdlib.h>

#include "decoder.h"
#include "vm.h"
#include "obj.h"

static inline uint32_t read_uint32(uint8_t *ins) {
	return (ins[0] << 24) | (ins[1] << 16) | (ins[2] << 8) | ins[3];
}

static inline uint64_t read_uint64(uint8_t *ins) {
	return ((uint64_t) read_uint32(ins) << 32) | ((uint64_t) read_uint32(&ins[4]));
}

static inline double read_double(uint8_t *ins) {
	return *(double *) ins;
}

static int skip_bookmarks(uint8_t *data) {
	int len = read_uint32(data);
	int pos = 4;

	for (int i = 0; i < len; i++) {
		int slen = read_uint32(&data[pos+12]);
		pos += 16 + slen;
	}

	return pos;
}

#include <stdio.h>
static int decode_objects(struct object *objs, uint8_t *data, size_t n) {
	int pos = 0;

	for (int i = 0; i < n; i++) {
		switch (data[pos++]) {
		case obj_null:
			objs[i] = null_obj;
			break;

		case obj_error: {
			int len = read_uint32(&data[pos]);
			pos += 4;
			char *msg = calloc(len+1, sizeof(char));
			memcpy(msg, &data[pos], len);
			objs[i] = new_error_obj(msg, len);
			break;
		}

		case obj_integer:
			objs[i] = new_integer_obj(read_uint64(&data[pos]));
			pos += 8;
			break;

		case obj_float:
			objs[i] = new_float_obj(read_double(&data[pos]));
			pos += 8;
			break;

		case obj_boolean:
			objs[i] = parse_bool(data[pos++]);
			break;

		case obj_string: {
			int len = read_uint32(&data[pos]);
			pos += 4;
			char *str = calloc(len+1, sizeof(char));
			memcpy(str, &data[pos], len);
			objs[i] = new_string_obj(str, len);
			break;
		}

		case obj_function: {
			int num_params = read_uint32(&data[pos]);
			int num_locals = read_uint32(&data[pos+4]);
			int ilen = read_uint32(&data[pos+8]);
			pos += 12;

			uint8_t *inst = malloc(sizeof(uint8_t) * ilen);
			memcpy(inst, &data[pos], ilen);
			pos += ilen;
			pos += skip_bookmarks(&data[pos]);
			// TODO: also decode bookmarks.

			objs[i] = new_function_obj(inst, ilen, num_locals, num_params);
			break;
		}

		default:
			return -1;
		}
	}
	return pos;
}

#include <stdio.h>
struct bytecode tau_decode(uint8_t *data, size_t len) {
	struct bytecode bc = {0};
	if (len == 0) {
		return bc;
	}

	int pos = 0;
	bc.len = read_uint32(data);
	bc.insts = malloc(sizeof(uint8_t) * bc.len);
	memcpy(bc.insts, &data[4], bc.len);
	pos += 4 + bc.len;

	bc.nconsts = read_uint32(&data[pos]);
	bc.consts = malloc(sizeof(struct object) * bc.nconsts);
	pos += 4;
	int obj_end = decode_objects(bc.consts, &data[pos], bc.nconsts);
	if (obj_end == -1) {
		return (struct bytecode) {0};
	}
	pos += obj_end;

	// TODO: also decode bookmarks.

	return bc;
}
