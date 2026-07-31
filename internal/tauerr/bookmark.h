#pragma once

#include <stdint.h>
#include <stddef.h>

struct bookmark {
	int32_t offset;
	int32_t lineno;
	int32_t pos;
	size_t len;
	char *line;
	// The file this offset came from, NULL when it is the one the VM is
	// running. A module made of several files compiles into one stream of
	// instructions, and without this every error in it would be blamed on
	// whichever file the compiler started with.
	char *file;
};
