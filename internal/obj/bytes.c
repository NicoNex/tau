#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "object.h"

void dispose_bytes_obj(struct object o) {
	// A slice doesn't own the buffer, its owner frees it.
	if (o.data.bytes->owner == NULL) {
		free(o.data.bytes->bytes);
	}
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
	return new_bytes_slice(bytes, len, NULL);
}

void mark_bytes_obj(struct object b) {
	obj_gc(b)->mark |= GC_MARK;
	mark_owner(b.data.bytes->owner);
}

struct object new_bytes_slice(uint8_t *bytes, size_t len, struct gc_header *owner) {
	struct gc_header *h = gc_alloc(sizeof(struct bytes));
	struct bytes *b = GC_PAYLOAD(h);
	b->bytes = bytes;
	b->len = len;
	b->owner = owner;

	h->obj = (struct object) {
		.data.bytes = b,
		.type = obj_bytes,
	};
	return h->obj;
}
