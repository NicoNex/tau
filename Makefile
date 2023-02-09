.PHONY: all

all:
	cd cmd/tau && go build -o ../../tau

install: all
	mv tau /usr/bin

run: all
	./tau
