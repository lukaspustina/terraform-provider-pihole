# Contributing to Pi-hole Terraform Provider

Thank you for your interest in contributing to the Pi-hole Terraform Provider! This document provides guidelines and information for contributors.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Release Process](#release-process)

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/terraform-provider-pihole.git`
3. Add upstream remote: `git remote add upstream https://github.com/lukaspustina/terraform-provider-pihole.git`

## Development Environment

### Prerequisites

- [Go](https://golang.org/doc/install) 1.22 or later
- [Terraform](https://developer.hashicorp.com/terraform/downloads) 1.5 or later
- [Pi-hole](https://pi-hole.net/) instance for testing
- [golangci-lint](https://golangci-lint.run/usage/install/) for code linting
- [GoReleaser](https://goreleaser.com/install/) for release testing (optional)

### Setup

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Build the provider:
   ```bash
   make build
   ```

3. Install the provider locally:
   ```bash
   make install
   ```

## Making Changes

### Branch Naming

Use descriptive branch names:
- `feature/add-blocklist-resource`
- `bugfix/fix-dns-record-validation`
- `docs/update-readme`

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat: add support for DNS record validation`
- `fix: resolve authentication timeout issue`
- `docs: update installation instructions`
- `test: add unit tests for client authentication`

## Testing

### Running Tests

```bash
# Run unit tests
make test-unit

# Run tests with coverage
make test-coverage

# Run all tests (including acceptance tests with TF_ACC=1)
make test

# Run acceptance tests (requires Pi-hole server)
TF_ACC=1 PIHOLE_URL=http://your-pihole:80 PIHOLE_PASSWORD=your-password make test-acc
```

### Test Requirements

- **Unit tests** are required for all new functionality
- **Acceptance tests** should be added for new resources and data sources
- Maintain or improve test coverage
- Tests should be deterministic and not rely on external services (use mocks)

### Test Environment

For acceptance tests, you can use Docker to run Pi-hole:

```bash
docker run -d \
  --name pihole-test \
  -p 8080:80 \
  -e WEBPASSWORD=test-password \
  pihole/pihole:latest

# Run acceptance tests
TF_ACC=1 PIHOLE_URL=http://localhost:8080 PIHOLE_PASSWORD=test-password make test-acc

# Cleanup
docker stop pihole-test && docker rm pihole-test
```

## Submitting Changes

### Pull Request Process

1. **Update your branch**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run quality checks**:
   ```bash
   make check-full  # Runs fmt, vet, and tests with coverage
   make lint        # Run golangci-lint
   ```

3. **Create a pull request** with:
   - Clear description of changes
   - Link to related issues
   - Screenshots if applicable
   - Updated documentation

### Pull Request Requirements

- [ ] All tests pass
- [ ] Code coverage maintained or improved
- [ ] Code is formatted (`go fmt`)
- [ ] No linting errors (`golangci-lint run`)
- [ ] Documentation updated if needed
- [ ] CHANGELOG.md updated (if applicable)

## Code Style

### Go Conventions

- Follow standard Go conventions and idioms
- Use meaningful variable and function names
- Add comments for public APIs
- Keep functions focused and small
- Use error wrapping: `fmt.Errorf("operation failed: %w", err)`

### Terraform Provider Conventions

- Follow [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) patterns
- Use appropriate validation functions
- Provide clear error messages
- Use consistent attribute naming

### Project Structure

```
.
├── internal/
│   └── provider/           # Provider implementation
│       ├── client.go       # Pi-hole API client
│       ├── client_test.go  # Client unit tests
│       ├── provider.go     # Provider definition
│       ├── *_resource.go   # Resource implementations
│       └── *_test.go       # Tests
├── examples/               # Example configurations
└── docs/                   # Documentation
```

## Release Process

### Versioning

This project follows [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Creating a Release

Releases are automated via GitHub Actions when a tag is pushed:

1. Update version references and CHANGELOG.md
2. Create and push a tag: `git tag v0.3.0 && git push origin v0.3.0`
3. GitHub Actions will build and publish the release

## Getting Help

- Open an [issue](https://github.com/lukaspustina/terraform-provider-pihole/issues) for bugs or feature requests
- Check existing issues before creating new ones
- Provide minimal reproduction examples
- Include environment details (Terraform version, Pi-hole version, etc.)

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/). Please be respectful and constructive in all interactions.

## License

By contributing to this project, you agree that your contributions will be licensed under the MIT License.