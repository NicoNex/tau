/*
Facciamo si che plugin() restituisca un oggetto di tipo obj_handle,
la chiamata a funzione avviene tramite il `getsetter`, perchè conosce la stringa
(il nome della funzione C) e può avere delle funzioni custom

L'oggetto restituito (di tipo obj_native) può essere passato alle builtin `int`, `float`, `string` eccetera...
*/

#include <dlfcn.h>
#include "object.h"

struct object native_getsetter_get(struct getsetter *gs) {
	// Pointer to the native function to call.
	void *fnptr = dlsym(gs->l.data.handle, gs->r.data.str->str);
	if (fnptr == NULL) {
		return errorf("no function with name '%s' found", gs->r.data.str->str);
	}
	return (struct object) {
		.data.handle = fnptr,
		.type = obj_native,
		.marked = MARKPTR()
	};
}

struct object native_getsetter_set(struct getsetter *gs, struct object val) {
	return errorf("cannot assign values to type %s", otype_str(gs->l.type));
}