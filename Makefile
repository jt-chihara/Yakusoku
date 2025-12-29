# Yakusoku Makefile

.PHONY: all build build-cli build-broker build-ui test lint clean install help

# Default target
all: lint test build

# Build targets
build: build-cli build-broker

build-cli:
	@echo "Building yakusoku CLI..."
	go build -o bin/yakusoku ./cmd/yakusoku

build-ui:
	@echo "Building Web UI..."
	cd web && pnpm install && pnpm build
	rm -rf internal/broker/ui/dist
	cp -r web/dist internal/broker/ui/

build-broker: build-ui
	@echo "Building yakusoku-broker..."
	go build -o bin/yakusoku-broker ./cmd/yakusoku-broker

# Test targets
test:
	@echo "Running tests..."
	go test -v -race ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-unit:
	@echo "Running unit tests..."
	go test -v -race ./tests/unit/...

test-integration:
	@echo "Running integration tests..."
	go test -v -race ./tests/integration/...

# Lint target
lint:
	@echo "Running linter..."
	golangci-lint run

lint-fix:
	@echo "Running linter with auto-fix..."
	golangci-lint run --fix

# Format target
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w -local github.com/jt-chihara/yakusoku .

# Clean target
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf web/dist web/node_modules

# Install targets
install: build
	@echo "Installing binaries..."
	cp bin/yakusoku $(GOPATH)/bin/
	cp bin/yakusoku-broker $(GOPATH)/bin/

# Development helpers
dev-deps:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Help target
help:
	@echo "Yakusoku Makefile targets:"
	@echo "  all            - Run lint, test, and build (default)"
	@echo "  build          - Build CLI and Broker binaries"
	@echo "  build-cli      - Build yakusoku CLI"
	@echo "  build-ui       - Build Web UI (requires pnpm)"
	@echo "  build-broker   - Build yakusoku-broker server (includes UI)"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  lint           - Run golangci-lint"
	@echo "  lint-fix       - Run golangci-lint with auto-fix"
	@echo "  fmt            - Format code with gofmt and goimports"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install binaries to GOPATH"
	@echo "  dev-deps       - Install development dependencies"
	@echo "  help           - Show this help message"
