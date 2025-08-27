# KryptX VPN Makefile

BINARY_NAME=kryptx
BUILD_DIR=build
SRC_DIR=cmd/client

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty)"
BUILD_FLAGS=-v $(LDFLAGS)

.PHONY: all build clean test deps install run

all: clean deps build

build:
	@echo "Building KryptX VPN..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)

build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(SRC_DIR)
	
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(SRC_DIR)
	
	# macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(SRC_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(SRC_DIR)

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

install: build
	@echo "Installing KryptX VPN..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

run: build
	@echo "Running KryptX VPN..."
	./$(BUILD_DIR)/$(BINARY_NAME)

dev:
	@echo "Running in development mode..."
	$(GOCMD) run $(SRC_DIR) -v

package: build-all
	@echo "Creating packages..."
	@mkdir -p $(BUILD_DIR)/packages
	
	# Linux package
	tar -czf $(BUILD_DIR)/packages/$(BINARY_NAME)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-amd64
	
	# Windows package
	zip -j $(BUILD_DIR)/packages/$(BINARY_NAME)-windows-amd64.zip $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	
	# macOS packages
	tar -czf $(BUILD_DIR)/packages/$(BINARY_NAME)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-amd64
	tar -czf $(BUILD_DIR)/packages/$(BINARY_NAME)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-arm64

setup-dev:
	@echo "Setting up development environment..."
	$(GOGET) fyne.io/fyne/v2/cmd/fyne
	$(GOGET) -u all

help:
	@echo "Available commands:"
	@echo "  build      - Build the application"
	@echo "  build-all  - Build for all platforms"
	@echo "  clean      - Clean build files"
	@echo "  test       - Run tests"
	@echo "  deps       - Install dependencies"
	@echo "  install    - Install to system"
	@echo "  run        - Build and run"
	@echo "  dev        - Run in development mode"
	@echo "  package    - Create distribution packages"
	@echo "  setup-dev  - Setup development environment"