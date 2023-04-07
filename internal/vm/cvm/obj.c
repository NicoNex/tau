#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "obj.h"

// static void dummy_dispose(struct object o) {}

// ============================= CLOSURE OBJECT =============================
static void dispose_function_data(struct function *fn);

static void dispose_closure_obj(struct object o) {
	dispose_function_data(o.data.cl->fn);
	free(o.marked);
	free(o.data.cl->free);
	free(o.data.cl);
}

static char *closure_str(struct object o) {
	char *str = calloc(35, sizeof(char));
	sprintf(str, "closure[%p]", o.data.cl->fn);

	return str;
}

void mark_closure_obj(struct object c) {
	*c.marked = 1;
	for (uint32_t i = 0; i < c.data.cl->num_free; i++) {
		mark_obj(c.data.cl->free[i]);
	}
}

struct object new_closure_obj(struct function *fn, struct object *free, size_t num_free) {
	struct closure *cl = malloc(sizeof(struct closure));
	cl->fn = fn;
	cl->free = free;
	cl->num_free = num_free;

	return (struct object) {
		.data.cl = cl,
		.type = obj_closure,
		.marked = MARKPTR(),
	};
}

// ============================= FUNCTION OBJECT =============================
static void dispose_function_data(struct function *fn) {
	for (int i = 0; i < fn->bklen; i++) {
		free(fn->bookmarks[i].line);
	}
	free(fn->bookmarks);
	free(fn->instructions);
	free(fn);
}

static void dispose_function_obj(struct object o) {
	dispose_function_data(o.data.fn);
	free(o.marked);
}

static char *function_str(struct object o) {
	char *str = calloc(35, sizeof(char));
	sprintf(str, "closure[%p]", o.data.fn);

	return str;
}

struct object new_function_obj(uint8_t *insts, size_t len, uint32_t num_locals, uint32_t num_params, struct bookmark *bmarks, uint32_t bklen) {
	struct function *fn = malloc(sizeof(struct function));
	fn->instructions = insts;
	fn->len = len;
	fn->num_locals = num_locals;
	fn->num_params = num_params;
	fn->bookmarks = bmarks;
	fn->bklen = bklen;

	return (struct object) {
		.data.fn = fn,
		.type = obj_function,
		.marked = MARKPTR(),
	};
}

// ============================= BUILTIN OBJECT =============================
static char *builtin_str(struct object o) {
	return strdup("<builtin function>");
}

struct object new_builtin_obj(struct object (*builtin)(struct object *args, size_t len)) {
	return (struct object) {
		.data.builtin = builtin,
		.type = obj_builtin,
		.marked = NULL,
	};
}

// ============================= ERROR OBJECT =============================
static void dispose_error_obj(struct object o) {
	free(o.marked);
	free(o.data.str->str);
	free(o.data.str);
}

static char *error_str(struct object o) {
	return strdup(o.data.str->str);
}

struct object new_error_obj(char *str, size_t len) {
	struct string *s = malloc(sizeof(struct string));
	s->str = str;
	s->len = len;

	return (struct object) {
		.data.str = s,
		.type = obj_error,
		.marked = MARKPTR(),
	};
}

// ============================= FLOAT OBJECT =============================
static char *float_str(struct object o) {
	char *str = calloc(35, sizeof(char));
	sprintf(str, "%f", o.data.f);

	return str;
}

struct object new_float_obj(double val) {
	return (struct object) {
		.data.f = val,
		.type = obj_float,
		.marked = NULL,
	};
}

// ============================= INTEGER OBJECT =============================
static char *integer_str(struct object o) {
	char *str = calloc(30, sizeof(char));

#ifdef __unix__
	sprintf(str, "%ld", o.data.i);
#else
	sprintf(str, "%lld", o.data.i);
#endif

	return str;
}

struct object new_integer_obj(int64_t val) {
	return (struct object) {
		.data.i = val,
		.type = obj_integer,
		.marked = NULL,
	};
}

// ============================= STRING OBJECT =============================
static void dispose_string_obj(struct object o) {
	if (!o.data.str->is_slice) {
		free(o.marked);
		free(o.data.str->str);
	}
	free(o.data.str);
}

static char *string_str(struct object o) {
	return strndup(o.data.str->str, o.data.str->len);
}

struct object new_string_obj(char *str, size_t len) {
	struct string *s = malloc(sizeof(struct string));
	s->str = str;
	s->len = len;
	s->is_slice = 0;

	return (struct object) {
		.data.str = s,
		.type = obj_string,
		.marked = MARKPTR(),
	};
}

struct object new_string_slice(char *str, size_t len, uint32_t *marked) {
	struct string *s = malloc(sizeof(struct string));
	s->str = str;
	s->len = len;
	s->is_slice = 1;

	return (struct object) {
		.data.str = s,
		.type = obj_string,
		.marked = marked,
	};
}

// ============================= GETSETTER OBJECT =============================
static void dispose_getsetter_obj(struct object o) {
	free(o.data.gs);
}

static char *getsetter_str(struct object o) {
	struct getsetter *gs = o.data.gs;
	return object_str(gs->get(gs));
}

struct object new_getsetter_obj(struct object l, struct object r, getfn get, setfn set) {
	struct getsetter *gs = malloc(sizeof(struct getsetter));
	gs->l = l;
	gs->r = r;
	gs->get = get;
	gs->set = set;

	// We shouldn't need the marked field here since the getsetter is freed
	// as soon as it's unwrapped soon after.
	return (struct object) {
		.data.gs = gs,
		.type = obj_getsetter,
		.marked = NULL,
	};
}

struct object object_getsetter_get(struct getsetter *gs) {
	return object_get(gs->l, gs->r.data.str->str);
}

struct object object_getsetter_set(struct getsetter *gs, struct object val) {
	return object_set(gs->l, gs->r.data.str->str, val);
}

struct object map_getsetter_get(struct getsetter *gs) {
	struct map_pair mp = map_get(gs->l, gs->r);
	return mp.val;
}

struct object map_getsetter_set(struct getsetter *gs, struct object val) {
	struct map_pair mp = map_set(gs->l, gs->r, val);
	return mp.val;
}

struct object list_getsetter_get(struct getsetter *gs) {
	struct object *list = gs->l.data.list->list;
	size_t listlen = gs->l.data.list->len;
	int64_t idx = gs->r.data.i;

	if (idx < 0 || idx >= listlen) {
		return new_error_obj(strdup("index out of range"), 18);
	}
	return list[idx];
}

struct object list_getsetter_set(struct getsetter *gs, struct object val) {
	struct object *list = gs->l.data.list->list;
	size_t listlen = gs->l.data.list->len;
	int64_t idx = gs->r.data.i;

	if (idx < 0 || idx >= listlen) {
		return new_error_obj(strdup("index out of range"), 18);
	}
	list[idx] = val;
	return val;
}

// ============================= LIST OBJECT =============================
static void dispose_list_obj(struct object o) {
	if (!o.data.list->is_slice) {
		free(o.marked);
		free(o.data.list->list);
	}
	free(o.data.list);
}

// TODO: optimise this.
static char *list_str(struct object o) {
	size_t len = o.data.list->len;
	struct object *list = o.data.list->list;
	char *strings[len];
	size_t string_len = 3;

	for (int i = 0; i < len; i++) {
		char *s = object_str(list[i]);
		strings[i] = s;
		string_len += i < len-1 ? strlen(s) + 2 : strlen(s);
	}

	char *str = calloc(string_len, sizeof(char));
	str[0] = '[';

	for (int i = 0; i < len; i++) {
		strcat(str, strings[i]);
		if (i < len-1) strcat(str, ", ");
		free(strings[i]);
	}
	strcat(str, "]");

	return str;
}

void mark_list_obj(struct object l) {
	*l.marked = 1;
	for (uint32_t i = 0; i < l.data.list->len; i++) {
		mark_obj(l.data.list->list[i]);
	}
}

struct object new_list_obj(struct object *list, size_t len) {
	struct list *l = malloc(sizeof(struct list));
	l->list = list;
	l->len = len;
	l->cap = len;
	l->is_slice = 0;

	return (struct object) {
		.data.list = l,
		.type = obj_list,
		.marked = MARKPTR(),
	};
}

struct object new_list_slice(struct object *list, size_t len, uint32_t *marked) {
	struct list *l = malloc(sizeof(struct list));
	l->list = list;
	l->len = len;
	l->cap = len;
	l->is_slice = 1;

	return (struct object) {
		.data.list = l,
		.type = obj_list,
		.marked = marked,
	};
}

// ============================= STATIC OBJECTS =============================
static char *boolean_str(struct object o) {
	return o.data.i ? strdup("true") : strdup("false");
}

inline __attribute__((always_inline))
struct object parse_bool(uint32_t b) {
	return b ? true_obj : false_obj;
}

struct object true_obj = (struct object) {
	.data.i = 1,
	.type = obj_boolean,
	.marked = NULL,
};

struct object false_obj = (struct object) {
	.data.i = 0,
	.type = obj_boolean,
	.marked = NULL,
};

struct object null_obj = (struct object) {
	.data.i = 0,
	.type = obj_null,
	.marked = NULL,
};

char *otype_str(enum obj_type t) {
	static char *strings[] = {
		"null",
		"bool",
		"int",
		"float",
		"builtin",
		"string",
		"error",
		"list",
		"map",
		"function",
		"closure",
		"object",
		"pipe",
		"bytes",
		"getsetter"
	};

	return strings[t];
}

char *object_str(struct object o) {
	switch (o.type) {
	case obj_null:
		return strdup("null");
	case obj_boolean:
		return boolean_str(o);
	case obj_integer:
		return integer_str(o);
	case obj_float:
		return float_str(o);
	case obj_builtin:
		return builtin_str(o);
	case obj_string:
		return string_str(o);
	case obj_error:
		return error_str(o);
	case obj_list:
		return list_str(o);
	case obj_map:
		return map_str(o);
	case obj_function:
		return function_str(o);
	case obj_closure:
		return closure_str(o);
	case obj_object:
		return object_obj_str(o);
	case obj_pipe:
		return strdup("<unimplemented pipe>");
	case obj_bytes:
		return strdup("<unimplemented bytes>");
	case obj_getsetter:
		return getsetter_str(o);
	default:
		return strdup("<unimplemented>");
	}
}

void print_obj(struct object o) {
	char *str = object_str(o);
	puts(str);
	free(str);
}

inline void mark_obj(struct object o) {
	if (o.type > obj_builtin) {
		switch (o.type) {
		case obj_object:
			mark_object_obj(o);
			break;
		case obj_list:
			mark_list_obj(o);
			break;
		case obj_closure:
			mark_closure_obj(o);
			break;
		case obj_map:
			mark_map_obj(o);
			break;
		default:
			*o.marked = 1;
			break;
		}
	}
}

void free_obj(struct object o) {
	switch (o.type) {
	case obj_string:
		dispose_string_obj(o);
		return;
	case obj_error:
		dispose_error_obj(o);
		return;
	case obj_list:
		dispose_list_obj(o);
		return;
	case obj_map:
		dispose_map_obj(o);
		return;
	case obj_function:
		dispose_function_obj(o);
		return;
	case obj_closure:
		dispose_closure_obj(o);
		return;
	case obj_object:
		dispose_object_obj(o);
		return;
	case obj_pipe:
		puts("no free function for pipe");
		return;
	case obj_bytes:
		puts("no free function for bytes");
		return;
	case obj_getsetter:
		dispose_getsetter_obj(o);
		return;
	default:
		return;
	}
}
