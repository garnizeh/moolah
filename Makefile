# Tool versions
GOFUMPT_VERSION=v0.7.0
FIELDALIGN_VERSION=v0.30.0
GOSEC_VERSION=v2.22.0
SQLC_VERSION=v1.28.0

all: help


help:
	@echo "Moolah Build System"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  deps        Install CI tools (gofumpt, fieldalignment, gosec, sqlc)"
	@echo "  lint        Run linting checks (gofumpt, fieldalignment)"
	@echo "  security    Run security scans (gosec)"
	@echo "  test        Run unit tests with race detection and coverage"
	@echo "  build       Build the API binary"
	@echo "  sqlc        Generate SQL code using sqlc"
	@echo "  check-ci    Run all CI steps in order (format, lint, security, test, build)"
	@echo "  clean       Remove build artifacts"

deps:
	@echo "Installing dependencies..."
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@$(FIELDALIGN_VERSION)
	go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)


lint:
	@echo "Running linting checks..."
	gofumpt -d . | tee /dev/stderr | (! grep -q .)
	fieldalignment ./...

security:
	@echo "Running security scans..."
	gosec ./...

test:
	@echo "Running unit tests..."
	go test -v -race -cover ./...

build:
	@echo "Building API..."
	mkdir -p bin
	go build -v -o bin/api ./cmd/api

sqlc:
	@echo "Generating SQL code..."
	sqlc generate

check-ci: deps format lint security test build

	@echo "CI check path passed successfully!"

format:
	@echo "Formatting code..."
	go fmt ./...
	go fix ./...
	gofumpt -l -w .

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
