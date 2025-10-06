# Project Information
PROJECT_NAME = iso-kit
VERSION_PKG  = github.com/rstms/iso-kit/pkg/version

# Executables
BINARIES = isoview isoextract isocreate

# Get the current Git branch, short commit hash, and timestamp
BRANCH  := $(shell git rev-parse --abbrev-ref HEAD)
REV     := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# LDFlags for injecting versioning into the binaries
LDFLAGS = -X '$(VERSION_PKG).version=$(VERSION)' \
          -X '$(VERSION_PKG).branch=$(BRANCH)' \
          -X '$(VERSION_PKG).date=$(DATE)' \
          -X '$(VERSION_PKG).revision=$(REV)'

# Default target: Build all binaries
all: build

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy

# Build all executables
build: $(BINARIES)

# Build each binary with versioning
$(BINARIES):
	@echo "Building $@..."
	go build -ldflags "$(LDFLAGS)" -o bin/$@ ./cmd/$@

# Run unit tests
test:
	@echo "Running unit tests..."
	go test ./... -v -coverprofile=coverage.out

# Run integration tests
integration-test:
	@echo "Running integration tests..."
	go test ./tests/integration -v -race

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf bin/* coverage.out

# Install binaries
install: build
	@echo "Installing binaries..."
	install -m 0755 bin/isoview /usr/local/bin/isoview
	install -m 0755 bin/isoextract /usr/local/bin/isoextract
	install -m 0755 bin/isocreate /usr/local/bin/isocreate

# Show build version information
version:
	@echo "Version: $(VERSION)"
	@echo "Branch: $(BRANCH)"
	@echo "Date: $(DATE)"
	@echo "Rev: $(REV)"

.PHONY: all deps build $(BINARIES) test integration-test clean install version
