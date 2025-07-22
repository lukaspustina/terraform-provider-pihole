# pihole_dns_record

Manages a DNS A record in Pi-hole. This resource allows you to create, read, update, and delete DNS A records that resolve domain names to IP addresses.

## Example Usage

### Basic Usage

```terraform
resource "pihole_dns_record" "server" {
  domain = "server.homelab.local"
  ip     = "192.168.1.100"
}
```

### Multiple DNS Records

```terraform
resource "pihole_dns_record" "nas" {
  domain = "nas.homelab.local"
  ip     = "192.168.1.101"
}

resource "pihole_dns_record" "docker" {
  domain = "docker.homelab.local"
  ip     = "192.168.1.102"
}

resource "pihole_dns_record" "router" {
  domain = "router.homelab.local"
  ip     = "192.168.1.1"
}
```

### Using Variables

```terraform
variable "servers" {
  type = map(string)
  default = {
    "nas.homelab.local"    = "192.168.1.101"
    "docker.homelab.local" = "192.168.1.102"
    "router.homelab.local" = "192.168.1.1"
  }
}

resource "pihole_dns_record" "servers" {
  for_each = var.servers
  
  domain = each.key
  ip     = each.value
}
```

## Schema

### Required Arguments

- `domain` (String) - The fully qualified domain name to resolve. Must be a valid domain name format.
- `ip` (String) - The IPv4 address that the domain should resolve to. Must be a valid IPv4 address format.

### Read-Only Attributes

- `id` (String) - The resource identifier. This is set to the domain name for uniqueness.

## Import

DNS records can be imported using the domain name:

```shell
terraform import pihole_dns_record.example server.homelab.local
```

## Validation

The resource performs validation on both arguments:

- **Domain**: Must be a valid fully qualified domain name (FQDN) format
- **IP**: Must be a valid IPv4 address format (e.g., `192.168.1.100`)

## Behavior Notes

- **Uniqueness**: Each domain can only have one DNS A record. If you attempt to create multiple records for the same domain, the last one will overwrite previous ones.
- **Case Sensitivity**: Domain names are case-insensitive and will be stored in Pi-hole as entered.
- **Updates**: Changing either the domain or IP will result in the old record being deleted and a new one created.
- **IPv6**: Currently only IPv4 addresses are supported. IPv6 support may be added in future versions.

## Error Handling

Common errors and their meanings:

- **Invalid domain format**: The domain name doesn't match FQDN requirements
- **Invalid IP format**: The IP address is not a valid IPv4 address
- **Authentication failed**: Pi-hole admin password is incorrect or API access is disabled
- **Connection timeout**: Pi-hole server is unreachable or overloaded

The provider includes automatic retry logic with exponential backoff for transient network errors.

## Related Resources

- [`pihole_cname_record`](./cname_record.md) - For creating CNAME aliases that point to DNS A records