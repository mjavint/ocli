.PHONY: help install build build-all build-linux build-darwin build-windows clean test

# Variables
BINARY_NAME=ocli
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=bin
GO=go
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -s -w"

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Show this help message
	@echo "$(GREEN)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

install: ## Install dependencies
	@echo "$(GREEN)Installing dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(GREEN)Dependencies installed successfully$(NC)"

build: ## Build for current platform
	@echo "$(GREEN)Building $(BINARY_NAME) for current platform...$(NC)"
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "$(GREEN)Build complete: $(BINARY_NAME)$(NC)"

build-all: build-linux build-darwin build-windows ## Build for all platforms (64-bit)
	@echo "$(GREEN)All builds completed successfully$(NC)"

build-linux: ## Build for Linux (amd64 and arm64)
	@echo "$(GREEN)Building for Linux amd64...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "$(GREEN)Building for Linux arm64...$(NC)"
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo "$(GREEN)Linux builds complete$(NC)"

build-darwin: ## Build for macOS (amd64 and arm64)
	@echo "$(GREEN)Building for macOS amd64...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@echo "$(GREEN)Building for macOS arm64...$(NC)"
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo "$(GREEN)macOS builds complete$(NC)"

build-windows: ## Build for Windows (amd64)
	@echo "$(GREEN)Building for Windows amd64...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "$(GREEN)Windows build complete: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe$(NC)"

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	$(GO) test -v ./...

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)Clean complete$(NC)"

run: build ## Build and run the application
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	./$(BINARY_NAME)

.DEFAULT_GOAL := help
