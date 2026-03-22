# Moolah Makefile
# Consolidates CI steps for local execution

.PHONY: all help deps lint security test build check-ci format clean sqlc


all: help

help:
	@echo "Moolah Build System"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  deps        Install development tools (golangci-lint, sqlc, gofumpt)"
	@echo "  lint        Run linting checks via golangci-lint"
	@echo "  security    Run security scans (gosec)"
	@echo "  test        Run unit tests with race detection and coverage"
	@echo "  build       Build the API binary"
	@echo "  sqlc        Generate SQL code using sqlc"
	@echo "  check-ci    Run all CI steps (format, lint, security, test, build)"
	@echo "  clean       Remove build artifacts"

deps:
	@echo "Installing dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install mvdan.cc/gofumpt@latest

lint:
	@echo "Running linting checks..."
	golangci-lint run

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

check-ci: format lint security test build
	@echo "CI check path passed successfully!"

format:
	@echo "Formatting code..."
	go fmt ./...
	go fix ./...
	gofumpt -l -w .

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
