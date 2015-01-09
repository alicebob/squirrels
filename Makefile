.PHONY: all build test vet lint bench

all: build test build vet

test:
	go test -v

build:
	go build

vet:
	go vet

lint:
	golint

bench: build
	go test -run=none -bench=. -test.benchtime=4s
