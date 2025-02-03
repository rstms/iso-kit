# Makefile for iso-kit

# Binary names for the two CLI applications
EXTRACT_BIN = isoextract
BUILDER_BIN = isobuilder

# Pattern to test all packages in your module
PKG = ./...

# Get version information from git (if available)
VERSION = $(shell git describe --tags --always --dirty)

.PHONY: all build build-isoextract build-isobuilder run-isoextract run-isobuilder \
        test lint vet fmt mod-tidy coverage clean help

# Default target builds both applications.
all: build

# Build both isoextract and isobuilder.
build: build-isoextract build-isobuilder

# Build the isoextract binary.
build-isoextract:
	@echo "Building isoextract..."
	go build -ldflags "-X main.version=$(VERSION)" -o $(EXTRACT_BIN) cmd/isoextract/main.go

# Build the isobuilder binary.
build-isobuilder:
	@echo "Building isobuilder..."
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILDER_BIN) cmd/isobuilder/main.go

# Run the isoextract binary.
run-isoextract: build-isoextract
	@echo "Running isoextract..."
	./$(EXTRACT_BIN)

# Run the isobuilder binary.
run-isobuilder: build-isobuilder
	@echo "Running isobuilder..."
	./$(BUILDER_BIN)

# Run tests with verbose output.
test:
	@echo "Running tests..."
	go test -v $(PKG)

# Run golangci-lint.
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

# Run go vet for static analysis.
vet:
	@echo "Running go vet..."
	go vet $(PKG)

# Format code using go fmt.
fmt:
	@echo "Running go fmt..."
	go fmt $(PKG)

# Tidy up module dependencies.
mod-tidy:
	@echo "Running go mod tidy..."
	go mod tidy

# Run tests with coverage and print a coverage summary.
coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out $(PKG)
	@go tool cover -func=coverage.out

# Clean up generated binaries and coverage file.
clean:
	@echo "Cleaning up..."
	rm -f $(EXTRACT_BIN) $(BUILDER_BIN) coverage.out

# Display help about available targets.
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@echo "  build             - Build both isoextract and isobuilder"
	@echo "  build-isoextract  - Build the isoextract binary"
	@echo "  build-isobuilder  - Build the isobuilder binary"
	@echo "  run-isoextract    - Run the isoextract binary"
	@echo "  run-isobuilder    - Run the isobuilder binary"
	@echo "  test              - Run tests"
	@echo "  lint              - Run golangci-lint"
	@echo "  vet               - Run go vet"
	@echo "  fmt               - Run go fmt"
	@echo "  mod-tidy          - Run go mod tidy"
	@echo "  coverage          - Run tests with coverage"
	@echo "  clean             - Clean up generated binaries and coverage file"
	@echo "  help              - Display this help message"
