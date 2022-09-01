version = $(shell git describe --tags --abbrev=0)
goversion = $(shell go version)

.PHONY: all

all:
	cd cmd/tau && go build -o ../../tau -ldflags="-X 'github.com/NicoNex/tau.Version=$(version)' -X 'github.com/NicoNex/tau.GoVersion=$(goversion)'"

windows:
	cd cmd/tau && GOOS=windows go build -o ../../tau -ldflags="-X 'github.com/NicoNex/tau.Version=$(version)' -X 'github.com/NicoNex/tau.GoVersion=$(goversion)'"
