#pragma once

#include <stdint.h>

struct bookmark {
	uint32_t offset;
	uint32_t lineno;
	uint32_t pos;
	size_t len;
	char *line;
};
