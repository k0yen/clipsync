BINARY_NAME=clipsync
MAIN_PATH=./cmd/main.go
BUILD_DIR=./bin

.PHONY: help run build clean fmt

all: build

help:
	@echo "ClipSync E2EE Backend"
	@echo ""
	@echo "Available commands:"
	@echo "  make run              - Run the server"
	@echo "  make build            - Build the binary"
	@echo "  make fmt              - Format code"
	@echo "  make clean            - Clean build artifacts and test databases"

run:
	@go run *.go

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_NAME)"

fmt:
	@go fmt ./...

clean:
	@rm -rf $(BUILD_DIR)
	@rm -f clipsync.db
	@go clean
	@echo "Clean complete"

