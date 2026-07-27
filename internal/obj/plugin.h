#pragma once

#include <stdio.h>
#include <stdlib.h>

#if !defined(_WIN32) && !defined(WIN32)
	#include <dlfcn.h>
#else
	#include <windows.h>

	#define RTLD_LAZY NULL
	#define dlopen(path, mode) LoadLibrary((path))
	#define dlclose(handle) FreeLibrary((HMODULE)(handle))
	#define dlsym(handle, name) GetProcAddress((handle), (name))

	inline char *dlerror() {
		DWORD dwError = GetLastError();
		char* lpMsgBuf = NULL;

		if (dwError != 0) {
			FormatMessage(
				FORMAT_MESSAGE_ALLOCATE_BUFFER |  FORMAT_MESSAGE_FROM_SYSTEM |  FORMAT_MESSAGE_IGNORE_INSERTS,
				NULL,
				dwError,
				MAKELANGID(LANG_NEUTRAL, SUBLANG_DEFAULT),
				(LPTSTR) &lpMsgBuf,
				0, 
				NULL
			);
		}
		return lpMsgBuf;
	}
#endif

// Opens a plugin looking for it where the modules live, so that a library
// shipped with the stdlib is found from any working directory. A plain name
// (e.g. "libm.so") is left to the loader of the system.
static inline void *plugin_open(const char *path) {
	void *handle = dlopen(path, RTLD_LAZY);
	if (handle != NULL) return handle;

	const char *home = getenv("HOME");
	char buf[4096];

	if (home != NULL) {
		snprintf(buf, sizeof(buf), "%s/.local/lib/tau/%s", home, path);
		if ((handle = dlopen(buf, RTLD_LAZY)) != NULL) return handle;
	}

	snprintf(buf, sizeof(buf), "/lib/tau/%s", path);
	return dlopen(buf, RTLD_LAZY);
}
