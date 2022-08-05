.PHONY: build

GOCMD=go

all: build

build: plugin.so

plugin.so:
	$(GOCMD) build -buildmode=plugin -o plugin.so
