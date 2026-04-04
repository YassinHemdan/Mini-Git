.DEFAULT_GOAL := all
PROJECT_NAME := "jit"
BINARY_NAME := "jit"
BIN_DIR := bin

.PHONY: all clean test test-verbose run build help

all: clean build test

clean:
	@echo "Cleaning ..."
	@rm -rf $(BIN_DIR) || true
	@rm -rf .jit/ || true

test: build
	go test ./...

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) main.go
	chmod +x $(BIN_DIR)/$(BINARY_NAME)

check:
	@./$(BIN_DIR)/jit add file.txt