#include <stdlib.h>
#include <stdio.h>
#include "../vm/thrd.h"
#include "object.h"

// Waiting for the mutex must not stall the collector: a thread blocked here
// while the lock owner waits for the world to restart would deadlock the GC.
static void pipe_lock(struct pipe *p) {
	gc_park();
	mtx_lock(&p->mu);
	gc_unpark();
}

int pipe_close(struct object pipe) {
	struct pipe *p = pipe.data.pipe;
	if (p->is_closed) {
		return 0;
	}

	pipe_lock(p);
	// Set the flag to indicate that the pipe is closed.
	p->is_closed = 1;
	// Unblock all the threads waiting on the pipe.
	cnd_broadcast(&p->not_empty);
	cnd_broadcast(&p->not_full);
	mtx_unlock(&p->mu);

	// The buffer and the mutex are freed by dispose_pipe_obj, when the
	// collector proves nobody can reach the pipe anymore.
	return 1;
}

void dispose_pipe_obj(struct object pipe) {
	struct pipe *p = pipe.data.pipe;

	// No locking here: the collector proved that nobody else can reach it.
	p->is_closed = 1;
	free(p->buf);
	mtx_destroy(&p->mu);
	cnd_destroy(&p->not_empty);
	cnd_destroy(&p->not_full);
	free(pipe.marked);
	free(p);
}

void mark_pipe_obj(struct object pipe) {
	struct pipe *p = pipe.data.pipe;

	// The buffer is a ring: the values start at head, not at 0.
	for (uint32_t i = 0; i < p->len; i++) {
		mark_obj(p->buf[(p->head + i) % p->cap]);
	}
	*pipe.marked |= GC_MARK;
}

int pipe_send(struct object pipe, struct object o) {
	struct pipe *p = pipe.data.pipe;
	if (p->is_closed) {
		return 0;
	}

	pipe_lock(p);
	if (p->is_buffered) {
		while (p->len == p->cap && !p->is_closed) {
			gc_park();
			cnd_wait(&p->not_full, &p->mu);
			gc_unpark();
		}
		if (p->is_closed) {
			mtx_unlock(&p->mu);
			return 0;
		}
	} else {
		if (p->len == p->cap) {
			// Grow unwrapping the ring, a plain realloc would scramble it.
			size_t cap = p->cap * 2;
			struct object *buf = malloc(cap * sizeof(struct object));

			for (size_t i = 0; i < p->len; i++) {
				buf[i] = p->buf[(p->head + i) % p->cap];
			}
			free(p->buf);
			p->buf = buf;
			p->cap = cap;
			p->head = 0;
			p->tail = p->len;
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

	pipe_lock(p);
	while (p->len == 0 && !p->is_closed) {
		// Parked: the collector must not wait for a thread that is sleeping
		// on a pipe, and this thread isn't touching any object meanwhile.
		gc_park();
		cnd_wait(&p->not_empty, &p->mu);
		gc_unpark();
	}

	// Values already in the buffer are still delivered after close.
	if (p->len == 0 && p->is_closed) {
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
	pipe->is_closed = 0;
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
	pipe->is_closed = 0;
	mtx_init(&pipe->mu, mtx_plain);
	cnd_init(&pipe->not_empty);
	cnd_init(&pipe->not_full);

	return (struct object) {
		.data.pipe = pipe,
		.type = obj_pipe,
		.marked = MARKPTR()
	};
}
