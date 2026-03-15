.PHONY: all build run test test-ui lint generate clean help task-check deps up down swagger templ tailwind web-build web

# Configuration
BINARY_NAME=moolah-api
WEB_BINARY_NAME=moolah-web
CMD_DIR=./cmd/api
WEB_CMD_DIR=./cmd/web
OUT_DIR=bin
SWAGGER_OUT=api
GO=go
TAILWIND_VERSION=4.1.8

BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH := $(shell git rev-parse HEAD)
GO_VERSION := $(shell go version | cut -c 14-)
VERSION_TAG := $(shell git describe --tags --match "v*" --abbrev=0 2>/dev/null || echo "v0.0.0-dev")

all: deps format lint generate test build swagger

## task-check: Run all checks required before completing a task (Linter, SQLC, Security, Unit Tests with Coverage)
task-check: deps format lint-check sqlc-check templ-check swagger-check security-check test-coverage

# deps: Install Go dependencies
deps: install-tailwind
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod vendor

## format: Run code formatters and fixers
format:
	@echo "Formatting code..."
	@go install mvdan.cc/gofumpt@latest
	@go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	@gofumpt -w .
	@go fmt .
	@go fix ./...
	@fieldalignment -fix ./...

## install-tailwind: Install Standalone Tailwind CSS CLI with checksum verification
install-tailwind:
	@echo "Standalone Tailwind CSS CLI..."
	@if [ ! -f /usr/local/bin/tailwindcss ]; then \
		curl -fsSL "https://github.com/tailwindlabs/tailwindcss/releases/download/v$(TAILWIND_VERSION)/tailwindcss-linux-x64" -o tailwindcss-linux-x64; \
		curl -fsSL "https://github.com/tailwindlabs/tailwindcss/releases/download/v$(TAILWIND_VERSION)/sha256sums.txt" -o sha256sums.txt; \
		grep "tailwindcss-linux-x64" sha256sums.txt | sha256sum -c -; \
		sudo install -m 0755 tailwindcss-linux-x64 /usr/local/bin/tailwindcss; \
		rm tailwindcss-linux-x64 sha256sums.txt; \
	fi

## build-api: Build the API binary
build-api:
	@echo "🏗️ Building API binary..."
	@mkdir -p $(OUT_DIR)
	$(GO) build -mod=vendor -ldflags="-s -w -X 'main.tagVersion=$(VERSION_TAG)' -X 'main.buildTime=$(BUILD_TIME)' -X 'main.commitHash=$(COMMIT_HASH)' -X 'main.goVersion=$(GO_VERSION)'" -o $(OUT_DIR)/$(BINARY_NAME) $(CMD_DIR)

## build-web: Build the web binary (templ + tailwind + go build)
build-web: templ tailwind
	@echo "🏗️ Building web binary..."
	@mkdir -p $(OUT_DIR)
	$(GO) build -mod=vendor -ldflags="-s -w -X 'main.tagVersion=$(VERSION_TAG)' -X 'main.buildTime=$(BUILD_TIME)' -X 'main.commitHash=$(COMMIT_HASH)' -X 'main.goVersion=$(GO_VERSION)'" -o $(OUT_DIR)/$(WEB_BINARY_NAME) $(WEB_CMD_DIR)

## build: Build API and Web binaries
build: build-api build-web

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
	golangci-lint run --build-tags=integration

## sqlc-check: Verify if sqlc generate is up to date
sqlc-check:
	@echo "⚙️ Checking sqlc generation..."
	@if [ -n "$$(ls internal/platform/db/queries/*.sql 2>/dev/null)" ]; then \
		go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; \
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
test-coverage: test-all test-integration
	@echo "📊 Aggregating coverage reports..."
	@COVERAGE=$$(go tool cover -func=unit.out -func=ui.out -func=integration.out | grep total | awk '{print $$3}' | tr -d '%'); \
	echo "Total combined coverage: $${COVERAGE}%"; \
	awk "BEGIN { if ($${COVERAGE} < 80) exit 1 }"; \
	if [ $$? -ne 0 ]; then \
		echo "❌ ERROR: Combined Coverage $${COVERAGE}% is below the 80% threshold"; \
		exit 1; \
	fi
	@echo "✅ Tests passed with sufficient coverage."

## smoke-test: Run Phase 1 end-to-end smoke test (docker required)
smoke-test:
	@echo "Running Phase 1 smoke test..."
	$(GO) test -v -race -count=1 -tags=integration -timeout=300s \
		-run TestSmoke_Phase1HappyPath \
		./internal/server/...

## templ: Run templ code generation
templ:
	@echo "==> Running templ generate..."
	@go install github.com/a-h/templ/cmd/templ@latest
	templ generate ./...

## templ-check: Verify if templ generation is up to date
templ-check: templ
	@echo "⚙️ Checking templ generation..."
	@if [ -n "$$(git diff --name-only internal/ui/)" ]; then \
		echo "❌ Error: Templ generated code is out of date. Run 'make templ' and commit the changes."; \
		exit 1; \
	fi; \
	echo "✅ Templ is up to date."

## tailwind: Build optimised Tailwind CSS bundle
tailwind:
	@echo "==> Building Tailwind CSS..."
	tailwindcss -i web/static/css/app.css -o web/static/css/app.min.css --minify

## run-web: Run the web UI server locally (development)
run-web:
	$(GO) run $(WEB_CMD_DIR)

## run-api: Run the API application
run-api:
	$(GO) run $(CMD_DIR)

## test: Run API and business logic unit tests (excludes UI)
test:
	@echo "🧪 Running API unit tests..."
	@$(GO) test -v -race -tags=integration -count=1 -timeout=300s -coverprofile=unit.out -covermode=atomic $$(go list ./... | grep -v /internal/ui | grep -v /cmd/api | grep -v /internal/platform/db/sqlc | grep -v /testutil/mocks | grep -v /api)

## test-ui: Run UI/Templ component tests
test-ui: templ
	@echo "🎨 Running UI component tests..."
	@$(GO) test -v -race -count=1 -timeout=300s -coverprofile=ui.out -covermode=atomic ./internal/ui/...

## test-all: Run API and UI tests in parallel
test-all:
	@echo "🚀 Running all tests in parallel..."
	@$(MAKE) -j2 test test-ui

## test-integration: Run integration tests (excludes UI)
test-integration:
	@echo "🧪 Running API integration tests..."
	@$(GO) test -v -race -count=1 -timeout=120s \
		-tags=integration \
		-coverprofile=integration.out \
		-covermode=atomic \
		./internal/platform/repository/...

## test-coverage: Run unit tests and enforce coverage (80% threshold)
test-coverage: test-all test-integration
	@echo "📊 Aggregating coverage reports..."
	@COVERAGE=$$(go tool cover -func=unit.out -func=ui.out -func=integration.out | grep total | awk '{print $$3}' | tr -d '%'); \
	echo "Total combined coverage: $${COVERAGE}%"; \
	awk "BEGIN { if ($${COVERAGE} < 80) exit 1 }"; \
	if [ $$? -ne 0 ]; then \
		echo "❌ ERROR: Combined Coverage $${COVERAGE}% is below the 80% threshold"; \
		exit 1; \
	fi
	@echo "✅ Tests passed with sufficient coverage."

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	golangci-lint run --build-tags=integration

## generate: Run sqlc and templ code generation
generate: templ
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
