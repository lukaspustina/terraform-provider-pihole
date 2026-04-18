# pihole_adlists

Retrieves all adlists (block and allow lists) configured in Pi-hole.

## Example Usage

```terraform
data "pihole_adlists" "all" {}

output "adlist_count" {
  value = length(data.pihole_adlists.all.adlists)
}

locals {
  block_lists = [
    for al in data.pihole_adlists.all.adlists : al
    if al.type == "block"
  ]

  enabled_lists = [
    for al in data.pihole_adlists.all.adlists : al
    if al.enabled
  ]
}
```

## Schema

### Read-Only Attributes

- `id` (String) - Data source identifier, always `"adlists"`.
- `adlists` (List of Object) - All adlists configured in Pi-hole. Each object has:
  - `id` (String) - Pi-hole integer ID as a string
  - `address` (String) - URL of the adlist
  - `enabled` (Boolean) - Whether the adlist is enabled
  - `comment` (String) - Description of the adlist
  - `type` (String) - `"block"` or `"allow"`
  - `groups` (List of Number) - Group IDs this adlist belongs to
