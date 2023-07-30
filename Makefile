DIR := $(shell pwd)

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    ACLOCAL_PATH := /usr/share/aclocal
    INSTALL_PATH := /usr/bin
endif
ifeq ($(UNAME_S),Darwin)
    ACLOCAL_PATH := /usr/local/share/aclocal
    INSTALL_PATH := /usr/local/bin
endif

.PHONY: all libffi bdwgc debug install profile run

all: libffi bdwgc
	cd cmd/tau && go build -o $(DIR)/tau

libffi:
	cd libffi && \
	ACLOCAL_PATH=$(ACLOCAL_PATH) autoreconf -i && \
	./configure --prefix=$(DIR)/internal/obj/libffi --disable-shared --enable-static && \
	make install

bdwgc:
	cd bdwgc && \
	./autogen.sh && \
	./configure --prefix=$(DIR)/internal/vm/bdwgc --disable-shared --enable-static && \
	make install

debug: CGO_CFLAGS='-DDEBUG' all

install: all
	mv tau $(INSTALL_PATH)

profile:
	@go build profile.go
	@./profile
	@rm profile

run: all
	./tau