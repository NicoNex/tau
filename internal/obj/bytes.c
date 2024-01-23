#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "object.h"

char *bytes_str(struct object o) {
	size_t slen = o.data.bytes->len * 5 + 3;
	char *s = calloc(slen, sizeof(char));
	s[0] = '[';

	char tmp[4] = {'\0'};
	size_t blen = o.data.bytes->len;

	for (uint32_t i = 0; i < blen; i++) {
		snprintf(tmp, 4, "%u", o.data.bytes->bytes[i]);
		strcat(s, tmp);
		if (i < blen-1) strcat(s, ", ");
	}
	strcat(s, "]");
	return s;
}

struct object new_bytes_obj(uint8_t *bytes, size_t len) {
	struct bytes *b = malloc(sizeof(struct bytes));
	b->bytes = bytes;
	b->len = len;

	return (struct object) {
		.data.bytes = b,
		.type = obj_bytes,
	};
}

struct object new_bytes_slice(uint8_t *bytes, size_t len) {
	struct bytes *b = malloc(sizeof(struct bytes));
	b->bytes = bytes;
	b->len = len;

	return (struct object) {
		.data.bytes = b,
		.type = obj_bytes,
	};
}
