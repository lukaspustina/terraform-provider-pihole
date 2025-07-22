# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[Unreleased]: https://github.com/lukaspustina/terraform-provider-pihole/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/lukaspustina/terraform-provider-pihole/releases/tag/v0.1.0
