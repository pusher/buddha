BUILD_VERSION=0.3.3
BUILD_REVISION=$(shell git rev-parse HEAD)

.PHONY: stage deps test build clean default

default: stage deps test build

deps:
	go get github.com/stretchr/testify/assert

build:
	go build -o bin/buddha -ldflags "-X main.BuildVersion=$(BUILD_VERSION) -X main.BuildRevision=$(BUILD_REVISION)" cmd/buddha/*.go

test:
	go test -v github.com/pusher/buddha/tcptest
	go test -v github.com/pusher/buddha/cmd/buddha
	go test -v github.com/pusher/buddha

stage:
	@mkdir -p bin

clean:
	@rm -rf bin/
