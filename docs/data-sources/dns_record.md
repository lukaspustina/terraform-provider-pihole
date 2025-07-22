# pihole_dns_record

Retrieves a specific DNS A record from Pi-hole by domain name. This data source is useful for looking up individual DNS records or validating that specific records exist.

## Example Usage

### Basic Usage

```terraform
data "pihole_dns_record" "server" {
  domain = "server.homelab.local"
}

output "server_ip" {
  value = data.pihole_dns_record.server.ip
}
```

### Using in Other Resources

```terraform
data "pihole_dns_record" "main_server" {
  domain = "server.homelab.local"
}

# Create a CNAME pointing to the existing server
resource "pihole_cname_record" "www" {
  domain = "www.homelab.local"
  target = data.pihole_dns_record.main_server.domain
}

# Create a backup server with a related IP
resource "pihole_dns_record" "backup" {
  domain = "backup.homelab.local"
  ip     = replace(data.pihole_dns_record.main_server.ip, "/\\.100$/", ".101")
}
```

### Conditional Logic

```terraform
data "pihole_dns_record" "gateway" {
  domain = "gateway.homelab.local"
}

# Only create monitoring if gateway exists and is in the expected subnet
resource "pihole_dns_record" "monitoring" {
  count = can(regex("^192\\.168\\.1\\.", data.pihole_dns_record.gateway.ip)) ? 1 : 0
  
  domain = "monitor.homelab.local"
  ip     = "192.168.1.200"
}
```

### Validation and Dependencies

```terraform
# Ensure a critical DNS record exists before proceeding
data "pihole_dns_record" "critical_service" {
  domain = "api.homelab.local"
}

# This will only succeed if the DNS record exists
resource "pihole_cname_record" "api_alias" {
  domain = "service.homelab.local"
  target = data.pihole_dns_record.critical_service.domain
  
  # Explicit dependency to ensure order
  depends_on = [data.pihole_dns_record.critical_service]
}
```

### IP Address Manipulation

```terraform
data "pihole_dns_record" "base_server" {
  domain = "server01.homelab.local"
}

locals {
  # Extract the base IP and increment for additional servers
  base_ip_parts = split(".", data.pihole_dns_record.base_server.ip)
  base_network = "${local.base_ip_parts[0]}.${local.base_ip_parts[1]}.${local.base_ip_parts[2]}"
  base_host = tonumber(local.base_ip_parts[3])
}

# Create additional servers with incremented IPs
resource "pihole_dns_record" "additional_servers" {
  count = 3
  
  domain = "server${format("%02d", count.index + 2)}.homelab.local"
  ip     = "${local.base_network}.${local.base_host + count.index + 1}"
}
```

## Schema

### Required Arguments

- `domain` (String) - The fully qualified domain name to look up. Must be a valid domain name format.

### Read-Only Attributes

- `id` (String) - Data source identifier (same as the domain name)
- `domain` (String) - The domain name that was looked up (echo of input)
- `ip` (String) - The IPv4 address that the domain resolves to

## Import

This is a data source and cannot be imported, but you can reference existing DNS records managed outside of Terraform.

## Error Handling

If the specified DNS record doesn't exist, Terraform will produce an error:

```
Error: DNS Record Not Found
No DNS record found for domain: nonexistent.homelab.local
```

### Handling Missing Records

You can use conditional logic to handle cases where a record might not exist:

```terraform
# This will fail if the record doesn't exist
# data "pihole_dns_record" "might_not_exist" {
#   domain = "optional.homelab.local"
# }

# Instead, use dns_records data source and filter
data "pihole_dns_records" "all" {}

locals {
  optional_record = [
    for record in data.pihole_dns_records.all.records :
    record if record.domain == "optional.homelab.local"
  ]
  has_optional_record = length(local.optional_record) > 0
  optional_record_ip = local.has_optional_record ? local.optional_record[0].ip : null
}

output "optional_record_status" {
  value = {
    exists = local.has_optional_record
    ip = local.optional_record_ip
  }
}
```

## Behavior Notes

- **Real-time Lookup**: This data source performs a fresh lookup on each Terraform run
- **Case Sensitivity**: Domain names are case-insensitive
- **Exact Match**: Only exact domain name matches are returned
- **Performance**: Individual lookups are faster than retrieving all records, but still subject to rate limiting

## Common Patterns

### Service Discovery

```terraform
# Look up service endpoints
data "pihole_dns_record" "database" {
  domain = "db.homelab.local"
}

data "pihole_dns_record" "cache" {
  domain = "redis.homelab.local"  
}

# Use in application configuration
resource "local_file" "app_config" {
  filename = "config.env"
  content = <<-EOF
    DATABASE_HOST=${data.pihole_dns_record.database.ip}
    CACHE_HOST=${data.pihole_dns_record.cache.ip}
  EOF
}
```

### Network Planning

```terraform
data "pihole_dns_record" "router" {
  domain = "router.homelab.local"
}

locals {
  # Determine network configuration from router IP
  network_parts = split(".", data.pihole_dns_record.router.ip)
  network_base = "${local.network_parts[0]}.${local.network_parts[1]}.${local.network_parts[2]}"
  
  # Define server IP range based on router
  server_ips = [
    for i in range(10, 20) : "${local.network_base}.${i}"
  ]
}
```

## Related Resources

- [`pihole_dns_records`](./dns_records.md) - For retrieving all DNS records
- [`pihole_dns_record` resource](../resources/dns_record.md) - For managing DNS A records
- [`pihole_cname_record`](./cname_record.md) - For retrieving a specific CNAME record