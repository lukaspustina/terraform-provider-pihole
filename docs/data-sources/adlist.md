# pihole_adlist

Retrieves a specific adlist from Pi-hole by its URL address.

## Example Usage

```terraform
data "pihole_adlist" "hagezi" {
  address = "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/adblock/pro.txt"
}

output "hagezi_id" {
  value = data.pihole_adlist.hagezi.id
}

output "hagezi_enabled" {
  value = data.pihole_adlist.hagezi.enabled
}
```

## Schema

### Required Arguments

- `address` (String) - URL of the adlist to look up.

### Read-Only Attributes

- `id` (String) - Pi-hole integer ID as a string.
- `enabled` (Boolean) - Whether the adlist is enabled.
- `comment` (String) - Description of the adlist.
- `type` (String) - `"block"` or `"allow"`.
- `groups` (List of Number) - Group IDs this adlist belongs to.
