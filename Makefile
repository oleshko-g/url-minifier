.DEFAULT_GOAL := build

.PHONY: fmt vet test build

test: vet
	go fmt
	go vet
	revive ./...
	go generate ./...
	go test ./...

build:
	go build

fullbuild: test
	go build

heap:
	go tool pprof -http=:8082 -seconds=15 http://localhost:8080/debug/pprof/heap
