.PHONY: build install clean test run-gui run-cli

# Binary name and directories
BINARY_NAME=resignipa
BIN_DIR=bin
BUILD_DIR=build

# Build the binary in bin directory
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME) main.go
	@echo "Build complete: ./$(BIN_DIR)/$(BINARY_NAME)"

# Build for multiple architectures in build directory
build-all:
	@echo "Building for multiple architectures..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-arm64 main.go
	@echo "Build complete for all architectures"
	@echo "Binaries location: ./$(BUILD_DIR)/"

# Install to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	rm -rf $(BUILD_DIR)
	rm -rf Resigned/
	rm -rf tmp/
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run GUI mode
run-gui: build
	@echo "Launching GUI..."
	./$(BIN_DIR)/$(BINARY_NAME)

# Run CLI mode (example)
run-cli: build
	@echo "Running CLI mode..."
	@echo "Binary location: ./$(BIN_DIR)/$(BINARY_NAME)"
	@echo "Usage: ./$(BIN_DIR)/$(BINARY_NAME) -s /path/to/app.ipa -c \"Certificate Name\""

# Run setup wizard
setup: build
	@echo "Running setup wizard..."
	./$(BIN_DIR)/$(BINARY_NAME) setup

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies downloaded"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Formatting complete"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Display help
help:
	@echo "ResignIPA Makefile Commands:"
	@echo ""
	@echo "  make build      - Build the binary"
	@echo "  make build-all  - Build for all architectures (amd64 + arm64)"
	@echo "  make install    - Install binary to /usr/local/bin"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make run-gui    - Launch GUI mode"
	@echo "  make run-cli    - Show CLI usage example"
	@echo "  make deps       - Download dependencies"
	@echo "  make fmt        - Format code"
	@echo "  make lint       - Run linter"
	@echo "  make help       - Show this help message"

