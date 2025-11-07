.PHONY: build test lint fmt clean install

# Build the binary
build:
	go build -o bin/nvolt ./cmd/nvolt

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run linter
lint:
	golangci-lint run ./...

# Format code
fmt:
	gofmt -s -w .
	goimports -w .

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Install dependencies
deps:
	go mod download
	go mod tidy

# Install the binary locally
install:
	go install ./cmd/nvolt

# Run all checks (format, lint, test)
check: fmt lint test

# Development build with race detector
dev:
	go build -race -o bin/nvolt-dev ./cmd/nvolt
