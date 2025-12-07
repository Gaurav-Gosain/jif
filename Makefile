.PHONY: all build test clean install run help generate-testdata benchmark coverage

# Default target
all: test build

# Build the binary
build:
	@echo "Building jif..."
	@go build -o bin/jif ./cmd/jif
	@echo "Build complete: ./bin/jif"

# Run all tests
test:
	@./test.sh

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	@go test -v -race

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem -run=^$$

# Generate test GIF files
generate-testdata:
	@echo "Generating test GIF files..."
	@go run testdata/generate_test_gifs.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install to GOPATH/bin
install:
	@echo "Installing jif..."
	@go install
	@echo "Installed to $(shell go env GOPATH)/bin/jif"

# Run with sample GIF
run:
	@if [ -f testdata/simple.gif ]; then \
		./bin/jif testdata/simple.gif; \
	else \
		echo "No test GIFs found. Run 'make generate-testdata' first."; \
		exit 1; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Show help
help:
	@echo "JIF - Terminal GIF Viewer"
	@echo ""
	@echo "Available targets:"
	@echo "  make              - Run tests and build"
	@echo "  make build        - Build the binary"
	@echo "  make test         - Run full test suite"
	@echo "  make test-unit    - Run unit tests only"
	@echo "  make coverage     - Generate test coverage report"
	@echo "  make benchmark    - Run performance benchmarks"
	@echo "  make generate-testdata - Generate test GIF files"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make install      - Install to GOPATH/bin"
	@echo "  make run          - Run with sample GIF"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Run linter"
	@echo "  make help         - Show this help"
