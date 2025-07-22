# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform provider project for managing Pi-hole DNS records and CNAME records using Pi-hole API version 6. The project is in its initial specification phase.

## Project Specifications

The complete project specifications are documented in `SPECS.md`, which includes:

- Pi-hole API documentation reference
- Authentication examples using curl
- API endpoints for DNS and CNAME record management
- Proposed provider configuration syntax
- Resource examples for DNS and CNAME records

## Provider Goals

- Manage Pi-hole DNS A records via `pihole_dns_record` resource
- Manage Pi-hole CNAME records via `pihole_cname_record` resource  
- Support authentication using Pi-hole admin password
- Target Pi-hole API v6 endpoints

## API Integration

The provider will authenticate using:
- POST to `/api/auth` with password
- Use returned session ID and CSRF token for subsequent requests
- Target endpoints: `/api/config/dns/hosts` and `/api/config/dns/cnameRecords`

## Development Commands

- `make build` - Build the provider binary
- `make install` - Install the provider locally for testing  
- `make test` - Run unit tests
- `make fmt` - Format Go code
- `make vet` - Run go vet for static analysis
- `make check` - Run all quality checks (fmt, vet, test)
- `make dev` - Build and run provider in debug mode
- `go mod tidy` - Update dependencies

## Project Structure

```
├── main.go                                    # Provider entry point
├── internal/provider/
│   ├── provider.go                           # Main provider configuration
│   ├── client.go                            # Pi-hole API client
│   ├── dns_record_resource.go               # DNS A record resource
│   └── cname_record_resource.go             # CNAME record resource
├── go.mod                                   # Go module definition
├── Makefile                                 # Build automation
└── SPECS.md                                 # Project specifications
```

## Implementation Details

The provider uses the Terraform Plugin Framework and implements:

- **Authentication**: Sessions via POST `/api/auth` with password
- **DNS Records**: Manage A records via `/api/config/dns/hosts`
- **CNAME Records**: Manage CNAME records via `/api/config/dns/cnameRecords`
- **TLS**: Accepts self-signed certificates (InsecureSkipVerify: true)

## Development Best Practices

- Always run linter and tests
- Always use the Makefile to run linter and tests