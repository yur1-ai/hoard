VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test lint clean run

build:
	go build $(LDFLAGS) -o bin/hoard ./cmd/hoard

run: build
	./bin/hoard

test:
	go test ./... -race -count=1

lint:
	golangci-lint run

clean:
	rm -rf bin/
