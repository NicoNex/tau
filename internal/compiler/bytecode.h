#pragma once

#include <stdint.h>

struct bytecode {
	uint8_t *insts;
	struct object *consts;
	uint32_t len;
	uint32_t nconsts;
	uint32_t bklen;
	struct bookmark *bookmarks;
	uint32_t ndefs;
};

struct buffer {
	uint8_t *buf;
	uint32_t len;
	uint32_t cap;
};

void free_buffer(struct buffer buf);
struct buffer tau_encode(struct bytecode bc);
struct bytecode tau_decode(uint8_t *bytes, size_t len);
