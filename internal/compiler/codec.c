#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <stdarg.h>
#include <stdio.h>
#include "bytecode.h"
#include "../obj/object.h"
#include "../tauerr/bookmark.h"

__attribute__((noreturn))
static inline void fatalf(char * restrict fmt, ...) {
	va_list args;
	va_start(args, fmt);
	vprintf(fmt, args);
	va_end(args);
	exit(1);
}

static inline void write_byte(struct buffer *buf, uint8_t b) {
	if (buf->cap == 0) {
		buf->buf = malloc(sizeof(uint8_t) * ++buf->cap);
	} else if (buf->len == buf->cap) {
		buf->cap *= 2;
		buf->buf = realloc(buf->buf, buf->cap * sizeof(uint8_t));
	}
	buf->buf[buf->len++] = b;
}

static inline void write_bytes(struct buffer *buf, uint8_t *bytes, size_t len) {
	for (int i = 0; i < len; i++) {
		write_byte(buf, bytes[i]);
	}
}

static inline void write_string(struct buffer *buf, const char *str) {
	for (int i = 0; str[i] != '\0'; i++) {
		write_byte(buf, str[i]);
	}
}

static inline void write_uint32(struct buffer *buf, uint32_t n) {
	write_byte(buf, n >> 24);
	write_byte(buf, n >> 16);
	write_byte(buf, n >> 8);
	write_byte(buf, n);
}

static inline void write_uint64(struct buffer *buf, uint64_t n) {
	write_uint32(buf, n >> 32);
	write_uint32(buf, n);
}

static inline void encode_bookmarks(struct buffer *buf, struct bookmark *bookmarks, size_t len) {
	for (int i = 0; i < len; i++) {
		struct bookmark b = bookmarks[i];

		write_uint32(buf, b.offset);
		write_uint32(buf, b.lineno);
		write_uint32(buf, b.pos);
		write_uint32(buf, b.len);
		write_string(buf, b.line);
	}
}

static inline void encode_objects(struct buffer *buf, struct object *objs, size_t len) {
	for (int i = 0; i < len; i++) {
		struct object o = objs[i];

		write_byte(buf, o.type);
		switch (o.type) {
		case obj_null:
			break;
		case obj_boolean:
			write_byte(buf, o.data.i);
			break;
		case obj_float:
		case obj_integer:
			write_uint64(buf, o.data.i);
			break;
		case obj_string:
			write_uint32(buf, o.data.str->len);
			write_string(buf, o.data.str->str);
			break;
		case obj_function: {
			struct function *fn = o.data.fn;
			write_uint32(buf, fn->num_params);
			write_uint32(buf, fn->num_locals);
			write_uint32(buf, fn->len);
			write_bytes(buf, fn->instructions, fn->len);
			write_uint32(buf, fn->bklen);
			encode_bookmarks(buf, fn->bookmarks, fn->bklen);
			break;
		}
		default:
			fatalf("encoder: unsupported encoding for type %s\n", otype_str(o.type));
		}
	}
}

inline struct buffer tau_encode(struct bytecode bc) {
	struct buffer buf = (struct buffer) {0};

	write_uint32(&buf, bc.ndefs);
	write_uint32(&buf, bc.len);
	write_bytes(&buf, bc.insts, bc.len);
	write_uint32(&buf, bc.nconsts);
	encode_objects(&buf, bc.consts, bc.nconsts);
	write_uint32(&buf, bc.bklen);
	encode_bookmarks(&buf, bc.bookmarks, bc.bklen);
	return buf;
}

inline void free_buffer(struct buffer buf) {
	free(buf.buf);
}

struct reader {
	uint8_t *buf;
	uint32_t len;
	uint32_t pos;
};

static inline uint8_t read_byte(struct reader *r) {
	if (r->pos == r->len) {
		fatalf("decoder: buffer overflow\n");
	}
	return r->buf[r->pos++];
}

static inline uint32_t read_uint32(struct reader *r) {
	return (read_byte(r) << 24) | (read_byte(r) << 16) | (read_byte(r) << 8) | read_byte(r);
}

static inline uint64_t read_uint64(struct reader *r) {
	return (((uint64_t) read_uint32(r)) << 32) | read_uint32(r);
}

static inline uint8_t *read_bytes(struct reader *r, size_t len) {
	uint8_t *b = malloc(sizeof(uint8_t) * len);

	for (int i = 0; i < len; i++) {
		b[i] = read_byte(r);
	}
	return b;
}

static inline char *read_string(struct reader *r, size_t len) {
	char *str = malloc(sizeof(char) * (len + 1));
	str[len] = '\0';

	for (int i = 0; i < len; i++) {
		str[i] = read_byte(r);
	}
	return str;
}

static inline struct bookmark *decode_bookmarks(struct reader *r, size_t len) {
	struct bookmark *bms = malloc(sizeof(struct bookmark) * len);

	for (int i = 0; i < len; i++) {
		bms[i].offset = read_uint32(r);
		bms[i].lineno = read_uint32(r);
		bms[i].pos = read_uint32(r);
		bms[i].len = read_uint32(r);
		bms[i].line = read_string(r, bms[i].len);
	}
	return bms;
}

static inline struct object *decode_objects(struct reader *r, size_t len) {
	struct object *objs = malloc(sizeof(struct object) * len);

	for (int i = 0; i < len; i++) {
		enum obj_type type = read_byte(r);

		switch (type) {
		case obj_null:
			objs[i] = null_obj;
			break;
		case obj_boolean:
			objs[i] = parse_bool(read_byte(r));
			break;
		case obj_integer:
			objs[i] = new_integer_obj(read_uint64(r));
			break;
		case obj_float:
			objs[i] = new_float_obj(read_uint64(r));
			break;
		case obj_string: {
			uint32_t len = read_uint32(r);
			objs[i] = new_string_obj(read_string(r, len), len);
			break;
		}
		case obj_function: {
			uint32_t nparams = read_uint32(r);
			uint32_t nlocals = read_uint32(r);
			uint32_t len = read_uint32(r);
			uint8_t *insts = read_bytes(r, len);
			uint32_t bklen = read_uint32(r);
			struct bookmark *bmarks = decode_bookmarks(r, bklen);
			objs[i] = new_function_obj(insts, len, nlocals, nparams, bmarks, bklen);
			break;
		}
		default:
			fatalf("decoder: unsupported decoding for type %s\n", otype_str(type));
		}
	}
	return objs;
}

inline struct bytecode tau_decode(uint8_t *bytes, size_t len) {
	if (len == 0) {
		fatalf("decoder: empty bytecode");
	}

	struct bytecode bc = (struct bytecode) {0};
	struct reader r = (struct reader) {
		.buf = bytes,
		.len = len,
		.pos = 0
	};

	bc.ndefs = read_uint32(&r);
	bc.len = read_uint32(&r);
	bc.insts = read_bytes(&r, bc.len);
	bc.nconsts = read_uint32(&r);
	bc.consts = decode_objects(&r, bc.nconsts);
	bc.bklen = read_uint32(&r);
	bc.bookmarks = decode_bookmarks(&r, bc.bklen);
	return bc;
}
