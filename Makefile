.PHONY: all

all:
	cd cmd/tau && go build -o ../../tau

windows:
	cd cmd/tau && GOOS=windows go build -o ../../tau

linux:
	cd cmd/tau && GOOS=linux go build -o ../../tau

darwin:
	cd cmd/tau && GOOS=darwin go build -o ../../tau

install: all
	mv tau /usr/bin

run: all
	./tau
