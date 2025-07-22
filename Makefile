.PHONY: build install test fmt vet clean

# Build the provider
build:
	go build -o terraform-provider-pihole

# Install the provider locally
install: build
	@ARCH=$$(go env GOARCH); \
	OS=$$(go env GOOS); \
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/lukaspustina/pihole/1.0.0/$${OS}_$${ARCH}/; \
	cp terraform-provider-pihole ~/.terraform.d/plugins/registry.terraform.io/lukaspustina/pihole/1.0.0/$${OS}_$${ARCH}/

# Run tests
test:
	go test -v ./...

# Format Go code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -f terraform-provider-pihole

# Download dependencies
deps:
	go mod tidy

# Run all quality checks
check: fmt vet test

# Development build with debugging
dev: build
	./terraform-provider-pihole -debug