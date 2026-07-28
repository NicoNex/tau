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

	// is_closed is only ever looked at with the mutex held: a check outside it
	// races with the thread that closes, and the answer would be stale by the
	// time the lock is taken anyway.
	pipe_lock(p);
	if (p->is_closed) {
		mtx_unlock(&p->mu);
		return 0;
	}
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
	free(p);
}

void mark_pipe_obj(struct object pipe) {
	struct pipe *p = pipe.data.pipe;

	// The buffer is a ring: the values start at head, not at 0.
	for (uint32_t i = 0; i < p->len; i++) {
		mark_obj(p->buf[(p->head + i) % p->cap]);
	}
	pipe.gc->mark |= GC_MARK;
}

int pipe_send(struct object pipe, struct object o) {
	struct pipe *p = pipe.data.pipe;

	pipe_lock(p);
	// Wait for room. Both kinds of pipe block here: an unbuffered one holds a
	// single value, a buffered one as many as it was made with. Neither grows,
	// a pipe that grows is a queue and never makes a sender wait.
	while (p->len == p->cap && !p->is_closed) {
		gc_park();
		cnd_wait(&p->not_full, &p->mu);
		gc_unpark();
	}
	if (p->is_closed) {
		mtx_unlock(&p->mu);
		return 0;
	}

	p->buf[p->tail] = o;
	p->tail = (p->tail + 1) % p->cap;
	p->len++;
	uint64_t ticket = p->sent++;
	cnd_signal(&p->not_empty);

	// An unbuffered send is a rendezvous: it is done when a receiver has taken
	// the value, not when the value has been put down. Waiting for the buffer
	// to have room again would not do, another sender could take that room
	// first and this one would wait for a receiver that already came.
	if (!p->is_buffered) {
		while (p->recvd <= ticket && !p->is_closed) {
			gc_park();
			cnd_wait(&p->not_full, &p->mu);
			gc_unpark();
		}
	}

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
	p->recvd++;
	// Broadcast, not signal: on not_full wait both the senders that want room
	// and the unbuffered ones that want their value taken, and only the right
	// one can tell that it is its turn.
	cnd_broadcast(&p->not_full);
	mtx_unlock(&p->mu);

	return val;
}

// An unbuffered pipe is one slot deep and hands the value over directly, a
// buffered one holds as many values as it was asked for. The only difference
// between the two is whether a sender waits for a receiver.
static struct object pipe_new(size_t cap, uint32_t is_buffered) {
	struct pipe *pipe = calloc(1, sizeof(struct pipe));

	pipe->buf = calloc(cap, sizeof(struct object));
	pipe->cap = cap;
	pipe->is_buffered = is_buffered;
	mtx_init(&pipe->mu, mtx_plain);
	cnd_init(&pipe->not_empty);
	cnd_init(&pipe->not_full);

	return (struct object) {
		.data.pipe = pipe,
		.type = obj_pipe,
		.gc = gc_header_alloc()
	};
}

struct object new_pipe() {
	return pipe_new(1, 0);
}

struct object new_buffered_pipe(size_t size) {
	return pipe_new(size > 0 ? size : 1, 1);
}
