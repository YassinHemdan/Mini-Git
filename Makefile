.DEFAULT_GOAL := all
PROJECT_NAME := "jit"
BINARY_NAME := "jit"

.PHONY: all clean test test-verbose run build help

all: clean build test

clean:
	@echo "Cleaning ..."
	@rm -f $(BINARY_NAME) || true
	@rm -r .jit/ || true

test: build
	go test ./...

build:
	go build -o $(BINARY_NAME) main.go
	chmod +x $(BINARY_NAME)

check:
	@./jit commit -m "first commit"