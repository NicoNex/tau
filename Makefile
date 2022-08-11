.PHONY: all

all:
	cd cmd/tau && go build -o ../../tau

windows:
	cd cmd/tau && GOOS=windows go build -o ../../tau
