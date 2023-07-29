#pragma once

#include "bdwgc/include/gc.h"

#define malloc(n) GC_MALLOC(n)
#define calloc(m, n) GC_MALLOC((m)*(n))
#define realloc(n, m) GC_REALLOC(n, m)
#define free(n) GC_FREE(n)
#define strdup(s) GC_STRDUP(s)
#define strndup(s, n) GC_STRNDUP(s, n)
