// The C side of the FFI test: one function per shape a signature can
// describe, so that what tau sends and what it gets back can be checked
// against a real C type and not against a machine word.

#include <stdint.h>
#include <string.h>
#include <stdlib.h>

double t_add_d(double a, double b) { return a + b; }
float t_add_f(float a, float b) { return a + b; }
int32_t t_add_i(int32_t a, int32_t b) { return a + b; }
uint32_t t_add_u(uint32_t a, uint32_t b) { return a + b; }
int8_t t_neg_c(int8_t a) { return -a; }
uint8_t t_not_b(uint8_t a) { return !a; }
int16_t t_add_s(int16_t a, int16_t b) { return a + b; }
int64_t t_add_l(int64_t a, int64_t b) { return a + b; }

static int64_t counter = 0;

void t_bump(int64_t n) { counter += n; }
int64_t t_counter(void) { return counter; }

const char *t_greet(void) { return "ciao"; }
const char *t_nullstr(void) { return NULL; }
int64_t t_strlen(const char *s) { return s != NULL ? (int64_t) strlen(s) : -1; }

// A buffer that lives in the library, so tau has to read it through a
// pointer rather than getting it handed over.
static uint8_t buf[4] = {1, 2, 3, 4};

void *t_buf(void) { return buf; }

// A struct comes back as a pointer, and tau reads its bytes.
struct point {
	int32_t x;
	int32_t y;
};

static struct point pt = {7, -1};

void *t_point(void) { return &pt; }

// An out parameter: the caller passes a buffer and the function fills it.
void t_fill(uint8_t *out, int32_t n, uint8_t v) {
	memset(out, v, n);
}
