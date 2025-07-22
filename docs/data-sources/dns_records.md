# pihole_dns_records

Retrieves all DNS A records from Pi-hole. This data source is useful for discovering existing DNS records or using them in other resources.

## Example Usage

### Basic Usage

```terraform
data "pihole_dns_records" "all" {}

output "dns_records" {
  value = data.pihole_dns_records.all.records
}
```

### Using DNS Records in Resources

```terraform
data "pihole_dns_records" "existing" {}

locals {
  existing_domains = [for record in data.pihole_dns_records.existing.records : record.domain]
}

# Create CNAME records for all existing DNS records
resource "pihole_cname_record" "www_aliases" {
  for_each = {
    for record in data.pihole_dns_records.existing.records :
    record.domain => record.domain
    if can(regex("^[^.]+\\.homelab\\.local$", record.domain))
  }
  
  domain = "www.${each.value}"
  target = each.value
}
```

### Filtering and Processing

```terraform
data "pihole_dns_records" "all" {}

locals {
  # Filter records by IP range
  local_records = [
    for record in data.pihole_dns_records.all.records :
    record if can(regex("^192\\.168\\.", record.ip))
  ]
  
  # Group records by IP
  records_by_ip = {
    for record in data.pihole_dns_records.all.records :
    record.ip => record.domain...
  }
}

output "local_dns_records" {
  value = local.local_records
}

output "grouped_records" {
  value = local.records_by_ip
}
```

## Schema

### Read-Only Attributes

- `id` (String) - Data source identifier (always "dns_records")
- `records` (List of Object) - List of DNS A records, where each record contains:
  - `domain` (String) - The fully qualified domain name
  - `ip` (String) - The IPv4 address that the domain resolves to

## Behavior Notes

- **Real-time Data**: This data source fetches the current state of DNS records from Pi-hole on each Terraform run
- **No Caching**: Records are not cached between Terraform runs, ensuring you always get the most up-to-date information
- **Performance**: For large numbers of DNS records, this operation may take some time due to rate limiting
- **Ordering**: Records are returned in the order provided by the Pi-hole API (typically insertion order)

## Use Cases

### Discovery and Documentation

Use this data source to discover what DNS records exist:

```terraform
data "pihole_dns_records" "audit" {}

output "dns_inventory" {
  value = {
    total_records = length(data.pihole_dns_records.audit.records)
    domains = [for r in data.pihole_dns_records.audit.records : r.domain]
    ip_addresses = distinct([for r in data.pihole_dns_records.audit.records : r.ip])
  }
}
```

### Validation and Checks

Validate that certain records exist:

```terraform
data "pihole_dns_records" "check" {}

locals {
  required_domains = ["gateway.homelab.local", "dns.homelab.local"]
  existing_domains = [for r in data.pihole_dns_records.check.records : r.domain]
  missing_domains = setsubtract(local.required_domains, local.existing_domains)
}

# This will cause an error if any required domains are missing
resource "null_resource" "validate_required_domains" {
  count = length(local.missing_domains) > 0 ? 1 : 0
  
  provisioner "local-exec" {
    command = "echo 'Missing required DNS records: ${join(", ", local.missing_domains)}' && exit 1"
  }
}
```

### Dynamic Resource Creation

Create resources based on existing records:

```terraform
data "pihole_dns_records" "servers" {}

# Create backup CNAME records for all server records
resource "pihole_cname_record" "backups" {
  for_each = {
    for record in data.pihole_dns_records.servers.records :
    record.domain => record.domain
    if can(regex("server", record.domain))
  }
  
  domain = replace(each.value, "server", "server-backup")
  target = each.value
}
```

## Error Handling

Common errors and their meanings:

- **Authentication failed**: Pi-hole admin password is incorrect or API access is disabled
- **Connection timeout**: Pi-hole server is unreachable or overloaded
- **Client Error**: Network issues or Pi-hole API errors

The provider includes automatic retry logic with exponential backoff for transient network errors.

## Related Resources

- [`pihole_dns_record`](./dns_record.md) - For retrieving a specific DNS record
- [`pihole_dns_record` resource](../resources/dns_record.md) - For managing DNS A records
- [`pihole_cname_records`](./cname_records.md) - For retrieving all CNAME records