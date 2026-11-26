# The verb contract of terraform-provider-pihole.

version := `git describe --tags --always --dirty 2>/dev/null || echo dev`

default: verify

# --- the contract ------------------------------------------------------------

# The gate: format check, vet, lint, unit tests. Writes nothing.
verify: fmt-check vet lint test-unit

# Everything: verify plus coverage over the full test set.
check: fmt-check vet lint test-coverage

# All tests. Acceptance tests skip themselves unless TF_ACC is set.
test:
    go test -v ./...

# Unit tests only — `-short` excludes the acceptance tests.
test-unit:
    go test -v -short ./...

# gofmt + goimports, reporting only. The writing form is `just fmt`.
fmt-check:
    #!/usr/bin/env bash
    set -euo pipefail
    offenders="$(gofmt -l . ; goimports -l .)"
    if [ -n "$offenders" ]; then
        echo "not formatted:" >&2
        printf '%s\n' "$offenders" | sort -u >&2
        echo "run 'just fmt'" >&2
        exit 1
    fi
    echo "ok gofmt and goimports are clean"

vet:
    go vet ./...

# golangci-lint with the repository's config.
lint:
    golangci-lint run --config .golangci.yml

# --- tests -------------------------------------------------------------------

# Acceptance tests. Needs a real Pi-hole server; reachable from no gate.
test-acc:
    TF_ACC=1 go test -v ./internal/provider -timeout 30m

test-race:
    go test -v -race ./...

# Coverage profile plus coverage.html.
test-coverage:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# usage: just test-run TestAccCNAMERecord
test-run test:
    go test -v -run {{test}} ./...

bench:
    go test -v -bench=. -benchmem ./...

# --- build -------------------------------------------------------------------

# Build the provider binary (formats and lints first).
build: fmt vet lint
    go build -o terraform-provider-pihole

# Install into ~/.terraform.d/plugins for local terraform runs.
install: build
    #!/usr/bin/env bash
    set -euo pipefail
    arch="$(go env GOARCH)"
    os="$(go env GOOS)"
    dest="$HOME/.terraform.d/plugins/registry.terraform.io/lukaspustina/pihole/{{version}}/${os}_${arch}"
    mkdir -p "$dest"
    cp terraform-provider-pihole "$dest/"

# Build and run with the debug flag.
dev: build
    ./terraform-provider-pihole -debug

# go mod tidy.
deps:
    go mod tidy

# --- fixing and housekeeping -------------------------------------------------

# gofmt + goimports, writing. No gate reaches this.
fmt:
    go fmt ./...
    goimports -w .

clean:
    rm -f terraform-provider-pihole

clean-test:
    rm -f coverage.out coverage.html

# GoReleaser dry run.
release-test:
    goreleaser release --snapshot --skip-publish --clean

# Point git at .githooks (formatting and linting on commit).
#
setup-dev:
    git config core.hooksPath .githooks
    @echo "Git hooks configured to use .githooks/"

# The Go tools verify calls, pinned, into Go's bin directory (on PATH):
# goimports for fmt-check, golangci-lint at the version CI uses (built with a Go that reads the local toolchain's export data).

# Install the pinned Go tools verify needs.
setup:
    go install golang.org/x/tools/cmd/goimports@v0.51.0
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
