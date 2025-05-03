BIN?=html-analyzer
REGISTRY?=localhost
TAG?=latest
GIT_SSH_KEY?=~/.ssh/id_rsa
PROJECT_ROOT=$(shell pwd)

default: run
.PHONY : build run fresh test clean build-static docker-build docker-release

lint:
	golangci-lint run -c .golangci.yml --sort-results

test:
	GO111MODULE=on GOPRIVATE="github.com" go test ./... -tags musl -coverprofile=coverage.txt -covermode count

test-dynamic:
	GO111MODULE=on GOPRIVATE="github.com" go test ./... -tags=dynamic,musl --cover

build:
	GO111MODULE=on GOPRIVATE="github.com" go build -o build/${BIN}

build-mocks:
	cd mocks/ && rm -rf -- */ && mockery --all

run: build
	./build/${BIN}

clean:
	go clean
	rm -rf build
