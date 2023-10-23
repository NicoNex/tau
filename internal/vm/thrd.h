#pragma once

#if __has_include(<threads.h>)
	#include <threads.h>
#elif defined(_WIN32) || defined(WIN32)
	#include <windows.h>

	#define thrd_t HANDLE
	#define thrd_success 0

	#define mtx_t CRITICAL_SECTION
	#define mtx_plain NULL
	#define mtx_init InitializeCriticalSection
	#define mtx_lock EnterCriticalSection
	#define mtx_unlock LeaveCriticalSection
	#define mtx_destroy DeleteCriticalSection

	#define cnd_t CONDITION_VARIABLE
	#define cnd_init(arg) InitializeConditionVariable(arg)
	#define cnd_broadcast WakeAllConditionVariable
	#define cnd_signal WakeConditionVariable
	#define cnd_wait(cond, mtx) while (!SleepConditionVariableCS(cond, mtx, INFINITE)) {}
	#define cnd_destroy DeleteConditionVariable

	#define thrd_create(thrd, fn, arg) ((*(thrd) = CreateThread(NULL, 0, (LPTHREAD_START_ROUTINE)(fn), (arg), 0, NULL)) == NULL)

#else
	#include <pthread.h>

	#define thrd_t pthread_t
	#define thrd_success 0

	#define mtx_t pthread_mutex_t
	#define mtx_plain NULL
	#define mtx_init pthread_mutex_init
	#define mtx_lock pthread_mutex_lock
	#define mtx_unlock pthread_mutex_unlock
	#define mtx_destroy pthread_mutex_destroy

	#define cnd_t pthread_cond_t
	#define cnd_init(arg) pthread_cond_init((arg), NULL)
	#define cnd_broadcast pthread_cond_broadcast
	#define cnd_signal pthread_cond_signal
	#define cnd_wait pthread_cond_wait
	#define cnd_destroy pthread_cond_destroy

	#define thrd_create(thrd, fn, arg) pthread_create((thrd), NULL, (fn), (arg))
#endif
