.DEFAULT_GOAL := build

.PHONY: fmt vet test build

test: vet
	go fmt
	go vet
	go test ./...

build:
	go build

fullbuild: test
	go build