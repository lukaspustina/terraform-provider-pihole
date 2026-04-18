# pihole_adlist

Manages an adlist (block or allow list) in Pi-hole. Adlists are URLs that Pi-hole fetches to build its DNS blocklist or allowlist.

## Example Usage

### Block List

```terraform
resource "pihole_adlist" "hagezi_pro" {
  address = "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/adblock/pro.txt"
  enabled = true
  comment = "HaGeZi Pro blocklist"
  type    = "block"
  groups  = [0]
}
```

### Allow List

```terraform
resource "pihole_adlist" "my_allowlist" {
  address = "https://example.com/allowlist.txt"
  enabled = true
  comment = "Custom allow list"
  type    = "allow"
}
```

### Managing Multiple Lists

```terraform
locals {
  blocklists = {
    "hagezi_pro"   = "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/adblock/pro.txt"
    "stevenblack"  = "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts"
    "oisd_big"     = "https://big.oisd.nl/domainswild"
  }
}

resource "pihole_adlist" "blocklists" {
  for_each = local.blocklists

  address = each.value
  comment = each.key
  type    = "block"
  enabled = true
}
```

## Schema

### Required Arguments

- `address` (String) - URL of the adlist. Must be a non-empty string.

### Optional Arguments

- `enabled` (Boolean) - Whether the adlist is active. Defaults to `true`.
- `comment` (String) - Description of the adlist. Defaults to `""`.
- `type` (String) - Type of list: `"block"` or `"allow"`. Defaults to `"block"`.
- `groups` (List of Number) - Group IDs this adlist belongs to. Defaults to `[0]` (the default group).

### Read-Only Attributes

- `id` (String) - The integer ID assigned by Pi-hole, stored as a string.

## Import

Adlists can be imported using the integer ID assigned by Pi-hole:

```shell
terraform import pihole_adlist.example 1
```

The ID can be found in the Pi-hole web interface under Lists, or via the API at `/api/lists`.

## Behavior Notes

- **Gravity**: Adding or removing adlists does not automatically trigger a gravity update. Run gravity manually or via the Pi-hole web interface to apply list changes to the DNS blocklist.
- **Groups**: An adlist must belong to at least one group to be active. The default group has ID `0`.
- **Updates**: All attributes except `id` can be updated in-place without destroying the resource.
- **Duplicate addresses**: Pi-hole does not allow two adlists with the same address. Attempting to create a duplicate will result in an API error.
