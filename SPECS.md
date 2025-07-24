# SPECS for Pi Hole Terraform Provider

I like to create a Terraform provider for Pi Hole to manage DNS records and CNAME records using the Pi Hole API version 6.

## Pi Hole API Documentation

* <https://ftl.pi-hole.net/master/docs/>

## Example API Calls

### Authentication

```bash
curl -X POST "https://:443/api/auth" -H 'accept: application/json' -H 'content-type: application/json'  -d '{"password":"XXX"}'
```

### Get DNS Records

```bash
curl -v -X GET --header "X-FTL-SID: XXX" --header "X-FTL-CSRF: XXX" "https://XXX:443/api/config/dns/hosts" -H 'accept: application/json'
```

### PUT DNS Record

```bash
 curl -v -X PUT --header "X-FTL-SID: XXX" --header "X-FTL-CSRF: XXX" "https://XXX:443/api/config/dns/hosts/192.168.0.22%20www.homelab.local" -H 'accept: application/json'
```

### Get CNAME Records

```bash
curl -v -X GET --header "X-FTL-SID: XXX" --header "X-FTL-CSRF: XXX" "https://XXX:443/api/config/dns/cnameRecords" -H 'accept: application/json'
```

### PUT CNAME Record

```bash
curl -v -X PUT --header "X-FTL-SID: XXX" --header "X-FTL-CSRF: XXX" "https://XXX:443/api/config/dns/cnameRecords/web,www" -H 'accept: application/json'
```

### Get Webserver Configuration

```bash
# Get webserver configuration (contains webserver.api.app_sudo)
curl -v -X GET --header "X-FTL-SID: XXX" --header "X-FTL-CSRF: XXX" "https://XXX:443/api/config/webserver" -H 'accept: application/json'
```

### Set Webserver Configuration

```bash
# Update webserver configuration (including webserver.api.app_sudo)
curl -v -X PUT --header "X-FTL-SID: XXX" --header "X-FTL-CSRF: XXX" "https://XXX:443/api/config/webserver" -H 'accept: application/json' -H 'content-type: application/json' -d '{"api": {"app_sudo": true}}'
```


## Provider

```hcl
provider "pihole" {
  url         = "https://example.com"
  password    = "xxx"
  insecure_tls = false  # Optional: Skip TLS certificate verification (default: false)
}
```

### Provider Configuration Options

- `url` (Required): Pi-hole server URL
- `password` (Required): Pi-hole admin password
- `insecure_tls` (Optional): Skip TLS certificate verification. Defaults to `false` for secure connections
- `max_connections` (Optional): Maximum concurrent connections (default: 1)  
- `request_delay_ms` (Optional): Delay between requests in milliseconds (default: 300)
- `retry_attempts` (Optional): Number of retry attempts (default: 3)
- `retry_backoff_base_ms` (Optional): Base retry backoff delay in milliseconds (default: 500)

## Resource - DNS Record

```hcl
resource "pihole_dns_record" "www" {
  domain = "www.homelab.local"
  ip     = "192.168.0.22"
}
```

## Resource - CNAME Record

```hcl
resource "pihole_cname_record" "web" {
  domain = "web.homelab.local"
  target = "www.homelab.local"
}
```

## Resource - Webserver Configuration Setting

**Important**: Webserver configuration changes require admin password, not application password. Application passwords cannot modify Pi-hole webserver configuration settings unless `webserver.api.app_sudo` is enabled.

```hcl
resource "pihole_config" "app_sudo" {
  key   = "webserver.api.app_sudo"
  value = "true"
}
```

## Data Source - Webserver Configuration Setting

```hcl
data "pihole_config" "app_sudo_status" {
  key = "webserver.api.app_sudo"
}

# Use the webserver configuration value in other resources
output "app_sudo_enabled" {
  value = data.pihole_config.app_sudo_status.value == "true"
}
```

## Data Source - DNS Records (All)

```hcl
# Get all DNS A records
data "pihole_dns_records" "all_dns" {}

# Use the DNS records in other resources
output "dns_record_count" {
  value = length(data.pihole_dns_records.all_dns.records)
}

# Filter for specific records
locals {
  homelab_records = [
    for record in data.pihole_dns_records.all_dns.records :
    record if can(regex("homelab\\.local$", record.domain))
  ]
}
```

## Data Source - DNS Record (Single)

```hcl
# Look up a specific DNS record
data "pihole_dns_record" "server" {
  domain = "server.homelab.local"
}

# Use the record in other resources
resource "local_file" "server_config" {
  filename = "server.env"
  content  = "SERVER_IP=${data.pihole_dns_record.server.ip}\n"
}
```

## Data Source - CNAME Records (All)

```hcl
# Get all CNAME records
data "pihole_cname_records" "all_cnames" {}

# Create SSL certificate requests for all web-related CNAMEs
locals {
  web_cnames = [
    for record in data.pihole_cname_records.all_cnames.records :
    record if can(regex("^(www|blog|mail)\\.", record.domain))
  ]
}
```

## Data Source - CNAME Record (Single)

```hcl
# Look up a specific CNAME record
data "pihole_cname_record" "www" {
  domain = "www.homelab.local"
}

# Validate that the CNAME points to the expected target
locals {
  www_target_valid = data.pihole_cname_record.www.target == "server.homelab.local"
}
```

## Security Considerations

### Admin vs Application Password

**⚠️ Critical Security Information:**

When using Pi-hole **application passwords** (not the main admin password), **NO modifications are possible** unless `webserver.api.app_sudo` is enabled:

- **❌ DNS/CNAME Records**: Application passwords **cannot** modify DNS or CNAME records unless `webserver.api.app_sudo` is enabled
- **❌ Webserver Configuration Changes**: Application passwords **cannot** modify Pi-hole webserver configuration settings unless `webserver.api.app_sudo` is enabled
- **🔐 Solution**: Use admin password to enable `webserver.api.app_sudo` first, then application passwords can modify all settings:

```hcl
# This requires admin password to set initially
resource "pihole_config" "enable_app_sudo" {
  key   = "webserver.api.app_sudo"  
  value = "true"
}
```

**Alternative**: Enable "Permit destructive actions via API" in Pi-hole web interface (Settings → API/Web interface)

### TLS Certificate Security

**⚠️ Security Warning:**

The `insecure_tls` parameter should be used with caution:

```hcl
provider "pihole" {
  url          = "https://pihole.homelab.local:443"
  password     = "your-admin-password-here"
  insecure_tls = true  # ⚠️ Only for self-signed certificates
}
```

**Security Guidelines:**
- **Default**: `insecure_tls = false` (secure TLS verification)
- **Only use `insecure_tls = true`** for local Pi-hole installations with self-signed certificates
- **Never use in production** environments or public-facing Pi-hole instances
- **Alternative**: Use proper SSL certificates from Let's Encrypt or CA

## Troubleshooting

### Authentication Issues

- Ensure your Pi-hole admin password is correct
- Check that your Pi-hole URL is accessible and uses the correct protocol (HTTP/HTTPS)
- Verify that API access is enabled in Pi-hole settings

### Webserver Configuration Access Denied

If you get errors like "app session is not allowed to modify Pi-hole config settings":

1. **Check password type**: Are you using an application password or admin password?
2. **Enable app_sudo**: Set `webserver.api.app_sudo = true` using admin password
3. **Web UI alternative**: Enable "Permit destructive actions via API" in Pi-hole web interface

### TLS Certificate Issues

**For self-signed certificates:**
```hcl
provider "pihole" {
  url          = "https://pihole.homelab.local:443"
  password     = "your-password"
  insecure_tls = true  # Allow self-signed certificates
}
```

**For certificate verification errors:**
- Verify your Pi-hole URL is correct
- Check if your Pi-hole uses HTTP or HTTPS
- Consider using proper SSL certificates

### Connection Issues

- **Rate limiting**: Increase `request_delay_ms` if you're hitting rate limits
- **Timeouts**: Increase `retry_attempts` and `retry_backoff_base_ms` for unstable connections
- **Concurrent access**: Adjust `max_connections` if needed (default: 1)
