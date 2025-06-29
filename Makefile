.PHONY: build clean test fmt vet lint help

# Variables
BINARY_NAME=moley
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Build the application
build: fmt vet
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o ${BINARY_NAME} .

# Install the application
install: build
	@echo "Installing ${BINARY_NAME}..."
	go install ${LDFLAGS} .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	go clean
	rm -f ${BINARY_NAME}

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Create a new release
release: clean build test
	@echo "Creating release for version ${VERSION}..."
	@echo "Binary: ${BINARY_NAME}"