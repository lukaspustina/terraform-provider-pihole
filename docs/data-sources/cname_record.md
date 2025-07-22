# pihole_cname_record

Retrieves a specific CNAME record from Pi-hole by domain name. This data source is useful for looking up individual CNAME records or validating that specific aliases exist.

## Example Usage

### Basic Usage

```terraform
data "pihole_cname_record" "www" {
  domain = "www.homelab.local"
}

output "www_target" {
  value = data.pihole_cname_record.www.target
}
```

### Using in Other Resources

```terraform
data "pihole_cname_record" "existing_alias" {
  domain = "api.homelab.local"
}

# Create a new CNAME pointing to the same target
resource "pihole_cname_record" "api_v2" {
  domain = "apiv2.homelab.local"
  target = data.pihole_cname_record.existing_alias.target
}

# Create a DNS record for load balancing
resource "pihole_dns_record" "load_balancer" {
  domain = "lb.${data.pihole_cname_record.existing_alias.target}"
  ip     = "192.168.1.150"
}
```

### Alias Chain Discovery

```terraform
data "pihole_cname_record" "first_alias" {
  domain = "app.homelab.local"
}

# Check if the target is also a CNAME (potential chain)
data "pihole_cname_record" "second_alias" {
  domain = data.pihole_cname_record.first_alias.target
}

output "alias_chain" {
  value = {
    original = data.pihole_cname_record.first_alias.domain
    first_target = data.pihole_cname_record.first_alias.target
    is_chained = data.pihole_cname_record.second_alias.target != null
    final_target = data.pihole_cname_record.second_alias.target
  }
}
```

### Conditional Resource Creation

```terraform
data "pihole_cname_record" "web_service" {
  domain = "web.homelab.local"
}

# Only create monitoring if the web service CNAME exists
resource "pihole_dns_record" "web_monitor" {
  count = data.pihole_cname_record.web_service.target != null ? 1 : 0
  
  domain = "monitor-${data.pihole_cname_record.web_service.domain}"
  ip     = "192.168.1.200"
}
```

### Target Validation and Management

```terraform
data "pihole_cname_record" "service_alias" {
  domain = "service.homelab.local"
}

# Ensure the target exists as a DNS record
data "pihole_dns_record" "target_record" {
  domain = data.pihole_cname_record.service_alias.target
}

# Create backup CNAME pointing to backup target
resource "pihole_cname_record" "service_backup" {
  domain = "service-backup.homelab.local"
  target = replace(data.pihole_dns_record.target_record.domain, "server", "backup-server")
  
  depends_on = [
    data.pihole_cname_record.service_alias,
    data.pihole_dns_record.target_record
  ]
}
```

## Schema

### Required Arguments

- `domain` (String) - The CNAME alias domain name to look up. Must be a valid domain name format.

### Read-Only Attributes

- `id` (String) - Data source identifier (same as the domain name)
- `domain` (String) - The CNAME domain name that was looked up (echo of input)
- `target` (String) - The target domain name that the CNAME points to

## Error Handling

If the specified CNAME record doesn't exist, Terraform will produce an error:

```
Error: CNAME Record Not Found
No CNAME record found for domain: nonexistent.homelab.local
```

### Handling Missing Records

You can use conditional logic to handle cases where a CNAME might not exist:

```terraform
# This will fail if the CNAME doesn't exist
# data "pihole_cname_record" "might_not_exist" {
#   domain = "optional.homelab.local"
# }

# Instead, use cname_records data source and filter
data "pihole_cname_records" "all" {}

locals {
  optional_cname = [
    for record in data.pihole_cname_records.all.records :
    record if record.domain == "optional.homelab.local"
  ]
  has_optional_cname = length(local.optional_cname) > 0
  optional_cname_target = local.has_optional_cname ? local.optional_cname[0].target : null
}

output "optional_cname_status" {
  value = {
    exists = local.has_optional_cname
    target = local.optional_cname_target
  }
}
```

## Behavior Notes

- **Real-time Lookup**: This data source performs a fresh lookup on each Terraform run
- **Case Sensitivity**: Domain names are case-insensitive
- **Exact Match**: Only exact domain name matches are returned
- **Performance**: Individual lookups are faster than retrieving all records, but still subject to rate limiting

## Common Patterns

### Service Alias Management

```terraform
# Look up main service alias
data "pihole_cname_record" "main_service" {
  domain = "app.homelab.local"
}

# Create related aliases pointing to the same target
resource "pihole_cname_record" "service_aliases" {
  for_each = toset(["api", "web", "frontend"])
  
  domain = "${each.value}.homelab.local"
  target = data.pihole_cname_record.main_service.target
}
```

### Load Balancer Configuration

```terraform
data "pihole_cname_record" "service" {
  domain = "myapp.homelab.local"
}

# Create multiple CNAMEs for load balancing
resource "pihole_cname_record" "service_instances" {
  count = 3
  
  domain = "myapp-${count.index + 1}.homelab.local"
  target = data.pihole_cname_record.service.target
}
```

### Environment-Based Aliases

```terraform
data "pihole_cname_record" "prod_service" {
  domain = "app.homelab.local"
}

# Create development and staging aliases
resource "pihole_cname_record" "env_aliases" {
  for_each = toset(["dev", "staging"])
  
  domain = "${each.value}-${data.pihole_cname_record.prod_service.domain}"
  target = replace(data.pihole_cname_record.prod_service.target, "prod", each.value)
}
```

### Alias Migration

```terraform
# Look up old alias
data "pihole_cname_record" "old_alias" {
  domain = "legacy.homelab.local"
}

# Create new alias pointing to the same target
resource "pihole_cname_record" "new_alias" {
  domain = "modern.homelab.local"
  target = data.pihole_cname_record.old_alias.target
}

# Output migration information
output "migration_info" {
  value = {
    old_alias = data.pihole_cname_record.old_alias.domain
    new_alias = pihole_cname_record.new_alias.domain
    shared_target = data.pihole_cname_record.old_alias.target
    migration_complete = true
  }
}
```

## Validation Patterns

### Target Existence Check

```terraform
data "pihole_cname_record" "alias" {
  domain = "service.homelab.local"
}

# Validate that the target exists (will fail if target DNS record doesn't exist)
data "pihole_dns_record" "target_validation" {
  domain = data.pihole_cname_record.alias.target
}

output "validation_result" {
  value = "CNAME ${data.pihole_cname_record.alias.domain} -> ${data.pihole_cname_record.alias.target} (${data.pihole_dns_record.target_validation.ip})"
}
```

## Related Resources

- [`pihole_cname_records`](./cname_records.md) - For retrieving all CNAME records
- [`pihole_cname_record` resource](../resources/cname_record.md) - For managing CNAME records
- [`pihole_dns_record`](./dns_record.md) - For retrieving a specific DNS record