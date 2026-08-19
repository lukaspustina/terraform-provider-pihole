# The verb contract of terraform-provider-pihole (pdt-adlc ADR 0008).
#
# Migrated from a Makefile on 2026-08-19, by reduction: four of 24 targets are
# gone — help had none to drop here, so it is check-full (its contents are what
# `check` and `test-coverage` already say), test-report (test-coverage plus two
# echo lines), clean-all (the union of two recipes one calls individually), and
# the old `check`, whose contents move to `adlc-verify`. `test-run TEST=X` became
# `just test-run <name>`.
#
# UNVERIFIED, AND SAID SO OUT LOUD. Every other repository in this migration had
# its gate run green before the commit. This one could not: on the machine where
# it was migrated, `go` reached no modules at all — `proxy.golang.org` answered
# curl in 160 ms and timed out for the go client against the same IP, and
# GOPROXY=direct failed on golang.org itself. So `go vet`, `golangci-lint` and
# `go test` all end in "setup failed" for a missing terraform-plugin-framework
# v1.17.0, and none of the recipes below has been executed. Run `just
# adlc-verify` once on a machine whose go can fetch, and fix what this header
# got wrong.
#
# What the migration DID change on purpose: the resolved gate used to be
# `check: fmt vet lint test-unit`, and `fmt` is `go fmt ./...` plus
# `goimports -w .` — it WRITES. A gate that rewrites the working tree cannot be
# attested, because the attestation binds a tree hash (repo-contract
# requirement 3; the same defect keeps pdt-wealth unprovable). adlc-verify now
# runs fmt-check, which lists offenders and changes nothing, and `fmt` stays a
# verb of its own that no gate reaches.

version := `git describe --tags --always --dirty 2>/dev/null || echo dev`

default: adlc-verify

# --- the contract ------------------------------------------------------------

# What the ADLC gate runs: format check, vet, lint, unit tests. Writes nothing.
adlc-verify: fmt-check vet lint test-unit

# Everything: the gate plus coverage over the full test set.
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
# NOTE: this REPLACES the global hook path, which is where the ADLC gate lives —
# so a repository that has run this commits with .githooks and no gate at all.
# That is the state this repository is in (pdt-adlc backlog A1), and this target
# is how it got there.
#
# Point git at .githooks — this turns the ADLC gate off, see above.
setup-dev:
    git config core.hooksPath .githooks
    @echo "Git hooks configured to use .githooks/"
