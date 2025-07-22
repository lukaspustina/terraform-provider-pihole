.PHONY: build install test test-coverage test-unit test-acc test-race bench test-run test-report fmt vet clean clean-all clean-test check check-full deps dev lint release-test

# Build the provider
build:
	go build -o terraform-provider-pihole

# Install the provider locally
install: build
	@ARCH=$$(go env GOARCH); \
	OS=$$(go env GOOS); \
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/lukaspustina/pihole/0.1.0/$${OS}_$${ARCH}/; \
	cp terraform-provider-pihole ~/.terraform.d/plugins/registry.terraform.io/lukaspustina/pihole/0.1.0/$${OS}_$${ARCH}/

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

# Format Go code
fmt:
	go fmt ./...

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

# Run all quality checks
check: fmt vet test-unit

# Run comprehensive checks including coverage
check-full: fmt vet test-coverage

# Clean test artifacts
clean-test:
	rm -f coverage.out coverage.html

# Development build with debugging
dev: build
	./terraform-provider-pihole -debug

# Run golangci-lint
lint:
	golangci-lint run

# Test GoReleaser configuration without releasing
release-test:
	goreleaser release --snapshot --skip-publish --clean