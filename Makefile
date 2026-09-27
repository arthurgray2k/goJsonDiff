.PHONY: all build test test-coverage fmt vet clean

BINARY_NAME=gojsondiff

all: build

build:
	go build -o $(BINARY_NAME) ./cmd/gojsondiff

test:
	go test -v ./...

test-coverage:
	go test -v -cover ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY_NAME)
