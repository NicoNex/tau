#pragma once

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
