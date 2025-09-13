# Assignment Pull Request Creator

.PHONY: build test clean lint fmt vet run help

# Build configurations
BINARY_NAME=assignment-pr-creator
BINARY_PATH=./bin/$(BINARY_NAME)
CMD_PATH=./cmd/assignment-pr-creator

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Default target
all: build

## build: Build the binary
build: clean
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) $(CMD_PATH)
	@echo "✅ Binary built at $(BINARY_PATH)"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf ./bin

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## run: Build and run with dry-run mode
run: build
	@echo "Running $(BINARY_NAME) in dry-run mode..."
	@GITHUB_TOKEN=dummy \
	 GITHUB_REPOSITORY=owner/repo \
	 DRY_RUN=true \
	 $(BINARY_PATH)

## run-live: Build and run in live mode (requires real environment variables)
run-live: build
	@echo "Running $(BINARY_NAME) in live mode..."
	@if [ -z "$(GITHUB_TOKEN)" ] || [ -z "$(GITHUB_REPOSITORY)" ]; then \
		echo "❌ Error: GITHUB_TOKEN and GITHUB_REPOSITORY environment variables are required for live mode"; \
		echo "Usage: GITHUB_TOKEN=your_token GITHUB_REPOSITORY=owner/repo make run-live"; \
		exit 1; \
	fi
	$(BINARY_PATH)

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(CMD_PATH)

## docker-build: Build Docker image (if Dockerfile exists)
docker-build:
	@if [ -f Dockerfile ]; then \
		echo "Building Docker image..."; \
		docker build -t $(BINARY_NAME) .; \
	else \
		echo "⚠️  Dockerfile not found"; \
	fi

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "✅ All checks passed"

## help: Show this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

# Version information
version:
	@echo "Go version: $(shell $(GOCMD) version)"
	@echo "Module: $(shell $(GOMOD) list -m)"