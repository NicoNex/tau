#pragma once

#define GC_THREADS
#include "bdwgc/include/gc/gc.h"

#define GC_CALLOC(m, n) GC_MALLOC((m)*(n))
