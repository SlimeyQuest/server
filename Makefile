.PHONY: build run vet proto-lint proto-gen

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

vet:
	go vet ./...

proto-lint:
	cd ../proto && buf lint

proto-gen:
	cd ../proto && buf generate
