DIR := $(shell pwd)
GCC := $(shell which gcc)
DEFAULT_CC = $(CC)

CFLAGS = -g -Ofast -I$(DIR)/internal/obj/libffi/include -mtune=native -march=native -ftree-vectorize
LDFLAGS = -L$(DIR)/internal/obj/libffi/lib $(DIR)/internal/obj/libffi/lib/libffi.a -lm -L$(DIR)/internal/vm/bdwgc/lib $(DIR)/internal/vm/bdwgc/lib/libgc.a

LIBFFI_CONFIGURE_FLAGS = --prefix=$(DIR)/internal/obj/libffi --disable-shared --enable-static --disable-multi-os-directory --disable-docs

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
ifeq ($(UNAME_S),Windows_NT)
	LIBFFI_CONFIGURE_FLAGS += CC="$(DIR)/libffi/msvcc.sh -m64" CXX="$(DIR)/libffi/msvcc.sh -m64"
endif

ifneq ($(origin CC), undefined)
	ifneq ($(CC),clang)
		ifneq ($(GCC),)
			CFLAGS += -fopenmp
			LDFLAGS += -fopenmp
			CC = $(GCC)
		endif
	endif
endif

.PHONY: all tau libffi install profile run bdwgc

all: libffi bdwgc tau

tau:
	cd cmd/tau && \
	CC=$(CC) CGO_CFLAGS="$(CFLAGS)" CGO_LDFLAGS="$(LDFLAGS)" go build -o $(DIR)/tau

libffi:
	if [ ! -d libffi ] || [ $$(ls -1q libffi | wc -l) -eq 0 ]; then \
	    git submodule init; \
	    git submodule update --recursive; \
	fi

	CC=$(CC) cd libffi && \
	ACLOCAL_PATH=$(ACLOCAL_PATH) autoreconf -i && \
	./configure --prefix=$(DIR)/internal/obj/libffi --disable-shared --enable-static --disable-multi-os-directory && \
	make install CC=$(CC)

bdwgc:
	cd bdwgc && \
	./autogen.sh && \
	./configure --prefix=$(DIR)/internal/vm/bdwgc --disable-shared --enable-static --disable-docs && \
	make install

debug:
	cd cmd/tau && \
	CC=$(CC) CGO_CFLAGS="$(CFLAGS) -DDEBUG -DGC_DEBUG" CGO_LDFLAGS="$(LDFLAGS)" go build -o $(DIR)/tau

install: all
	mv tau $(INSTALL_PATH)

profile:
	CC=$(CC) CGO_CFLAGS="$(CFLAGS)" CGO_LDFLAGS="$(LDFLAGS)" go build profile.go

test:
	CC=$(CC) CGO_CFLAGS="$(CFLAGS) -DDEBUG -DGC_DEBUG" CGO_LDFLAGS="$(LDFLAGS)" go test ./...

run: all
	./tau