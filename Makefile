.PHONY: build install test test-coverage test-unit test-acc test-race bench test-run test-report fmt vet clean clean-all clean-test check check-full deps dev lint release-test setup-dev

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build the provider (with linting)
build: fmt vet lint
	go build -o terraform-provider-pihole

# Install the provider locally
install: build
	@ARCH=$$(go env GOARCH); \
	OS=$$(go env GOOS); \
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/lukaspustina/pihole/$(VERSION)/$${OS}_$${ARCH}/; \
	cp terraform-provider-pihole ~/.terraform.d/plugins/registry.terraform.io/lukaspustina/pihole/$(VERSION)/$${OS}_$${ARCH}/

# Run all tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run unit tests only (exclude acceptance tests)  
test-unit:
	go test -v -short ./...

# Run acceptance tests (requires real Pi-hole server)
test-acc:
	TF_ACC=1 go test -v ./internal/provider -timeout 30m

# Run tests with race detection
test-race:
	go test -v -race ./...

# Run benchmarks
bench:
	go test -v -bench=. -benchmem ./...

# Run specific test
test-run:
	@echo "Usage: make test-run TEST=TestName"
	go test -v -run $(TEST) ./...

# Generate test report
test-report: test-coverage
	@echo "Test coverage report generated: coverage.html"
	@echo "Open coverage.html in your browser to view detailed coverage"

# Format Go code with gofmt and goimports
fmt:
	@echo "Running gofmt..."
	go fmt ./...
	@echo "Running goimports..."
	goimports -w .

# Run go vet
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -f terraform-provider-pihole

# Clean all artifacts
clean-all: clean clean-test

# Download dependencies
deps:
	go mod tidy

# Run all quality checks (includes linting)
check: fmt vet lint test-unit

# Run comprehensive checks including coverage
check-full: fmt vet lint test-coverage

# Clean test artifacts
clean-test:
	rm -f coverage.out coverage.html

# Development build with debugging
dev: build
	./terraform-provider-pihole -debug

# Run golangci-lint with config
lint:
	@echo "Running golangci-lint..."
	golangci-lint run --config .golangci.yml

# Test GoReleaser configuration without releasing
release-test:
	goreleaser release --snapshot --skip-publish --clean

# Setup development environment with git hooks
setup-dev:
	@echo "Setting up development environment..."
	@git config core.hooksPath .githooks
	@echo "✅ Git hooks configured to use .githooks/"
	@echo "💡 Now all commits will automatically run formatting and linting!"