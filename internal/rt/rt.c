// The runtime a bundled program is built on, with no Go in it at all.
//
// A bundled executable is a copy of this followed by the program and a trailer
// saying how long it is. Everything the program needs travels compiled, so
// what happens here is only reading, and there is no lexer, parser, syntax
// tree or compiler anywhere in the binary.

// nftw, for the walk that removes the unpacked plugins on the way out.
#define _XOPEN_SOURCE 700

// mkdtemp is a BSD extension: on macOS defining _XOPEN_SOURCE hides it unless
// _DARWIN_C_SOURCE is on too. A no-op elsewhere.
#ifdef __APPLE__
#define _DARWIN_C_SOURCE
#endif

#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/stat.h>
#include <libgen.h>
#include <ftw.h>

#include "../vm/vm.h"
#include "../obj/object.h"
#include "../compiler/bytecode.h"

// Windows has neither POSIX mkdir(path, mode) - mingw's takes a path alone -
// nor setenv; the environment is set through _putenv_s. Shimmed after the
// system headers so mingw's own one-argument mkdir declaration is read first.
#if defined(_WIN32) || defined(WIN32)
#include <direct.h>
#define mkdir(path, mode) _mkdir(path)
static inline int setenv(const char *name, const char *value, int overwrite) {
	(void)overwrite;
	return _putenv_s(name, value);
}
#endif

#define TRAILER_LEN 12
static const char exec_magic[4] = {'T', 'A', 'U', 'X'};
static const char bundle_magic[5] = {'T', 'A', 'U', 'B', 0x03};

__attribute__((noreturn))
static void fatalf(const char *fmt, ...) {
	va_list args;
	va_start(args, fmt);
	vfprintf(stderr, fmt, args);
	va_end(args);
	exit(1);
}

// A module of the bundle: its bytecode, and where in the globals each of the
// names it gives away ended up.
struct rt_export {
	char *name;
	uint32_t idx;
};

struct rt_module {
	char *name;
	struct bytecode bc;
	struct rt_export *exports;
	uint32_t nexports;
};

struct rt_bundle {
	struct rt_module *mods;
	uint32_t nmods;
	struct bytecode bc;
	// Where the shared objects were written, so that they can be removed on
	// the way out.
	char *plugin_dir;
};

// The same shape the writer uses: a big endian uint32 before whatever it
// counts. Reading past the end is a broken file and not something to recover
// from, so it ends the program rather than being threaded back to a caller.
struct reader {
	const uint8_t *buf;
	size_t len;
	size_t pos;
};

static const uint8_t *rt_take(struct reader *r, size_t n) {
	if (r->pos + n > r->len) {
		fatalf("bundle: it ends in the middle of something\n");
	}
	const uint8_t *p = r->buf + r->pos;
	r->pos += n;
	return p;
}

static uint32_t rt_uint32(struct reader *r) {
	const uint8_t *b = rt_take(r, 4);
	return ((uint32_t) b[0] << 24) | ((uint32_t) b[1] << 16) | ((uint32_t) b[2] << 8) | b[3];
}

// A length and that many bytes. The result points into the buffer, which lives
// as long as the program does.
static const uint8_t *rt_blob(struct reader *r, uint32_t *len) {
	*len = rt_uint32(r);
	return rt_take(r, *len);
}

static char *rt_str(struct reader *r) {
	uint32_t len;
	const uint8_t *b = rt_blob(r, &len);
	char *s = malloc(len + 1);

	memcpy(s, b, len);
	s[len] = '\0';
	return s;
}

// Where the shared objects were unpacked, kept for the walk that removes them.
static char *plugin_tmpdir;

static int remove_one(const char *path, const struct stat *st, int type, struct FTW *ftw) {
	(void)st; (void)type; (void)ftw;
	return remove(path);
}

// The directory goes when the program does, however it ends: the exit builtin
// calls exit(), and so does a program that just runs out of instructions.
static void remove_plugins(void) {
	if (plugin_tmpdir == NULL) return;
	nftw(plugin_tmpdir, remove_one, 8, FTW_DEPTH | FTW_PHYS);
}

// writes_plugins puts the shared objects where the loader will find them and
// points TAUPATH at the directory, the way the interpreter does.
static char *write_plugins(struct reader *r) {
	uint32_t n = rt_uint32(r);
	if (n == 0) return NULL;

	char tmpl[] = "/tmp/tau-plugins-XXXXXX";
	char *dir = mkdtemp(tmpl);
	if (dir == NULL) {
		fatalf("bundle: cannot make a directory for the plugins\n");
	}
	dir = strdup(dir);
	plugin_tmpdir = dir;
	atexit(remove_plugins);

	for (uint32_t i = 0; i < n; i++) {
		char *name = rt_str(r);
		uint32_t len;
		const uint8_t *so = rt_blob(r, &len);

		char path[4096];
		snprintf(path, sizeof(path), "%s/%s", dir, name);

		// A name may hold a directory of its own, as an import does.
		char *copy = strdup(path);
		char parent[4096];
		snprintf(parent, sizeof(parent), "%s", dirname(copy));
		if (strcmp(parent, dir) != 0) mkdir(parent, 0755);
		free(copy);

		FILE *f = fopen(path, "wb");
		if (f == NULL || fwrite(so, 1, len, f) != len) {
			fatalf("bundle: cannot write the plugin %s\n", name);
		}
		fclose(f);
		// Executable: it is about to be opened as one.
		chmod(path, 0755);
		free(name);
	}

	const char *old = getenv("TAUPATH");
	if (old != NULL && *old != '\0') {
		char joined[8192];
		snprintf(joined, sizeof(joined), "%s:%s", dir, old);
		setenv("TAUPATH", joined, 1);
	} else {
		setenv("TAUPATH", dir, 1);
	}
	return dir;
}

static struct rt_bundle rt_decode(const uint8_t *raw, size_t len) {
	if (len < sizeof(bundle_magic) || memcmp(raw, bundle_magic, sizeof(bundle_magic)) != 0) {
		fatalf("bundle: this is not a bundle this runtime can read\n");
	}

	struct reader r = {
		.buf = raw + sizeof(bundle_magic),
		.len = len - sizeof(bundle_magic),
		.pos = 0
	};
	struct rt_bundle b = {0};

	b.nmods = rt_uint32(&r);
	b.mods = calloc(b.nmods, sizeof(struct rt_module));

	for (uint32_t i = 0; i < b.nmods; i++) {
		struct rt_module *m = &b.mods[i];

		m->name = rt_str(&r);

		uint32_t bclen;
		const uint8_t *bc = rt_blob(&r, &bclen);
		m->bc = tau_decode((uint8_t *) bc, bclen);

		m->nexports = rt_uint32(&r);
		m->exports = calloc(m->nexports, sizeof(struct rt_export));
		for (uint32_t j = 0; j < m->nexports; j++) {
			m->exports[j].name = rt_str(&r);
			m->exports[j].idx = rt_uint32(&r);
		}
	}

	b.plugin_dir = write_plugins(&r);
	b.bc = tau_decode((uint8_t *) r.buf + r.pos, r.len - r.pos);
	return b;
}

// The bundle of this executable, read from the trailer at its very end.
static uint8_t *rt_payload(const char *self, size_t *len) {
	FILE *f = fopen(self, "rb");
	if (f == NULL) {
		fatalf("bundle: cannot read %s\n", self);
	}

	fseek(f, 0, SEEK_END);
	long size = ftell(f);
	if (size < TRAILER_LEN) {
		fatalf("tau-rt: this is the runtime a bundled program is built on, it carries no program of its own\n");
	}

	uint8_t trailer[TRAILER_LEN];
	fseek(f, size - TRAILER_LEN, SEEK_SET);
	if (fread(trailer, 1, TRAILER_LEN, f) != TRAILER_LEN) {
		fatalf("bundle: cannot read the end of %s\n", self);
	}
	if (memcmp(trailer, exec_magic, sizeof(exec_magic)) != 0) {
		fatalf("tau-rt: this is the runtime a bundled program is built on, it carries no program of its own\n");
	}

	uint64_t plen = 0;
	for (int i = 0; i < 8; i++) {
		plen = (plen << 8) | trailer[sizeof(exec_magic) + i];
	}
	long start = size - TRAILER_LEN - (long) plen;
	if (plen == 0 || start < 0) {
		fatalf("bundle: the trailer of this executable is not sane\n");
	}

	uint8_t *raw = malloc(plen);
	fseek(f, start, SEEK_SET);
	if (fread(raw, 1, plen, f) != plen) {
		fatalf("bundle: cannot read the program out of %s\n", self);
	}
	fclose(f);

	*len = plen;
	return raw;
}

// The modules of the bundle, which came compiled and are loaded in the order
// they were compiled for, before the program starts. That order is what makes
// the globals and the constants they were given line up.
static struct modtab *loaded;

static void load_modules(struct vm *vm, struct rt_bundle *b) {
	// Set before the loop and not after it: a module imports the ones that
	// came before it, and it is running while the loop is still going.
	loaded = vm->state.mods;

	for (uint32_t i = 0; i < b->nmods; i++) {
		struct rt_module *m = &b->mods[i];

		struct vm *tvm = new_vm_with_state(strdup(m->name), m->bc, vm->state);
		if (vm_run(tvm) != 0) {
			fatalf("import: %s failed to load\n", m->name);
		}
		vm->state = tvm->state;

		struct object mod = new_object();
		for (uint32_t j = 0; j < m->nexports; j++) {
			struct object o = vm->state.globals->list[m->exports[j].idx];

			if (o.type == obj_object) {
				object_set(mod, m->exports[j].name, object_to_module(o));
			} else {
				object_set(mod, m->exports[j].name, o);
			}
		}
		modtab_put(vm->state.mods, m->name, mod);
		vm_dispose(tvm);
	}
}

// The loader an import reaches. Every module came with the program and was
// loaded before it started, so all there is to do is find one.
int vm_exec_load_module(struct vm *vm, char *path) {
	if (path == NULL || *path == '\0') {
		go_vm_errorf(vm, "import: no file provided");
		return 1;
	}

	struct object mod;
	if (loaded != NULL && modtab_get(loaded, path, &mod)) {
		vm->stack[vm->sp++] = mod;
		return 0;
	}

	char msg[512];
	snprintf(msg, sizeof(msg), "import: no module named \"%s\" came with this program", path);
	go_vm_errorf(vm, msg);
	return 1;
}

// The command line, handed to the program in the environment the way the
// interpreter hands it: the os module is the only thing that knows these
// variables exist, and it clears them once read.
static void set_args(int argc, char **argv) {
	char buf[32], num[32];

	snprintf(num, sizeof(num), "%d", argc);
	setenv("TAU_ARGC", num, 1);
	for (int i = 0; i < argc; i++) {
		snprintf(buf, sizeof(buf), "TAU_ARG%d", i);
		setenv(buf, argv[i], 1);
	}
}

int main(int argc, char **argv) {
	set_exit();
	set_args(argc, argv);

	size_t len;
	uint8_t *raw = rt_payload(argv[0], &len);
	struct rt_bundle b = rt_decode(raw, len);

	struct vm *vm = new_vm(strdup(argv[0]), b.bc);
	load_modules(vm, &b);

	int ret = vm_run(vm);
	fflush(stdout);
	return ret == 0 ? 0 : 1;
}
