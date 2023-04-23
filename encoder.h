#pragma once

#include <stdint.h>
#include "internal/compiler/bytecode.h"

struct buffer {
	uint8_t *buf;
	uint32_t len;
	uint32_t cap;
};

void free_buffer(struct buffer buf);
struct buffer tau_encode(struct bytecode bc);
struct bytecode tau_decode(uint8_t *bytes, size_t len);
