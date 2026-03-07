.PHONY: all build run test lint generate migrate-up migrate-down clean help

# Configuration
BINARY_NAME=moolah-api
CMD_DIR=./cmd/api
OUT_DIR=bin

# Default Go toolchain command
GO=go

all: lint generate test build

## build: Build the API binary
build:
	@echo "Building binary..."
	@mkdir -p $(OUT_DIR)
	$(GO) build -o $(OUT_DIR)/$(BINARY_NAME) $(CMD_DIR)

## run: Run the API application
run:
	$(GO) run $(CMD_DIR)

## test: Run unit tests
test:
	@echo "Running unit tests..."
	$(GO) test -v ./...

## test-integration: Run integration tests (docker required)
test-integration:
	@echo "Running integration tests..."
	$(GO) test -v -tags=integration ./...

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...

## generate: Run sqlc generate
generate:
	@echo "Generating SQL code..."
	sqlc generate

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf $(OUT_DIR)
	rm -rf vendor/

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@$(GO) run -e ' \
		f, _ := os.Open("Makefile"); \
		scanner := bufio.NewScanner(f); \
		for scanner.Scan() { \
			line := scanner.Text(); \
			if strings.HasPrefix(line, "## ") { \
				parts := strings.Split(strings.TrimPrefix(line, "## "), ":"); \
				fmt.Printf("  %-15s %s\n", parts[0], parts[1]); \
			} \
		}' # This is a conceptual help command, usually implemented via grep/awk in shell
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
