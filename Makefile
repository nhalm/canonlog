.PHONY: help test lint bench coverage fmt check

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: ## Run tests with race detector
	go test -v -race ./...

lint: ## Run linter
	golangci-lint run

bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

coverage: ## Generate coverage report
	go test -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html

fmt: ## Format code
	gofmt -w .

check: fmt lint test ## Run fmt, lint, and test
