# pihole_cname_record

Manages a CNAME record in Pi-hole. This resource allows you to create, read, update, and delete CNAME records that create aliases pointing to other domain names.

## Example Usage

### Basic Usage

```terraform
# First create the target DNS record
resource "pihole_dns_record" "server" {
  domain = "server.homelab.local"
  ip     = "192.168.1.100"
}

# Then create a CNAME alias pointing to it
resource "pihole_cname_record" "www" {
  domain = "www.homelab.local"
  target = "server.homelab.local"
}
```

### Multiple CNAME Records

```terraform
resource "pihole_dns_record" "server" {
  domain = "server.homelab.local"
  ip     = "192.168.1.100"
}

resource "pihole_dns_record" "nas" {
  domain = "nas.homelab.local"
  ip     = "192.168.1.101"
}

# Multiple aliases for the same server
resource "pihole_cname_record" "www" {
  domain = "www.homelab.local"
  target = "server.homelab.local"
}

resource "pihole_cname_record" "blog" {
  domain = "blog.homelab.local"
  target = "server.homelab.local"
}

resource "pihole_cname_record" "api" {
  domain = "api.homelab.local"
  target = "server.homelab.local"
}

# Alias for NAS
resource "pihole_cname_record" "files" {
  domain = "files.homelab.local"
  target = "nas.homelab.local"
}
```

### Using Variables and Dependencies

```terraform
variable "services" {
  type = map(object({
    ip      = string
    aliases = list(string)
  }))
  default = {
    server = {
      ip      = "192.168.1.100"
      aliases = ["www", "blog", "api"]
    }
    nas = {
      ip      = "192.168.1.101"
      aliases = ["files", "media"]
    }
  }
}

# Create DNS A records
resource "pihole_dns_record" "services" {
  for_each = var.services
  
  domain = "${each.key}.homelab.local"
  ip     = each.value.ip
}

# Create CNAME aliases
resource "pihole_cname_record" "aliases" {
  for_each = {
    for service, config in var.services :
    service => config.aliases
  }
  
  count = length(each.value)
  
  domain = "${each.value[count.index]}.homelab.local"
  target = "${each.key}.homelab.local"
  
  depends_on = [pihole_dns_record.services]
}
```

### External Target

```terraform
# CNAME pointing to external domain
resource "pihole_cname_record" "external_alias" {
  domain = "external.homelab.local"
  target = "example.com"
}
```

## Schema

### Required Arguments

- `domain` (String) - The fully qualified domain name for the CNAME alias. Must be a valid domain name format.
- `target` (String) - The target domain name that this CNAME should point to. Must be a valid domain name format.

### Read-Only Attributes

- `id` (String) - The resource identifier. This is set to the domain name for uniqueness.

## Import

CNAME records can be imported using the domain name:

```shell
terraform import pihole_cname_record.example www.homelab.local
```

## Validation

The resource performs validation on both arguments:

- **Domain**: Must be a valid fully qualified domain name (FQDN) format
- **Target**: Must be a valid fully qualified domain name (FQDN) format

## Behavior Notes

- **Uniqueness**: Each domain can only have one CNAME record. You cannot create multiple CNAME records for the same domain.
- **Circular References**: Pi-hole will prevent circular CNAME references (e.g., A pointing to B, B pointing to A).
- **Mixed Records**: A domain cannot have both a DNS A record and a CNAME record. They are mutually exclusive.
- **Case Sensitivity**: Domain names are case-insensitive and will be stored in Pi-hole as entered.
- **Updates**: Changing either the domain or target will result in the old record being deleted and a new one created.
- **Target Resolution**: The target domain does not need to be managed by this provider - it can point to external domains or existing Pi-hole records.

## Dependencies

When creating CNAME records that point to DNS A records managed by the same Terraform configuration, use explicit dependencies:

```terraform
resource "pihole_dns_record" "server" {
  domain = "server.homelab.local"  
  ip     = "192.168.1.100"
}

resource "pihole_cname_record" "www" {
  domain = "www.homelab.local"
  target = "server.homelab.local"
  
  depends_on = [pihole_dns_record.server]
}
```

## Error Handling

Common errors and their meanings:

- **Invalid domain format**: The domain or target name doesn't match FQDN requirements
- **Circular reference**: The CNAME would create a circular reference chain
- **Conflicting record**: Attempting to create a CNAME for a domain that already has an A record
- **Authentication failed**: Pi-hole admin password is incorrect or API access is disabled
- **Connection timeout**: Pi-hole server is unreachable or overloaded

The provider includes automatic retry logic with exponential backoff for transient network errors.

## Related Resources

- [`pihole_dns_record`](./dns_record.md) - For creating DNS A records that CNAME records can point to