BUILD_VERSION=0.2.3
BUILD_REVISION=$(shell git rev-parse HEAD)

.PHONY: stage test build clean default

default: stage test build

build:
	go build -o bin/buddha -ldflags "-X main.BuildVersion=$(BUILD_VERSION) -X main.BuildRevision=$(BUILD_REVISION)" cmd/*.go

test:
	go test -v github.com/pusher/buddha/tcptest
	go test -v github.com/pusher/buddha/flock
	go test -v github.com/pusher/buddha

stage:
	@mkdir -p bin

clean:
	@rm -rf bin/
