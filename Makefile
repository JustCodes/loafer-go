# Project settings
MODULE        := $(shell go list -m)
PKGS          := $(shell go list ./... | grep -Ev 'example|fake')
COVER_TMP     := cover.out.tmp
COVER_FINAL   := cover.out
COVER_TOTAL   := geral.out

# Tool versions
GOLANGCI_LINT_VERSION := v1
LEFTHOOK_VERSION      := v1.4.8

.PHONY: all clean format lint test cover configure \
        install-goimports install-golang-ci install-lefthook install-fieldalignment \
        update-dependencies check

# Default task
all: lint test

# Format code
format:
	@echo "Formatting code..."
	@goimports -local $(MODULE) -w -l .
	@fieldalignment -fix ./...

# Run linters
lint: format
	@echo "Running linters..."
	@golangci-lint run --allow-parallel-runners --max-same-issues 0 ./...

# Run tests with race condition check and coverage
test: clean
	@echo "Running tests..."
	@go test -timeout 1m -race -covermode=atomic -coverprofile=$(COVER_TOTAL) $(PKGS)
	@go tool cover -func=$(COVER_TOTAL)

# Generate filtered coverage profile (excluding fakes)
cover:
	@echo "Generating filtered test coverage..."
	@go test -covermode=count -coverprofile=$(COVER_TMP) ./...
	@cat $(COVER_TMP) | grep -v '/example' > $(COVER_FINAL)
	@go tool cover -func=$(COVER_FINAL)

# Remove test cache
clean:
	@echo "Cleaning test cache..."
	@go clean -testcache

# Install all dev tools
configure: install-goimports install-fieldalignment install-golang-ci install-lefthook
	@echo "Installing lefthook hooks..."
	@lefthook install

# Install tools individually
install-goimports:
	@echo "Installing goimports..."
	@go install golang.org/x/tools/cmd/goimports@latest

install-fieldalignment:
	@echo "Installing fieldalignment..."
	@go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest

install-golang-ci:
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

install-lefthook:
	@echo "Installing lefthook..."
	@go install github.com/evilmartians/lefthook@$(LEFTHOOK_VERSION)

# Update module dependencies
update-dependencies:
	@echo "Updating dependencies..."
	@go get -t -u ./...
	@go mod tidy

# Quick health check
check: lint test

test-bench:
	@$(MAKE) -s clean
	@go test -bench=. ./... -benchtime=5s -count 1 -benchmem
