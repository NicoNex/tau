#include <dlfcn.h>
#include "libffi/include/ffi.h"
#include "object.h"

struct object native_get(struct object obj, char *name) {
	void *fnptr = dlsym(obj.data.handle, name);
	if (!fnptr) {
		return errorf(dlerror());
	}

	return 
}

struct object new_native()
