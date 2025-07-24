# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 24.07.2025

### Added
- **Configuration Management**: Added `pihole_config` resource and data source for managing Pi-hole webserver configuration settings
  - Support for `webserver.api.app_sudo` configuration to enable application password permissions
  - Comprehensive documentation about admin vs application password requirements
- **Security Enhancements**: 
  - TLS certificate verification is now secure by default (changed from insecure hardcoded behavior)
  - Added optional `insecure_tls` parameter for self-signed certificates
  - Enhanced security warnings and documentation

### Changed
- **Breaking Change**: TLS certificate verification now defaults to secure (was previously insecure by default)
- **Configuration API**: Webserver configuration management now properly handles nested JSON structure navigation

## [0.2.0] - 22.07.2025

### Added
- **Data Sources**: Added comprehensive data sources for reading existing Pi-hole DNS configuration
  - `pihole_dns_records` - Retrieve all DNS A records from Pi-hole
  - `pihole_cname_records` - Retrieve all CNAME records from Pi-hole
  - `pihole_dns_record` - Retrieve a specific DNS A record by domain name
  - `pihole_cname_record` - Retrieve a specific CNAME record by domain name
- **Documentation**: Added Terraform registry compatible documentation
- **Testing**: Added comprehensive acceptance tests for all data sources
- **Examples**: Added usage examples for all data sources with various scenarios

### Improved
- **Test Coverage**: Enhanced test coverage with unit and acceptance tests
- **Error Handling**: Improved error handling for missing records and edge cases
- **Code Quality**: Added comprehensive linting and formatting checks

### Fixed
- **GoReleaser**: Removed 386 and FreeBSD build targets as requested
- **Test Reliability**: Fixed acceptance test issues with Pi-hole behavior

## [0.1.0] - 22.07.2025

### Added
- Initial release of Pi-hole Terraform provider
- Support for Pi-hole v6 API
- DNS A record management via `pihole_dns_record` resource
- CNAME record management via `pihole_cname_record` resource
- Session-based authentication with Pi-hole v6
- Resource import functionality using domain names
- Client-side validation for domain names and IP addresses
- Comprehensive test coverage including acceptance tests
- Retry logic with exponential backoff for resilient API calls
- Session caching to reduce API load

### Features
- **Resources:**
  - `pihole_dns_record`: Manage DNS A records
  - `pihole_cname_record`: Manage CNAME records
- **Provider Configuration:**
  - `url`: Pi-hole server URL
  - `password`: Pi-hole admin password
  - `max_connections`: Connection pool configuration
  - `request_delay_ms`: Rate limiting
  - `retry_attempts`: Error retry configuration
  - `retry_backoff_base_ms`: Retry timing configuration

### Technical Details
- Uses Terraform Plugin Framework v1.15.0
- Compatible with Pi-hole v6 API
- Supports import of existing records
- Implements proper resource lifecycle management
- Includes comprehensive validation and error handling

[Unreleased]: https://github.com/lukaspustina/terraform-provider-pihole/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/lukaspustina/terraform-provider-pihole/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/lukaspustina/terraform-provider-pihole/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/lukaspustina/terraform-provider-pihole/releases/tag/v0.1.0
