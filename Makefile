BINARY_NAME=manager
MAIN_PACKAGE=./cmd/main.go
BUILD_DIR=bin
GO_TAGS=sqlite_fts5
SQLC_DIR=internal/db

all: $(manager)
	
.PHONY: help build-windows build-linux run clean setup clean generate fmt install-tools 
	
.DEFAULT-GOAL := help

help: 
	@echo "HELP!!!!"

setup: install-tools generate  ## Install tools and generate code

install-tools:  ## Install development tools
	@echo "Installing tools..."
	go install github.com/air-verse/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

	
generate: ## Generate templ and sqlc code
	@echo "Generating code..."
	@cd $(SQLC_DIR) && sqlc generate
	@templ generate
	@echo "Code generation complete!"
	
build: generate fmt vet ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LD_FLAGS) -tags "$(GO_TAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	
build-windows: generate  ## Build for Windows
	GOOS=windows GOARCH=amd64 go build $(LD_FLAGS) -tags "$(GO_TAGS)" -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PACKAGE)
	
run: build  ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

fmt:  ## Format Go code
	go fmt ./...
	go mod tidy

vet:  ## Run go vet
	go vet -tags "$(GO_TAGS)" ./...
	
clean:
	go clean -cache
	go clean -modcache
	go clean -testcache