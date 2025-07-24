# Pi-hole Provider

The Pi-hole provider allows you to manage Pi-hole DNS records and CNAME records using the Pi-hole API v6.

This provider enables Infrastructure as Code (IaC) practices for your Pi-hole DNS configuration, making it easy to version control and automate your local DNS setup.

## Example Usage

```hcl
terraform {
  required_providers {
    pihole = {
      source  = "registry.terraform.io/lukaspustina/pihole"
      version = "0.2.0"
    }
  }
}

provider "pihole" {
  url      = "https://pihole.homelab.local:443"
  password = var.pihole_password
}

# DNS A Record
resource "pihole_dns_record" "server" {
  domain = "server.homelab.local"
  ip     = "192.168.1.100"
}

# CNAME Record
resource "pihole_cname_record" "www" {
  domain = "www.homelab.local"
  target = "server.homelab.local"
}

# Webserver Configuration Management (requires admin password)
resource "pihole_config" "enable_app_sudo" {
  key   = "webserver.api.app_sudo"
  value = "true"
}

# Data Sources
data "pihole_dns_records" "existing" {}

data "pihole_cname_record" "web_alias" {
  domain = "www.homelab.local"
}

data "pihole_config" "app_sudo_status" {
  key = "webserver.api.app_sudo"
}

# Additional data source examples
data "pihole_dns_records" "all_dns" {}
data "pihole_cname_records" "all_cnames" {}
```

## Schema

### Required

- `url` (String) - Pi-hole server URL (e.g., `https://pihole.homelab.local:443`)
- `password` (String, Sensitive) - Pi-hole admin password

### Optional

- `insecure_tls` (Boolean) - Skip TLS certificate verification. Default: `false`
- `max_connections` (Number) - Maximum number of concurrent connections to Pi-hole. Default: `1`
- `request_delay_ms` (Number) - Delay in milliseconds between API requests. Default: `300`
- `retry_attempts` (Number) - Number of retry attempts for failed requests. Default: `3`
- `retry_backoff_base_ms` (Number) - Base delay in milliseconds for retry backoff. Default: `500`

## Features

### Resources
- **DNS A Records**: Manage custom DNS A records that resolve domain names to IP addresses
- **CNAME Records**: Manage CNAME aliases that point to other domain names
- **Webserver Configuration Settings**: Manage Pi-hole webserver configuration (requires admin password)

### Data Sources
- **DNS Records Discovery**: Retrieve all existing DNS A records from Pi-hole
- **CNAME Records Discovery**: Retrieve all existing CNAME records from Pi-hole
- **Individual Record Lookup**: Look up specific DNS or CNAME records by domain name
- **Webserver Configuration Reading**: Read current Pi-hole webserver configuration settings

### Technical Features
- **Pi-hole API v6 Compatible**: Full compatibility with modern Pi-hole installations
- **Robust Error Handling**: Automatic retries with exponential backoff
- **Rate Limited**: Built-in request delays prevent API overload
- **TLS Support**: Secure TLS verification by default, with optional bypass for self-signed certificates
- **Connection Management**: Configurable connection limits and retry behavior

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- Pi-hole installation with API v6 support
- Pi-hole admin password

## API Compatibility

This provider targets **Pi-hole API v6** and uses the following endpoints:

- `POST /api/auth` - Authentication
- `GET /api/config/dns/hosts` - Retrieve DNS records  
- `PUT /api/config/dns/hosts/{ip}%20{domain}` - Create/update DNS records
- `DELETE /api/config/dns/hosts/{ip}%20{domain}` - Delete DNS records
- `GET /api/config/dns/cnameRecords` - Retrieve CNAME records
- `PUT /api/config/dns/cnameRecords/{domain},{target}` - Create/update CNAME records  
- `DELETE /api/config/dns/cnameRecords/{domain},{target}` - Delete CNAME records
- `GET /api/config/webserver` - Retrieve webserver configuration settings
- `PUT /api/config/webserver` - Update webserver configuration settings

## Advanced Configuration

For busy Pi-hole instances or unstable network connections, you can adjust the connection parameters:

```hcl
provider "pihole" {
  url      = "https://pihole.homelab.local:443"
  password = var.pihole_password
  
  # TLS configuration for self-signed certificates
  insecure_tls = true
  
  # Slower requests for busy Pi-hole
  request_delay_ms       = 500
  
  # More retries for unstable connections  
  retry_attempts         = 5
  retry_backoff_base_ms  = 1000
}
```

## Troubleshooting

### Connection Issues

The provider includes automatic retry logic with exponential backoff and rate limiting to prevent overwhelming Pi-hole. If you experience persistent connection issues:

1. Verify your Pi-hole URL is accessible
2. Check that your admin password is correct
3. Ensure API access is enabled in Pi-hole settings
4. Try increasing `request_delay_ms` and `retry_attempts`

### Authentication Issues

- Ensure your Pi-hole admin password is correct
- Verify that your Pi-hole URL uses the correct protocol (HTTP/HTTPS)
- Check that API access is enabled in Pi-hole admin interface

### TLS Certificate Issues

By default, the provider verifies TLS certificates for secure connections. If your Pi-hole uses self-signed certificates, you can disable certificate verification:

```hcl
provider "pihole" {
  url          = "https://pihole.homelab.local:443"
  password     = var.pihole_password
  insecure_tls = true  # Allow self-signed certificates
}
```

**Security Note**: Only use `insecure_tls = true` for local Pi-hole installations with self-signed certificates. For production environments, keep the default secure verification.

### Webserver Configuration Management Issues

If you're using a Pi-hole **application password** (not the main admin password), **NO modifications are possible** unless `webserver.api.app_sudo` is enabled:

- **DNS/CNAME operations**: Application passwords **cannot** modify DNS or CNAME records unless `webserver.api.app_sudo` is enabled
- **Webserver configuration changes**: Application passwords **cannot** modify webserver configuration unless `webserver.api.app_sudo` is enabled
- **Solution**: Use admin password to enable `webserver.api.app_sudo` first (this allows application passwords to modify all settings), or enable "Permit destructive actions via API" in Pi-hole web interface

```hcl
# Enable all modifications for application passwords (requires admin password initially)
resource "pihole_config" "enable_app_sudo" {
  key   = "webserver.api.app_sudo"
  value = "true"
}
```

## Development and Testing

### Running Tests

The provider includes comprehensive unit tests that can be run without a live Pi-hole instance:

```bash
# Run all tests
make test

# Run specific tests
go test -v ./internal/provider -run TestConfigResource

# Format code
make fmt

# Run linter
make check
```

### Local Development

```bash
# Build the provider
make build

# Install locally for testing
make install

# Run in development mode
make dev
```