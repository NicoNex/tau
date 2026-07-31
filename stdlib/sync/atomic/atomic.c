// atomic.c - the reads and writes no other routine can see half of.
//
// Every one of these is a single instruction the processor guarantees, not a
// lock taken and given back: that is the whole point of the file. A counter
// behind a mutex costs a pipe, a wait and a wake; a counter here costs one
// locked add, which is what a hot counter shared by every routine wants.
//
// The cell lives on the tau side, as the eight bytes of a bytes() buffer, and
// only its address comes down here. That is deliberate: the collector owns
// that memory and frees it when the value it belongs to is gone, so nothing
// here has to be freed by hand and forgetting to leaks nothing. bytes(n) is
// calloc underneath, so the address is aligned for an int64_t, which is what
// makes a single instruction possible in the first place.
//
// Everything is sequentially consistent. It is the ordering that needs no
// explanation at the call site, and the weaker ones buy nothing measurable
// next to the cost of the interpreter making the call.

#include <stdint.h>

// at_load reads the cell.
int64_t at_load(void *p) {
	return __atomic_load_n((int64_t *) p, __ATOMIC_SEQ_CST);
}

// at_store writes v into the cell.
void at_store(void *p, int64_t v) {
	__atomic_store_n((int64_t *) p, v, __ATOMIC_SEQ_CST);
}

// at_add adds delta to the cell and returns what it holds afterwards, the way
// Go's Add does: the answer is the caller's own, no other routine was handed
// the same one.
int64_t at_add(void *p, int64_t delta) {
	return __atomic_add_fetch((int64_t *) p, delta, __ATOMIC_SEQ_CST);
}

// at_swap writes v and returns what was there before.
int64_t at_swap(void *p, int64_t v) {
	return __atomic_exchange_n((int64_t *) p, v, __ATOMIC_SEQ_CST);
}

// at_cas writes new if the cell holds old, and answers whether it did. This is
// the one the others could all be built from, and the one a loop that has to
// retry is written with.
int64_t at_cas(void *p, int64_t old, int64_t new) {
	return __atomic_compare_exchange_n(
		(int64_t *) p, &old, new, 0, __ATOMIC_SEQ_CST, __ATOMIC_SEQ_CST
	);
}
