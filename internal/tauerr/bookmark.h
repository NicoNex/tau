#pragma once

#include <stdint.h>

struct bookmark {
	int32_t offset;
	int32_t lineno;
	int32_t pos;
	size_t len;
	char *line;
};
