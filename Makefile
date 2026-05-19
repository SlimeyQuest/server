.PHONY: build run vet proto-lint proto-gen

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

vet:
	go vet ./...

proto-lint:
	$(MAKE) -C ../proto proto-lint

proto-gen:
	$(MAKE) -C ../proto proto-gen
