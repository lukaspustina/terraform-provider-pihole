# pihole_cname_records

Retrieves all CNAME records from Pi-hole. This data source is useful for discovering existing CNAME aliases or using them in other resources.

## Example Usage

### Basic Usage

```terraform
data "pihole_cname_records" "all" {}

output "cname_records" {
  value = data.pihole_cname_records.all.records
}
```

### Using CNAME Records in Resources

```terraform
data "pihole_cname_records" "existing" {}

locals {
  # Get all target domains that have CNAME aliases
  aliased_domains = distinct([for record in data.pihole_cname_records.existing.records : record.target])
}

# Create monitoring records for all aliased domains
resource "pihole_dns_record" "monitoring" {
  for_each = toset(local.aliased_domains)
  
  domain = "monitor.${each.value}"
  ip     = "192.168.1.200"
}
```

### Analyzing CNAME Chains

```terraform
data "pihole_cname_records" "all" {}

locals {
  # Find domains that are both CNAME sources and targets (potential chains)
  cname_sources = [for r in data.pihole_cname_records.all.records : r.domain]
  cname_targets = [for r in data.pihole_cname_records.all.records : r.target]
  potential_chains = setintersection(cname_sources, cname_targets)
}

output "cname_analysis" {
  value = {
    total_cnames = length(data.pihole_cname_records.all.records)
    unique_targets = length(distinct(local.cname_targets))
    potential_chains = local.potential_chains
  }
}
```

## Schema

### Read-Only Attributes

- `id` (String) - Data source identifier (always "cname_records")
- `records` (List of Object) - List of CNAME records, where each record contains:
  - `domain` (String) - The CNAME alias domain name
  - `target` (String) - The target domain name that the CNAME points to

## Behavior Notes

- **Real-time Data**: This data source fetches the current state of CNAME records from Pi-hole on each Terraform run
- **No Caching**: Records are not cached between Terraform runs, ensuring you always get the most up-to-date information
- **Performance**: For large numbers of CNAME records, this operation may take some time due to rate limiting
- **Ordering**: Records are returned in the order provided by the Pi-hole API (typically insertion order)

## Use Cases

### Discovery and Documentation

Use this data source to understand your CNAME structure:

```terraform
data "pihole_cname_records" "audit" {}

output "cname_inventory" {
  value = {
    total_aliases = length(data.pihole_cname_records.audit.records)
    alias_domains = [for r in data.pihole_cname_records.audit.records : r.domain]
    target_domains = distinct([for r in data.pihole_cname_records.audit.records : r.target])
    aliases_per_target = {
      for target in distinct([for r in data.pihole_cname_records.audit.records : r.target]) :
      target => [for r in data.pihole_cname_records.audit.records : r.domain if r.target == target]
    }
  }
}
```

### Validation and Consistency Checks

Validate CNAME targets exist as DNS records:

```terraform
data "pihole_dns_records" "dns_records" {}
data "pihole_cname_records" "cname_records" {}

locals {
  dns_domains = [for r in data.pihole_dns_records.dns_records.records : r.domain]
  cname_targets = distinct([for r in data.pihole_cname_records.cname_records.records : r.target])
  
  # Find CNAME targets that don't have corresponding DNS A records (external targets are OK)
  internal_targets = [for target in local.cname_targets : target if can(regex("\\.homelab\\.local$", target))]
  missing_targets = setsubtract(local.internal_targets, local.dns_domains)
}

output "cname_validation" {
  value = {
    total_cname_targets = length(local.cname_targets)
    internal_targets = length(local.internal_targets)
    missing_internal_targets = local.missing_targets
    validation_passed = length(local.missing_targets) == 0
  }
}
```

### Cleanup and Maintenance

Identify unused patterns or inconsistencies:

```terraform
data "pihole_cname_records" "all" {}

locals {
  # Find CNAME records that might be temporary or testing records
  temp_records = [
    for record in data.pihole_cname_records.all.records :
    record if can(regex("(test|temp|staging)", record.domain))
  ]
  
  # Find CNAME records pointing to external domains
  external_targets = [
    for record in data.pihole_cname_records.all.records :
    record if !can(regex("\\.homelab\\.local$", record.target))
  ]
}

output "maintenance_info" {
  value = {
    temporary_records = local.temp_records
    external_references = local.external_targets
  }
}
```

### Dynamic Resource Management

Create additional resources based on CNAME patterns:

```terraform
data "pihole_cname_records" "web_services" {}

# Create SSL certificate requests for all web-related CNAMEs
resource "local_file" "ssl_domains" {
  filename = "ssl-domains.txt"
  content = join("\n", [
    for record in data.pihole_cname_records.web_services.records :
    record.domain if can(regex("^(www|web|api|app)\\.", record.domain))
  ])
}
```

## Error Handling

Common errors and their meanings:

- **Authentication failed**: Pi-hole admin password is incorrect or API access is disabled
- **Connection timeout**: Pi-hole server is unreachable or overloaded
- **Client Error**: Network issues or Pi-hole API errors

The provider includes automatic retry logic with exponential backoff for transient network errors.

## Related Resources

- [`pihole_cname_record`](./cname_record.md) - For retrieving a specific CNAME record
- [`pihole_cname_record` resource](../resources/cname_record.md) - For managing CNAME records
- [`pihole_dns_records`](./dns_records.md) - For retrieving all DNS A records