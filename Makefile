DIR := $(shell pwd)
GCC := $(shell which gcc)
DEFAULT_CC = $(CC)

CFLAGS = -g -Ofast -I$(DIR)/internal/obj/libffi/include
LDFLAGS = -L$(DIR)/internal/obj/libffi/lib $(DIR)/internal/obj/libffi/lib/libffi.a -lm

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    ACLOCAL_PATH := /usr/share/aclocal
    INSTALL_PATH := /usr/bin
endif
ifeq ($(UNAME_S),Darwin)
    ACLOCAL_PATH := /usr/local/share/aclocal
    INSTALL_PATH := /usr/local/bin
    GCC := $(shell which gcc-14)
endif

# Check if CC is defined
ifneq ($(origin CC), undefined)
    # Check if CC is not clang
    ifneq ($(CC), clang)
        # Check if the compiler is actually GCC by looking for "GCC" in the version output
        GCC_CHECK := $(shell $(CC) --version 2>/dev/null | head -n 1 | grep -i "gcc")
        ifneq ($(GCC_CHECK),)
            CFLAGS += -fopenmp
            LDFLAGS += -fopenmp
        endif
    endif
endif

# Default compiler fallback to GCC if GCC environment variable is set
ifneq ($(GCC),)
    CC = $(GCC)
endif

.PHONY: all tau tau-lsp libffi plugins syscall math install uninstall clean fmt profile test run

# Where install puts things. PREFIX=/usr/local make install for a system wide
# one.
PREFIX ?= $(HOME)/.local

all: libffi plugins tau

libffi:
	if [ ! -d libffi ] || [ $$(ls -1q libffi | wc -l) -eq 0 ]; then \
	    git submodule init; \
	    git submodule update --recursive; \
	fi

	CC=$(CC) cd libffi && \
	ACLOCAL_PATH=$(ACLOCAL_PATH) autoreconf -i && \
	./configure --prefix=$(DIR)/internal/obj/libffi --disable-shared --enable-static --disable-multi-os-directory && \
	make install CC=$(CC)

libffi-windows:
	if [ ! -d libffi ] || [ $$(ls -1q libffi | wc -l) -eq 0 ]; then \
	    git submodule init; \
	    git submodule update --recursive; \
	fi

	CC=$(CC) cd libffi && \
	ACLOCAL_PATH=$(ACLOCAL_PATH) autoreconf -i && \
	./configure --host=x86_64-w64-mingw32 --prefix=$(DIR)/internal/obj/libffi --disable-shared --enable-static --disable-multi-os-directory && \
	make install CC=x86_64-w64-mingw32-gcc AR=x86_64-w64-mingw32-ar RANLIB=x86_64-w64-mingw32-ranlib

tau:
	cd cmd/tau && \
	CC=$(CC) \
	CGO_CFLAGS="$(CFLAGS)" \
	CGO_LDFLAGS="$(LDFLAGS)" \
	go build -o $(DIR)/tau

tau-windows:
	cd cmd/tau && \
	CC=x86_64-w64-mingw32-gcc \
	RANLIB=x86_64-w64-mingw32-ranlib \
	CGO_ENABLED=1 \
	CGO_CFLAGS="$(CFLAGS)" \
	CGO_LDFLAGS="$(LDFLAGS)" \
	GOOS=windows \
	GOARCH=amd64 \
	go build -o $(DIR)/tau.exe

windows: libffi-windows tau-windows

debug:
	cd cmd/tau && \
	CC=$(CC) CGO_CFLAGS="$(CFLAGS) -DDEBUG" CGO_LDFLAGS="$(LDFLAGS)" go build -o $(DIR)/tau

gc-debug:
	cd cmd/tau && \
	CC=$(CC) CGO_CFLAGS="$(CFLAGS) -DGC_DEBUG" CGO_LDFLAGS="$(LDFLAGS)" go build -o $(DIR)/tau

# The shared objects the stdlib opens with plugin(). One directory each,
# added here when a new one shows up.
PLUGINS = syscall math

plugins:
	for p in $(PLUGINS); do $(MAKE) -C stdlib/$$p CC=$(CC) || exit 1; done

# The language server, which speaks LSP over stdin and stdout and uses the
# parser and the formatter of this repo.
tau-lsp:
	go build -o $(DIR)/tau-lsp ./cmd/tau-lsp

syscall:
	$(MAKE) -C stdlib/syscall CC=$(CC)

math:
	$(MAKE) -C stdlib/math CC=$(CC)

# The tests live next to what they test, so they are dropped after the copy
# rather than avoided during it, subdirectories included.
install: tau plugins
	mkdir -p $(PREFIX)/bin $(PREFIX)/lib/tau
	cp tau $(PREFIX)/bin/tau
	cp -r stdlib/. $(PREFIX)/lib/tau
	find $(PREFIX)/lib/tau -name '*_test.tau' -delete
	find $(PREFIX)/lib/tau \( -name 'Makefile' -o -name '*.c' \) -delete

uninstall:
	rm -f $(PREFIX)/bin/tau
	rm -rf $(PREFIX)/lib/tau

clean:
	rm -f tau tau.exe tau-lsp profile
	for p in $(PLUGINS); do $(MAKE) -C stdlib/$$p clean; done
	find . -name '*.tauc' -delete

# The canonical style, over everything that is tau source here.
fmt: tau
	./tau fmt -w stdlib tests examples

profile:
	CC=$(CC) CGO_CFLAGS="$(CFLAGS)" CGO_LDFLAGS="$(LDFLAGS)" go build profile.go

# The Go tests first, then the ones written in tau.
test: tau plugins
	CC=$(CC) CGO_CFLAGS="$(CFLAGS)" CGO_LDFLAGS="$(LDFLAGS)" go test ./internal/... ./cmd/...
	TAUPATH=$(DIR)/stdlib ./tau test stdlib

run: all
	./tau
