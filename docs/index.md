# Pi-hole Provider

The Pi-hole provider allows you to manage Pi-hole DNS records and CNAME records using the Pi-hole API v6.

This provider enables Infrastructure as Code (IaC) practices for your Pi-hole DNS configuration, making it easy to version control and automate your local DNS setup.

## Example Usage

```terraform
terraform {
  required_providers {
    pihole = {
      source  = "lukaspustina/pihole"
      version = "~> 0.1"
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
```

## Schema

### Required

- `url` (String) - Pi-hole server URL (e.g., `https://pihole.homelab.local:443`)
- `password` (String, Sensitive) - Pi-hole admin password

### Optional

- `max_connections` (Number) - Maximum number of concurrent connections to Pi-hole. Default: `1`
- `request_delay_ms` (Number) - Delay in milliseconds between API requests. Default: `300`
- `retry_attempts` (Number) - Number of retry attempts for failed requests. Default: `3`
- `retry_backoff_base_ms` (Number) - Base delay in milliseconds for retry backoff. Default: `500`

## Features

- **DNS A Records**: Manage custom DNS A records that resolve domain names to IP addresses
- **CNAME Records**: Manage CNAME aliases that point to other domain names
- **Pi-hole API v6 Compatible**: Full compatibility with modern Pi-hole installations
- **Robust Error Handling**: Automatic retries with exponential backoff
- **Rate Limited**: Built-in request delays prevent API overload
- **TLS Support**: Works with HTTPS Pi-hole installations including self-signed certificates
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

## Advanced Configuration

For busy Pi-hole instances or unstable network connections, you can adjust the connection parameters:

```terraform
provider "pihole" {
  url      = "https://pihole.homelab.local:443"
  password = var.pihole_password
  
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

The provider accepts self-signed certificates by default, which is common for local Pi-hole installations.