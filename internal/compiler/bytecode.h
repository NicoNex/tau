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
