APP_NAME := ms-eligiblity
BIN_DIR := bin

.PHONY: all build run test clean

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/server ./cmd/server

run: build
	./$(BIN_DIR)/server

test:
	go test -v -race -cover ./...

clean:
	rm -rf $(BIN_DIR)
