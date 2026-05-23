.PHONY: build run vet http-smoke

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

vet:
	go vet ./...

http-smoke:
	go run ./cmd/http-smoke
