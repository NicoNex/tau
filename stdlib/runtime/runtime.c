// runtime.c - what a program can know about the machine it is running on.
//
// Only what a shared object can answer on its own: the interpreter's own
// numbers, how many tau routines are alive or how full the heap is, are static
// to the interpreter and no plugin can reach them.
//
// The system and the architecture come from the macros the compiler sets and
// not from asking the machine, so they say what this was built for. That is
// what a bundled program needs: it carries this object with it and the answer
// has to keep meaning the same thing wherever it lands.

#include <stdint.h>

#if defined(_WIN32) || defined(WIN32)
	#include <windows.h>
#elif defined(__APPLE__)
	#include <sys/sysctl.h>
	#include <sys/types.h>
#else
	#include <unistd.h>
#endif

// rt_numcpu is how many cores the program may actually run on, which is what
// sizing a pool of tau routines asks for. It answers 1 rather than an error
// when the machine will not say: a pool of one works.
int64_t rt_numcpu(void) {
#if defined(_WIN32) || defined(WIN32)
	SYSTEM_INFO si;
	GetSystemInfo(&si);
	return si.dwNumberOfProcessors > 0 ? (int64_t) si.dwNumberOfProcessors : 1;
#elif defined(__APPLE__)
	int n = 0;
	size_t len = sizeof(n);
	// The ones this process may use, not the ones soldered on.
	if (sysctlbyname("hw.logicalcpu", &n, &len, NULL, 0) == 0 && n > 0) {
		return n;
	}
	return 1;
#else
	long n = sysconf(_SC_NPROCESSORS_ONLN);
	return n > 0 ? (int64_t) n : 1;
#endif
}

// The names are the ones Go uses, because the rest of the standard library is
// shaped after Go and a program that has to tell systems apart is easier to
// read when the names are the expected ones.
const char *rt_os(void) {
#if defined(_WIN32) || defined(WIN32)
	return "windows";
#elif defined(__APPLE__)
	return "darwin";
#elif defined(__linux__)
	return "linux";
#elif defined(__FreeBSD__)
	return "freebsd";
#elif defined(__OpenBSD__)
	return "openbsd";
#elif defined(__NetBSD__)
	return "netbsd";
#else
	return "unknown";
#endif
}

const char *rt_arch(void) {
#if defined(__x86_64__) || defined(_M_X64)
	return "amd64";
#elif defined(__aarch64__) || defined(_M_ARM64)
	return "arm64";
#elif defined(__i386__) || defined(_M_IX86)
	return "386";
#elif defined(__arm__) || defined(_M_ARM)
	return "arm";
#elif defined(__riscv) && __riscv_xlen == 64
	return "riscv64";
#else
	return "unknown";
#endif
}
