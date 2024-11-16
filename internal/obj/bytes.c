#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "object.h"

void dispose_bytes_obj(struct object o) {
	// Free everything if it's not a slice (marked parent bit is set to NULL).
	if (o.data.str->parent == NULL) {
		free(o.gcdata);
		free(o.data.bytes->bytes);
	}
	free(o.data.bytes);
}

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
	b->parent = NULL;

	return (struct object) {
		.data.bytes = b,
		.type = obj_bytes,
		.gcdata = new_gcdata(),
	};
}

void mark_bytes_obj(struct object b) {
	b.gcdata->marked = 1;
	if (b.data.bytes->parent != NULL) {
		mark_bytes_obj(*b.data.bytes->parent);
	}
}

struct object new_bytes_slice(uint8_t *bytes, size_t len, struct object *parent) {
	struct bytes *b = malloc(sizeof(struct bytes));
	b->bytes = bytes;
	b->len = len;
	b->parent = parent;

	return (struct object) {
		.data.bytes = b,
		.type = obj_bytes,
		.gcdata = new_gcdata(),
	};
}
