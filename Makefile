.PHONY: all build run test lint generate clean help task-check deps up down swagger

# Configuration
BINARY_NAME=moolah-api
CMD_DIR=./cmd/api
OUT_DIR=bin
SWAGGER_OUT=api

# Default Go toolchain command
GO=go

all: deps lint generate test build swagger

## task-check: Run all checks required before completing a task (Linter, SQLC, Security, Unit Tests with Coverage)
task-check: deps lint-check sqlc-check security-check test-coverage swagger-check

deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod vendor
	@gofumpt -w .
	@go fmt .
	@go fix ./...

## swagger: Generate Swagger documentation
swagger:
	@echo "📝 Generating Swagger documentation..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@mkdir -p $(SWAGGER_OUT)
	@$$(go env GOPATH)/bin/swag init --dir ./cmd/api,./internal/handler,./internal/domain --output $(SWAGGER_OUT)

## swagger-check: Verify if swagger documentation is up to date
swagger-check: swagger
	@echo "⚙️ Checking swagger documentation..."
	@if [ -n "$$(git diff --name-only $(SWAGGER_OUT)/)" ]; then \
		echo "❌ Error: Swagger documentation is out of date. Commit the changes."; \
		exit 1; \
	fi; \
	echo "✅ Swagger documentation is up to date."

## lint-check: Run golangci-lint
lint-check:
	@echo "🔍 Running linter..."
	golangci-lint run --build-tags=integration ./...

## sqlc-check: Verify if sqlc generate is up to date
sqlc-check:
	@echo "⚙️ Checking sqlc generation..."
	@if [ -n "$$(ls internal/platform/db/queries/*.sql 2>/dev/null)" ]; then \
		sqlc generate; \
		if [ -n "$$(git diff --name-only internal/platform/db/sqlc/)" ]; then \
			echo "❌ Error: sqlc generated code is out of date. Commit the changes."; \
			exit 1; \
		fi; \
		echo "✅ sqlc is up to date."; \
	else \
		echo "⏭️ No SQL queries found in internal/platform/db/queries/. Skipping sqlc generate."; \
	fi

## security-check: Run security scans (govulncheck and gosec)
security-check:
	@echo "🛡️ Running security scans..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@gosec ./...

## test-coverage: Run unit tests and enforce coverage (80% threshold)
test-coverage:
	@echo "🧪 Running unit tests with coverage..."
	@$(GO) test -v -race -count=1 -tags=integration -timeout=600s -coverprofile=coverage.out -covermode=atomic $$(go list ./... | grep -v /platform/db/sqlc | grep -v /testutil/mocks | grep -v /api)
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | tr -d '%'); \
	echo "Total coverage: $${COVERAGE}%"; \
	awk "BEGIN { if ($${COVERAGE} < 80) exit 1 }"; \
	if [ $$? -ne 0 ]; then \
		echo "❌ ERROR: Coverage $${COVERAGE}% is below the 80% threshold"; \
		exit 1; \
	fi
	@echo "✅ Tests passed with sufficient coverage."

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
	$(GO) test -v -tags=integration ./...

## test-integration: Run integration tests (docker required)
test-integration:
	@echo "Running integration tests..."
	$(GO) test -v -tags=integration ./...

## smoke-test: Run Phase 1 end-to-end smoke test (docker required)
smoke-test:
	@echo "Running Phase 1 smoke test..."
	$(GO) test -v -race -count=1 -tags=integration -timeout=300s \
		-run TestSmoke_Phase1HappyPath \
		./internal/server/...

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	golangci-lint run --build-tags=integration ./...

## generate: Run sqlc generate
generate:
	@echo "Generating SQL code..."
	sqlc generate

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf $(OUT_DIR)
	rm -rf vendor/

## clean-branches: Remove local branches that no longer exist on origin
clean-branches:
	@echo "🧹 Cleaning up local branches that are gone from origin..."
	@git fetch -p
	@git branch -vv | grep ': gone]' | awk '{print $$1}' | xargs -r git branch -D

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

## up: Start the Docker Compose
up:
	@echo "Starting Docker Compose..."
	@docker compose up -d

## down: Stop the Docker Compose
down:
	@echo "Stopping Docker Compose..."
	@docker compose down
