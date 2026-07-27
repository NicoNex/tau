// math.c - the C math library, reachable from tau.
//
// A native call always brings back a machine word, so a double comes back as
// its bits and float(x, 64) on the tau side reads them as the number they
// are.

#include <stdint.h>
#include <string.h>
#include <math.h>

static inline int64_t bits(double d) {
    int64_t i;
    memcpy(&i, &d, sizeof i);
    return i;
}

int64_t m_sqrt(double x) { return bits(sqrt(x)); }
int64_t m_cbrt(double x) { return bits(cbrt(x)); }
int64_t m_exp(double x) { return bits(exp(x)); }
int64_t m_log(double x) { return bits(log(x)); }
int64_t m_log2(double x) { return bits(log2(x)); }
int64_t m_log10(double x) { return bits(log10(x)); }
int64_t m_sin(double x) { return bits(sin(x)); }
int64_t m_cos(double x) { return bits(cos(x)); }
int64_t m_tan(double x) { return bits(tan(x)); }
int64_t m_asin(double x) { return bits(asin(x)); }
int64_t m_acos(double x) { return bits(acos(x)); }
int64_t m_atan(double x) { return bits(atan(x)); }
int64_t m_sinh(double x) { return bits(sinh(x)); }
int64_t m_cosh(double x) { return bits(cosh(x)); }
int64_t m_tanh(double x) { return bits(tanh(x)); }
int64_t m_floor(double x) { return bits(floor(x)); }
int64_t m_ceil(double x) { return bits(ceil(x)); }
int64_t m_round(double x) { return bits(round(x)); }
int64_t m_trunc(double x) { return bits(trunc(x)); }
int64_t m_fabs(double x) { return bits(fabs(x)); }
int64_t m_pow(double x, double y) { return bits(pow(x, y)); }
int64_t m_atan2(double y, double x) { return bits(atan2(y, x)); }
int64_t m_fmod(double x, double y) { return bits(fmod(x, y)); }
int64_t m_hypot(double x, double y) { return bits(hypot(x, y)); }
int64_t m_inf(int64_t sign) { return bits(sign < 0 ? -INFINITY : INFINITY); }
int64_t m_nan(void) { return bits(NAN); }
int64_t m_isnan(double x) { return isnan(x) ? 1 : 0; }
int64_t m_isinf(double x) { return isinf(x) ? 1 : 0; }
