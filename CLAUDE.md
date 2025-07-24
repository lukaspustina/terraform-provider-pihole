# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a fully functional Terraform provider for managing Pi-hole DNS records, CNAME records, and configuration settings using Pi-hole API version 6. The provider includes comprehensive resource management, data sources, and robust error handling.

## Project Specifications

The complete project specifications are documented in `SPECS.md`, which includes:

- Pi-hole API documentation reference
- Authentication examples using curl
- API endpoints for DNS and CNAME record management
- Proposed provider configuration syntax
- Resource examples for DNS and CNAME records

## Provider Features

- **Resources**: Manage Pi-hole DNS A records via `pihole_dns_record` resource
- **Resources**: Manage Pi-hole CNAME records via `pihole_cname_record` resource
- **Resources**: Manage Pi-hole configuration settings via `pihole_config` resource
- **Data Sources**: Query DNS records with `pihole_dns_records` (all) and `pihole_dns_record` (single)
- **Data Sources**: Query CNAME records with `pihole_cname_records` (all) and `pihole_cname_record` (single)
- **Data Sources**: Query configuration settings with `pihole_config` data source
- **Security**: Support both admin passwords and application passwords (with limitations)
- **TLS**: Configurable TLS verification with secure defaults
- **Reliability**: Automatic retries, rate limiting, and connection management
- **Compatibility**: Full Pi-hole API v6 support

## API Integration

The provider authenticates using:
- POST to `/api/auth` with password (admin or application)
- Uses returned session ID and CSRF token for subsequent requests
- Target endpoints:
  - `/api/config/dns/hosts` - DNS A record management
  - `/api/config/dns/cnameRecords` - CNAME record management
  - `/api/config/{key}` - Configuration management
  - `/api/config/webserver` - Webserver configuration

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
│   ├── cname_record_resource.go             # CNAME record resource
│   ├── config_resource.go                   # Configuration management resource
│   ├── dns_records_data_source.go           # DNS records data source (all)
│   ├── dns_record_data_source.go            # DNS record data source (single)
│   ├── cname_records_data_source.go         # CNAME records data source (all)
│   ├── cname_record_data_source.go          # CNAME record data source (single)
│   ├── config_data_source.go                # Configuration data source
│   └── *_test.go                            # Comprehensive test files
├── docs/                                    # Terraform registry documentation
│   ├── index.md                            # Provider documentation
│   ├── resources/                          # Resource documentation
│   └── data-sources/                       # Data source documentation
├── examples/                                # Usage examples
├── go.mod                                   # Go module definition
├── Makefile                                 # Build automation
├── SPECS.md                                 # Project specifications
├── README.md                                # Project README
└── CLAUDE.md                                # This file
```

## Implementation Details

The provider uses the Terraform Plugin Framework and implements:

- **Authentication**: Sessions via POST `/api/auth` with password
- **DNS Records**: Manage A records via `/api/config/dns/hosts`
- **CNAME Records**: Manage CNAME records via `/api/config/dns/cnameRecords`
- **Configuration**: Manage settings via `/api/config/{key}` endpoints
- **TLS**: Configurable TLS verification (default: secure, optional: insecure for self-signed)
- **Error Handling**: Automatic retries with exponential backoff
- **Rate Limiting**: Built-in delays to prevent API overload
- **Connection Management**: Configurable connection limits and session caching

## Security Considerations

**Important**: When using application passwords (not admin passwords):
- DNS and CNAME record management works normally
- Configuration changes require `webserver.api.app_sudo = true` to be set first
- This setting must be enabled using an admin password initially
- Alternative: Enable "Permit destructive actions via API" in Pi-hole web interface

## Development Best Practices

- Always run linter and tests using the Makefile
- Use `make fmt` to format code before committing
- Run `make check` to run all quality checks (fmt, vet, test)
- Ensure all tests pass before submitting changes
- Follow existing code patterns and naming conventions