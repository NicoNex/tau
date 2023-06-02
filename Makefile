.PHONY: all

all:
	cd cmd/tau && go build -o ../../tau

install: all
	mv tau /usr/bin

profile:
	@go build profile.go
	@./profile
	@rm profile

run: all
	./tau
