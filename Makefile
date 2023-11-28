DIR := $(shell pwd)
GCC := $(shell which gcc)
DEFAULT_CC = $(CC)

CFLAGS = -g -Ofast -I$(DIR)/internal/obj/libffi/include
LDFLAGS = -L$(DIR)/internal/obj/libffi/include -lffi -lm

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

ifneq ($(origin CC), undefined)
	ifneq ($(CC),clang)
		ifneq ($(GCC),)
			CFLAGS += -fopenmp
			LDFLAGS += -fopenmp
			CC = $(GCC)
		endif
	endif
endif

.PHONY: all tau libffi debug install profile run

all: libffi tau

tau:
	cd cmd/tau && \
	CC=$(CC) CGO_CFLAGS="$(CFLAGS)" CGO_LDFLAGS="$(LDFLAGS)" go build -o $(DIR)/tau

libffi:
	cd libffi && \
	ACLOCAL_PATH=$(ACLOCAL_PATH) autoreconf -i && \
	./configure --prefix=$(DIR)/internal/obj/libffi --disable-shared --enable-static && \
	make install CC=$(CC)

debug: CGO_CFLAGS='-DDEBUG' all

install: all
	mv tau /usr/bin

profile:
	CC=$(CC) CGO_CFLAGS="$(CFLAGS)" CGO_LDFLAGS="$(LDFLAGS)" go build profile.go

run: all
	./tau
