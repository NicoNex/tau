#include <stdlib.h>
#include <stdio.h>
#include "../vm/thrd.h"
#include "object.h"

int pipe_close(struct object pipe) {
	struct pipe *p = pipe.data.pipe;
	if (p->is_closed) {
		return 0;
	}

	mtx_lock(&p->mu);
	// Set the flag to indicate that the pipe is closed.
	p->is_closed = 1;
	// Unblock all threads waiting on not_empty.
	cnd_broadcast(&p->not_empty);
	mtx_unlock(&p->mu);

	free(p->buf);
	mtx_destroy(&pipe.data.pipe->mu);
	cnd_destroy(&pipe.data.pipe->not_empty);
	cnd_destroy(&pipe.data.pipe->not_full);
	return 1;
}

void dispose_pipe_obj(struct object pipe) {
	pipe_close(pipe);
	free(pipe.data.pipe);
}

void mark_pipe_obj(struct object pipe) {
	struct pipe *p = pipe.data.pipe;

	for (uint32_t i = 0; i < p->len; i++) {
		mark_obj(p->buf[i]);
	}
	*pipe.marked = 1;
}

int pipe_send(struct object pipe, struct object o) {
	struct pipe *p = pipe.data.pipe;
	if (p->is_closed) {
		return 0;
	}

	mtx_lock(&p->mu);
	if (p->is_buffered) {
		while (p->len == p->cap) {
			cnd_wait(&p->not_full, &p->mu);
		}
	} else {
		if (p->len == p->cap) {
			p->cap *= 2;
			p->buf = realloc(p->buf, p->cap * sizeof(struct object));
		}
	}
	p->buf[p->tail] = o;
	p->tail = (p->tail + 1) % p->cap;
	p->len++;
	cnd_signal(&p->not_empty);
	mtx_unlock(&p->mu);
	return 1;
}

struct object pipe_recv(struct object pipe) {
	struct pipe *p = pipe.data.pipe;

	mtx_lock(&p->mu);
	while (p->len == 0 && !p->is_closed) {
		cnd_wait(&p->not_empty, &p->mu);
	}

	if (p->is_closed) {
		mtx_unlock(&p->mu);
		return null_obj;
	}

	struct object val = p->buf[p->head];
	p->head = (p->head + 1) % p->cap;
	p->len--;
	cnd_signal(&p->not_full);
	mtx_unlock(&p->mu);

	return val;
}

struct object new_pipe() {
	struct pipe *pipe = malloc(sizeof(struct pipe));
	pipe->buf = calloc(1, sizeof(struct object));
	pipe->cap = 1;
	pipe->len = 0;
	pipe->head = 0;
	pipe->tail = 0;
	pipe->is_buffered = 0;
	mtx_init(&pipe->mu, mtx_plain);
	cnd_init(&pipe->not_empty);
	cnd_init(&pipe->not_full);

	return (struct object) {
		.data.pipe = pipe,
		.type = obj_pipe,
		.marked = MARKPTR()
	};
}

struct object new_buffered_pipe(size_t size) {
	struct pipe *pipe = malloc(sizeof(struct pipe));
	pipe->buf = calloc(size, sizeof(struct object));
	pipe->cap = size;
	pipe->len = 0;
	pipe->head = 0;
	pipe->tail = 0;
	pipe->is_buffered = 1;
	mtx_init(&pipe->mu, mtx_plain);
	cnd_init(&pipe->not_empty);
	cnd_init(&pipe->not_full);

	return (struct object) {
		.data.pipe = pipe,
		.type = obj_pipe,
		.marked = MARKPTR()
	};
}
