DIR := $(shell pwd)

.PHONY: all libffi profile

all: libffi

libffi:
	cd libffi && \
	ACLOCAL_PATH=/usr/share/aclocal autoreconf -i && \
	./configure --prefix=$(DIR)/internal/obj/libffi --disable-shared --enable-static && \
	make install

all:
	cd cmd/tau && go build -o $(DIR)/tau

debug: CGO_CFLAGS='-DDEBUG' all

install: all
	mv tau /usr/bin

profile:
	@go build profile.go
	@./profile
	@rm profile

run: all
	./tau
