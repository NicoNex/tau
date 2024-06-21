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
    GCC := $(shell which gcc-13)
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

.PHONY: all tau libffi install profile run

all: libffi tau

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

install:
	mkdir -p ~/.local/bin
	mkdir -p ~/.local/lib/tau
	cp tau ~/.local/bin/tau
	cp -r stdlib/* ~/.local/lib/tau

profile:
	CC=$(CC) CGO_CFLAGS="$(CFLAGS)" CGO_LDFLAGS="$(LDFLAGS)" go build profile.go

test:
	CC=$(CC) CGO_CFLAGS="$(CFLAGS) -DDEBUG -DGC_DEBUG" CGO_LDFLAGS="$(LDFLAGS)" go test ./...

run: all
	./tau
