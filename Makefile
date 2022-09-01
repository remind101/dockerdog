.PHONY: all fmt check lint test bin/dockerdog
.DEFAULT_GOAL: all

export CGO_ENABLED=0

all: check bin/dockerdog

fmt:
	go fmt ./...

check: lint test

lint:
	@test -z $(shell gofmt -l . | tee /dev/stderr) || { echo "files above are not go fmt"; exit 1; }
	go vet ./...

test:
	go test ./...

bin/dockerdog:
	go build -ldflags '-extldflags "-static"' -o $@ .
