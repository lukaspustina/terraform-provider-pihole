# pihole_config (Data Source)

Reads a Pi-hole configuration setting. This data source can be used to read current configuration values, such as checking if `webserver.api.app_sudo` is enabled.

## Example Usage

```terraform
data "pihole_config" "app_sudo_status" {
  key = "webserver.api.app_sudo"
}

output "app_sudo_enabled" {
  value = data.pihole_config.app_sudo_status.value
}
```

## Schema

### Required Arguments

- `key` (String) - Configuration key to read using dot notation (e.g., `webserver.api.app_sudo`).

### Read-Only Attributes

- `value` (String) - Current configuration value. Boolean values are returned as `"true"` or `"false"`.
- `id` (String) - Data source identifier (same as key).

## Related Resources

- [`pihole_config` resource](../resources/config.md) - For managing configuration values
