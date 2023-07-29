DIR := $(shell pwd)

.PHONY: all libffi bdwgc debug install profile run

all: libffi bdwgc
	cd cmd/tau && go build -o $(DIR)/tau

libffi:
	cd libffi && \
	ACLOCAL_PATH=/usr/share/aclocal autoreconf -i && \
	./configure --prefix=$(DIR)/internal/obj/libffi --disable-shared --enable-static && \
	make install

bdwgc:
	cd bdwgc && \
	./autogen.sh && \
	./configure --prefix=$(DIR)/internal/vm/bdwgc --disable-shared --enable-static && \
	make install

debug: CGO_CFLAGS='-DDEBUG' all

install: all
	mv tau /usr/bin

profile:
	@go build profile.go
	@./profile
	@rm profile

run: all
	./tau
