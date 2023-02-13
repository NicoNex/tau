#include <stdint.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

#include "decoder.h"
#include "vm.h"
#include "obj.h"

static inline uint32_t read_uint32(uint8_t *ins) {
	return (ins[0] << 24) | (ins[1] << 16) | (ins[2] << 8) | ins[3];
}

static inline uint64_t read_uint64(uint8_t *ins) {
	return ((uint64_t) read_uint32(ins) << 32) | ((uint64_t) read_uint32(&ins[4]));
}

static inline __attribute__((always_inline))
char *string(size_t len) {
	char *s = malloc(sizeof(char) * (len+1));
	s[len] = '\0';

	return s;
}

static uint8_t *read_bookmarks(uint8_t *data, struct bookmark *bmarks, size_t len) {
	for (int i = 0; i < len; i++) {
		uint32_t offset = read_uint32(data); data += 4;
		uint32_t lineno = read_uint32(data); data += 4;
		uint32_t pos = read_uint32(data); data += 4;
		uint32_t len = read_uint32(data); data += 4;
		char *line = string(len);

		memcpy(line, data, len);
		data += len;
		bmarks[i] = (struct bookmark) {
			.offset = offset,
			.lineno = lineno,
			.pos = pos,
			.len = len,
			.line = line
		};
	}

	return data;
}

static uint8_t *decode_objects(uint8_t *data, struct object *objs, size_t n) {
	for (int i = 0; i < n; i++) {
		switch (*data++) {
		case obj_null:
			objs[i] = null_obj;
			break;

		case obj_error: {
			size_t len = read_uint32(data);
			data += 4;
			char *msg = string(len);
			memcpy(msg, data, len);
			data += len;
			objs[i] = new_error_obj(msg, len);
			break;
		}

		case obj_integer:
			objs[i] = new_integer_obj(read_uint64(data));
			data += 8;
			break;

		case obj_float: {
			struct object f = new_float_obj(0);
			f.data.i = read_uint64(data);
			objs[i] = f;
			data += 8;
			break;
		}

		case obj_boolean:
			objs[i] = parse_bool(*data++);
			break;

		case obj_string: {
			size_t len = read_uint32(data);
			data += 4;
			char *str = string(len);
			memcpy(str, data, len);
			data += len;
			objs[i] = new_string_obj(str, len);
			break;
		}

		case obj_function: {
			size_t num_params = read_uint32(data);
			data += 4;
			size_t num_locals = read_uint32(data);
			data += 4;
			size_t ilen = read_uint32(data);
			data += 4;

			uint8_t *inst = malloc(sizeof(uint8_t) * ilen);
			memcpy(inst, data, ilen);
			data += ilen;
			size_t blen = read_uint32(data);
			data += 4;
			struct bookmark *bookmarks = malloc(sizeof(struct bookmark) * blen);
			data = read_bookmarks(data, bookmarks, blen);
			objs[i] = new_function_obj(inst, ilen, num_locals, num_params, bookmarks, blen);
			break;
		}

		default:
			puts("decoder: unsupported type");
			return NULL;
		}
	}
	return data;
}

struct bytecode tau_decode(uint8_t *data, size_t len) {
	struct bytecode bc = {0};
	if (len == 0) {
		return bc;
	}

	// Decode instructions.
	bc.len = read_uint32(data); data += 4;
	bc.insts = malloc(sizeof(uint8_t) * bc.len);
	memcpy(bc.insts, data, bc.len);
	data += bc.len;

	// Decode constants.
	bc.nconsts = read_uint32(data); data += 4;
	bc.consts = malloc(sizeof(struct object) * bc.nconsts);
	if ((data = decode_objects(data, bc.consts, bc.nconsts)) == NULL) {
		return (struct bytecode) {0};
	}

	// Decode bookmarks.
	bc.bklen = read_uint32(data); data += 4;
	bc.bookmarks = malloc(sizeof(struct bookmark) * bc.bklen);
	data = read_bookmarks(data, bc.bookmarks, bc.bklen);

	return bc;
}
