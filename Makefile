.PHONY: build install clean test build-linux run help

# Variables
BINARY_NAME=phppark
VERSION=$(shell git describe --tags --always --dirty)
BUILD_DIR=dist
GO_FILES=$(shell find . -name '*.go' -type f)

# Default target
all: build

# Build for current platform
build: $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/phppark

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux ./cmd/phppark

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run the application
run: build
	@$(BUILD_DIR)/$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	go clean

# Install locally (for development)
install: build
	@echo "Installing to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Show help
help:
	@echo "PHPark Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        Build for current platform"
	@echo "  make build-linux  Build for Linux (amd64)"
	@echo "  make deps         Install dependencies"
	@echo "  make test         Run tests"
	@echo "  make run          Build and run"
	@echo "  make clean        Remove build artifacts"
	@echo "  make install      Install locally"
	@echo "  make help         Show this help"
