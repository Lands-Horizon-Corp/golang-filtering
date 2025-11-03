.PHONY: help build run test test-verbose test-coverage clean lint fmt vet install-deps update-deps

# Default target
help:
	@echo "Available targets:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make test          - Run all tests"
	@echo "  make test-verbose  - Run tests with verbose output"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make clean         - Clean build artifacts and test files"
	@echo "  make lint          - Run golangci-lint"
	@echo "  make fmt           - Format code with gofmt"
	@echo "  make vet           - Run go vet"
	@echo "  make install-deps  - Install dependencies"
	@echo "  make update-deps   - Update all dependencies"

# Build the application
build:
	@echo "Building..."
	@go build -o bin/app .

# Run the application
run:
	@echo "Running..."
	@go run .

# Run all tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@go test ./... -v -count=1

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./test -coverpkg=./filter -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo ""
	@echo "Coverage Summary:"
	@go tool cover -func=coverage.out | grep total

# Clean build artifacts and test files
clean:
	@echo "Cleaning..."
	@rm -f bin/app
	@rm -f test.db
	@rm -f coverage.out coverage.html
	@rm -rf bin/

# Run golangci-lint
lint:
	@echo "Running linter..."
	@golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Update all dependencies
update-deps:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Run all quality checks
check: fmt vet lint test
	@echo "All checks passed!"
