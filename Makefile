# The tree's own directory, absolute. On Windows the compiler and linker cgo
# drives are native, not the MSYS ones, and read an MSYS path like /d/a/tau as
# the D: drive's \d\a\tau; cygpath -m gives the mixed form D:/a/tau that both
# the native tools and the MSYS shell understand.
ifeq ($(OS),Windows_NT)
    DIR := $(shell cygpath -m "$$(pwd)")
    # Windows executables end in .exe: gcc and go both write the name that way,
    # so the rules that make and copy them have to say so too.
    EXE := .exe
else
    DIR := $(shell pwd)
    EXE :=
endif

# The version `tau version` prints, taken from the tag the tree is on rather
# than written in the source: a release is a tag and nothing else. A tree
# between two tags says v2.0.15-7-gabc1234, a dirty one says -dirty, and a
# checkout with no tags in it falls back to the commit. Override it (make
# VERSION=v2.1.0) where git is not around, as the release workflow does.
VERSION ?= $(shell git describe --tags --dirty --always 2>/dev/null)
LDFLAGS_VERSION = -X 'github.com/NicoNex/tau.tagVersion=$(VERSION)'
GCC := $(shell which gcc)
DEFAULT_CC = $(CC)

CFLAGS = -g -Ofast -I$(DIR)/internal/obj/libffi/include
LDFLAGS = -L$(DIR)/internal/obj/libffi/lib $(DIR)/internal/obj/libffi/lib/libffi.a -lm

# GNU libtool's bootstrapper. Everywhere it is `libtoolize`; on macOS Homebrew
# ships it as `glibtoolize` so it does not clash with Apple's own libtool.
LIBTOOLIZE := libtoolize

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    ACLOCAL_PATH := /usr/share/aclocal
    INSTALL_PATH := /usr/bin
endif
ifeq ($(UNAME_S),Darwin)
    ACLOCAL_PATH := /usr/local/share/aclocal
    INSTALL_PATH := /usr/local/bin
    LIBTOOLIZE := glibtoolize
    # The default `gcc` on macOS is clang, whose integrated assembler rejects
    # libffi's aarch64 .S (invalid CFI advance_loc). A real GNU gcc is needed,
    # and Homebrew installs it version-suffixed (gcc-16). Pick the newest one
    # on PATH rather than pinning a version that ages out of the formula.
    GCC := $(shell for d in $$(echo $$PATH | tr ':' ' '); do for f in "$$d"/gcc-[0-9]*; do [ -x "$$f" ] && echo "$$f"; done; done 2>/dev/null | grep -E '/gcc-[0-9]+$$' | awk -F- '{print $$NF, $$0}' | sort -n | tail -1 | cut -d' ' -f2-)
endif

# Check if CC is defined

# Default compiler fallback to GCC if GCC environment variable is set
ifneq ($(GCC),)
    CC = $(GCC)
endif



.PHONY: all tau tau-rt tau-lsp libffi plugins syscall install uninstall clean fmt profile test run

# Where install puts things: the binary in PREFIX/bin and everything it opens
# at runtime in PREFIX/lib/tau. The default is the user's own prefix, no root
# needed; PREFIX=/usr/local make install for a system wide one.
PREFIX ?= $(HOME)/.local

all: libffi plugins tau tau-rt

# The static library everything links against. Building it means autoreconf
# and configure, minutes of work, so it is a file target: once it is there
# nothing runs again until it is deleted.
#
# Every target that compiles against it names it as a prerequisite and not
# just `all` does, because a fresh checkout is what CI hands to `make test`,
# and ffi.h is not there until this has run.
LIBFFI_A = $(DIR)/internal/obj/libffi/lib/libffi.a

libffi: $(LIBFFI_A)

$(LIBFFI_A):
	if [ ! -d libffi ] || [ $$(ls -1q libffi | wc -l) -eq 0 ]; then \
	    git submodule init; \
	    git submodule update --recursive; \
	fi

	cd libffi && \
	ACLOCAL_PATH=$(ACLOCAL_PATH) $(LIBTOOLIZE) --copy --force && \
	ACLOCAL_PATH=$(ACLOCAL_PATH) autoreconf -i && \
	CC=$(CC) ./configure --prefix=$(DIR)/internal/obj/libffi --disable-shared --enable-static --disable-multi-os-directory && \
	make install CC=$(CC)

libffi-windows:
	if [ ! -d libffi ] || [ $$(ls -1q libffi | wc -l) -eq 0 ]; then \
	    git submodule init; \
	    git submodule update --recursive; \
	fi

	cd libffi && \
	ACLOCAL_PATH=$(ACLOCAL_PATH) $(LIBTOOLIZE) --copy --force && \
	ACLOCAL_PATH=$(ACLOCAL_PATH) autoreconf -i && \
	CC=$(CC) ./configure --host=x86_64-w64-mingw32 --prefix=$(DIR)/internal/obj/libffi --disable-shared --enable-static --disable-multi-os-directory && \
	make install CC=x86_64-w64-mingw32-gcc AR=x86_64-w64-mingw32-ar RANLIB=x86_64-w64-mingw32-ranlib

tau: $(LIBFFI_A)
	cd cmd/tau && \
	CC=$(CC) \
	CGO_CFLAGS="$(CFLAGS)" \
	CGO_LDFLAGS="$(LDFLAGS)" \
	go build -ldflags "$(LDFLAGS_VERSION)" -o $(DIR)/tau$(EXE)

# The runtime a bundled program is built on: the VM, the objects and the
# bytecode decoder, and nothing else. No lexer, no parser, no syntax tree, no
# compiler, and no Go either, so none of the Go runtime is in it. `tau bundle`
# appends to this rather than to the interpreter, and the executables it writes
# are a tenth of the size for it.
RT_SRC = internal/rt/rt.c \
	internal/vm/vm.c internal/vm/gc.c internal/vm/pool.c \
	internal/compiler/codec.c \
	$(wildcard internal/obj/*.c)

# Windows has no libdl: what dlopen is there comes from plugin.h, which calls
# LoadLibrary, and threads are in the C library itself.
ifeq ($(OS),Windows_NT)
    RT_LIBS =
else
    RT_LIBS = -ldl -lpthread
endif

tau-rt: $(LIBFFI_A)
	$(CC) -DTAU_RT -o $(DIR)/tau-rt$(EXE) $(RT_SRC) $(CFLAGS) $(LDFLAGS) $(RT_LIBS)
	strip $(DIR)/tau-rt$(EXE)

tau-windows:
	cd cmd/tau && \
	CC=x86_64-w64-mingw32-gcc \
	RANLIB=x86_64-w64-mingw32-ranlib \
	CGO_ENABLED=1 \
	CGO_CFLAGS="$(CFLAGS)" \
	CGO_LDFLAGS="$(LDFLAGS)" \
	GOOS=windows \
	GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_VERSION)" -o $(DIR)/tau.exe

windows: libffi-windows tau-windows

debug: $(LIBFFI_A)
	cd cmd/tau && \
	CC=$(CC) CGO_CFLAGS="$(CFLAGS) -DDEBUG" CGO_LDFLAGS="$(LDFLAGS)" go build -ldflags "$(LDFLAGS_VERSION)" -o $(DIR)/tau

gc-debug: $(LIBFFI_A)
	cd cmd/tau && \
	CC=$(CC) CGO_CFLAGS="$(CFLAGS) -DGC_DEBUG" CGO_LDFLAGS="$(LDFLAGS)" go build -ldflags "$(LDFLAGS_VERSION)" -o $(DIR)/tau

# The shared objects the stdlib opens with plugin(). One directory each,
# added here when a new one shows up.
PLUGINS = syscall runtime sync/atomic

plugins:
	for p in $(PLUGINS); do $(MAKE) -C stdlib/$$p CC=$(CC) || exit 1; done

# The language server, which speaks LSP over stdin and stdout and uses the
# parser and the formatter of this repo.
tau-lsp:
	go build -o $(DIR)/tau-lsp ./cmd/tau-lsp

syscall:
	$(MAKE) -C stdlib/syscall CC=$(CC)

# The tests live next to what they test, so they are dropped after the copy
# rather than avoided during it, subdirectories included. libffi comes first
# because the interpreter is linked against it.
# DESTDIR is prepended to every path, so that the tree can be built somewhere
# it is not meant to run from. That is what a package is: the files laid out
# as they will be, rooted anywhere. Empty for an ordinary install.
install: libffi plugins tau tau-rt
	# The library directory goes first, so that a module dropped from the
	# stdlib doesn't stay installed forever.
	rm -rf $(DESTDIR)$(PREFIX)/lib/tau
	mkdir -p $(DESTDIR)$(PREFIX)/bin $(DESTDIR)$(PREFIX)/lib/tau
	cp tau$(EXE) $(DESTDIR)$(PREFIX)/bin/tau$(EXE)
	cp tau-rt$(EXE) $(DESTDIR)$(PREFIX)/lib/tau/tau-rt$(EXE)
	cp -r stdlib/. $(DESTDIR)$(PREFIX)/lib/tau
	find $(DESTDIR)$(PREFIX)/lib/tau -name '*_test.tau' -delete
	find $(DESTDIR)$(PREFIX)/lib/tau \( -name 'Makefile' -o -name '*.c' \) -delete
	@echo
	@echo "tau        $(DESTDIR)$(PREFIX)/bin/tau"
	@echo "stdlib     $(DESTDIR)$(PREFIX)/lib/tau"
	@command -v tau >/dev/null || echo "note: $(PREFIX)/bin is not in your PATH"

uninstall:
	rm -f $(PREFIX)/bin/tau$(EXE)
	rm -rf $(PREFIX)/lib/tau

clean:
	rm -f tau tau.exe tau-lsp profile
	for p in $(PLUGINS); do $(MAKE) -C stdlib/$$p clean; done
	find . -name '*.tauc' -delete

# The canonical style, over everything that is tau source here.
fmt: tau
	./tau fmt -w stdlib tests examples

profile: $(LIBFFI_A)
	CC=$(CC) CGO_CFLAGS="$(CFLAGS)" CGO_LDFLAGS="$(LDFLAGS)" go build profile.go

# The Go tests first, then the ones written in tau.
# tau-rt too: a bundled program is built on it, so a runtime left behind by an
# older build is a decoder that no longer agrees with the encoder.
test: tau tau-rt plugins
	CC=$(CC) CGO_CFLAGS="$(CFLAGS)" CGO_LDFLAGS="$(LDFLAGS)" go test . ./internal/... ./cmd/...
	TAUPATH=$(DIR)/stdlib ./tau test stdlib

run: all
	./tau
