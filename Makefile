version = $(shell git describe --tags --abbrev=0)
goversion = $(shell go version)

.PHONY: all

all:
	cd cmd/tau && go build -o ../../tau -ldflags="-X 'github.com/NicoNex/tau.TauVersion=$(version)' -X 'github.com/NicoNex/tau.GoVersion=$(goversion)'"

windows:
	cd cmd/tau && GOOS=windows go build -o ../../tau -ldflags="-X 'github.com/NicoNex/tau.TauVersion=$(version)' -X 'github.com/NicoNex/tau.GoVersion=$(goversion)'"

linux:
	cd cmd/tau && GOOS=linux go build -o ../../tau -ldflags="-X 'github.com/NicoNex/tau.TauVersion=$(version)' -X 'github.com/NicoNex/tau.GoVersion=$(goversion)'"

darwin:
	cd cmd/tau && GOOS=darwin go build -o ../../tau -ldflags="-X 'github.com/NicoNex/tau.TauVersion=$(version)' -X 'github.com/NicoNex/tau.GoVersion=$(goversion)'"

install: all
	mv tau /usr/bin

run: all
	./tau
