#pragma once

#if __has_include(<threads.h>)
	#include <threads.h>
#elif defined(_WIN32) || defined(WIN32)
	#include <windows.h>

	#define thrd_success 0
	#define thrd_t struct {}

	struct warg {
		void (*fn)(void *);
		void *arg;
	};

	static DWORD WINAPI wrap(void *arg) {
		struct warg *a = arg;
		a->fn(a->arg);
		return 0;
	}

	inline int thrd_create(void *t, void (*fn)(void *), void *arg) {
		struct warg w = (struct warg) {.fn = fn, .arg = arg};
		HANDLE handle = CreateThread(NULL, 0, wrap, &w, 0, NULL);
		int ret = handle == NULL;
		CloseHandle(handle);
		return ret;
	}

#else
	#include <pthread.h>

	#define thrd_t pthread_t
	#define thrd_success 0

	#define mtx_t pthread_mutex_t
	#define mtx_plain NULL
	#define mtx_init pthread_mutex_init

	#define cnd_t pthread_cond_t
	#define cnd_init(arg) pthread_cond_init((arg), NULL)

	#define thrd_create(thrd, fn, arg) pthread_create((thrd), NULL, (fn), (arg))
#endif
