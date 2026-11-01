APP_NAME := ms-cust-onboard-client
BIN_DIR := bin

.PHONY: all build run test clean

all: build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/server ./cmd/server

run: build
	./$(BIN_DIR)/server

test:
	go test -v ./...

clean:
	rm -rf $(BIN_DIR)
