#pragma once

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

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
// The directory of the file being run, set by the runtime: a module that
// ships a shared object next to itself is found from any working directory.
extern char *tau_module_dir;

// The directory the interpreter itself is in, set by the runtime. An install
// that carries its own library tree is found through this and nothing else,
// which is what makes one relocatable - and on Windows it is the only way,
// since none of the directories at the bottom of this function are there.
extern char *tau_exec_dir;

static inline void *plugin_search(const char *path) {
	void *handle = dlopen(path, RTLD_LAZY);
	if (handle != NULL) return handle;

	char buf[4096];
	char *taupath = getenv("TAUPATH");

	if (tau_module_dir != NULL) {
		snprintf(buf, sizeof(buf), "%s/%s", tau_module_dir, path);
		if ((handle = dlopen(buf, RTLD_LAZY)) != NULL) return handle;
	}

	if (tau_exec_dir != NULL) {
		snprintf(buf, sizeof(buf), "%s/../lib/tau/%s", tau_exec_dir, path);
		if ((handle = dlopen(buf, RTLD_LAZY)) != NULL) return handle;

		snprintf(buf, sizeof(buf), "%s/lib/tau/%s", tau_exec_dir, path);
		if ((handle = dlopen(buf, RTLD_LAZY)) != NULL) return handle;
	}

	// TAUPATH first, same order the modules are looked up in.
	if (taupath != NULL) {
		char *dirs = strdup(taupath);

		for (char *dir = strtok(dirs, ":"); dir != NULL; dir = strtok(NULL, ":")) {
			snprintf(buf, sizeof(buf), "%s/%s", dir, path);
			if ((handle = dlopen(buf, RTLD_LAZY)) != NULL) {
				free(dirs);
				return handle;
			}
		}
		free(dirs);
	}

	const char *home = getenv("HOME");
	if (home != NULL) {
		snprintf(buf, sizeof(buf), "%s/.local/lib/tau/%s", home, path);
		if ((handle = dlopen(buf, RTLD_LAZY)) != NULL) return handle;
	}

	snprintf(buf, sizeof(buf), "/usr/local/lib/tau/%s", path);
	if ((handle = dlopen(buf, RTLD_LAZY)) != NULL) return handle;

	snprintf(buf, sizeof(buf), "/lib/tau/%s", path);
	return dlopen(buf, RTLD_LAZY);
}

// The stdlib names its plugins with a ".so" suffix, the extension the loader
// ships on Linux. macOS builds them ".dylib" and Windows ".dll", so when the
// literal name is not found, retry the whole search with the native suffix.
// On Linux ".so" is already native, so this is a no-op and the first search
// is the only one that runs.
static inline void *plugin_open(const char *path) {
	void *handle = plugin_search(path);
	if (handle != NULL) return handle;

#if defined(_WIN32) || defined(WIN32)
	const char *ext = ".dll";
#elif defined(__APPLE__)
	const char *ext = ".dylib";
#else
	const char *ext = NULL;
#endif

	size_t n = strlen(path);
	if (ext != NULL && n > 3 && strcmp(path + n - 3, ".so") == 0) {
		char alt[4096];
		snprintf(alt, sizeof(alt), "%.*s%s", (int)(n - 3), path, ext);
		return plugin_search(alt);
	}
	return handle;
}
