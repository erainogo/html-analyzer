BIN?=html-analyzer
REGISTRY?=eranga567
TAG?=latest
WEB_IMAGE_NAME=$(REGISTRY)/$(BIN):$(TAG)-web
CLI_IMAGE_NAME=$(REGISTRY)/$(BIN):$(TAG)-cli

lint:
	golangci-lint run -c .golangci.yml --sort-results

test:
	GO111MODULE=on go test ./... -tags musl -coverprofile=coverage.txt -covermode count

web-build:
	GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o build/web-${BIN} ./cmd/server

cli-build:
	GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o build/cli-${BIN} ./cmd/cli

build-mocks:
	cd mocks/ && rm -rf -- */ && mockery --all

clean:
	go clean
	rm -rf build

docker-web-build: web-build
	docker build -f web.Dockerfile -t $(WEB_IMAGE_NAME) .

docker-cli-build: cli-build
	docker build -f cli.Dockerfile -t $(CLI_IMAGE_NAME) .

docker-web-push:
	docker push $(WEB_IMAGE_NAME)

docker-cli-push:
	docker push $(CLI_IMAGE_NAME)