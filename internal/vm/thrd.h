#pragma once

#if __has_include(<threads.h>)
	#include <threads.h>
#elif defined(_WIN32) || defined(WIN32)
	#include <windows.h>
	#include <process.h>

	// Thread
	#define thrd_t HANDLE
	#define thrd_success 0
	#define thrd_create(thrd, fn, arg) ((*(thrd) = CreateThread(NULL, 0, (LPTHREAD_START_ROUTINE)(fn), (arg), 0, NULL)) == NULL)

	// Mutex
	#define mtx_t CRITICAL_SECTION
	#define mtx_plain NULL
	#define mtx_init(cs, mode) InitializeCriticalSection((cs))
	#define mtx_lock EnterCriticalSection
	#define mtx_unlock LeaveCriticalSection
	#define mtx_destroy DeleteCriticalSection

	// Condition
	#define cnd_t CONDITION_VARIABLE
	#define cnd_init(arg) InitializeConditionVariable(arg)
	#define cnd_broadcast WakeAllConditionVariable
	#define cnd_signal WakeConditionVariable
	#define cnd_wait(cond, mtx) while (!SleepConditionVariableCS(cond, mtx, INFINITE)) {}
	#define cnd_destroy(cnd)
#else
	#include <pthread.h>

	#ifdef __clang__
		struct warg {
			int (*fn)(void *);
			void *arg;
		};

		static void *wrapper(void *arg) {
			struct warg *a = arg;
			a->fn(a->arg);
			return NULL;
		}
	#endif

	// Thread
	#define thrd_t pthread_t
	#define thrd_success 0
	#ifdef __clang__
		#define thrd_create(thrd, _fn, _arg) ({ \
			struct warg a = (struct warg) {.fn = (_fn), .arg = (_arg)}; \
			pthread_create((thrd), NULL, wrapper, &a); \
		})
	#else
		#define thrd_create(thrd, fn, arg) ({ \
			void *wrapper(void *a) { return (void *)(intptr_t)fn(a); } \
			pthread_create((thrd), NULL, wrapper, (arg)); \
		})
	#endif

	// Mutex
	#define mtx_t pthread_mutex_t
	#define mtx_plain NULL
	#define mtx_init pthread_mutex_init
	#define mtx_lock pthread_mutex_lock
	#define mtx_unlock pthread_mutex_unlock
	#define mtx_destroy pthread_mutex_destroy

	// Condition
	#define cnd_t pthread_cond_t
	#define cnd_init(arg) pthread_cond_init((arg), NULL)
	#define cnd_broadcast pthread_cond_broadcast
	#define cnd_signal pthread_cond_signal
	#define cnd_wait pthread_cond_wait
	#define cnd_destroy pthread_cond_destroy
#endif
