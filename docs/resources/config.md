# pihole_config

Manages a Pi-hole configuration setting. This resource allows you to create, read, update, and reset configuration values using dot-notation keys.

**Important**: Configuration changes may require an admin password. Application passwords cannot modify Pi-hole configuration settings unless `webserver.api.app_sudo` is enabled. This setting can be enabled via the Pi-hole web interface under Settings > API/Web interface > "Permit destructive actions via API".

## Example Usage

### Enable Application Password Sudo

```terraform
resource "pihole_config" "enable_app_sudo" {
  key   = "webserver.api.app_sudo"
  value = "true"
}
```

## Schema

### Required Arguments

- `key` (String) - Configuration key using dot notation (e.g., `webserver.api.app_sudo`). Changing this forces a new resource.
- `value` (String) - Configuration value. For boolean settings, use `"true"` or `"false"`.

### Read-Only Attributes

- `id` (String) - The resource identifier (same as key).

## Import

Configuration settings can be imported using the configuration key:

```shell
terraform import pihole_config.example webserver.api.app_sudo
```

## Behavior Notes

- **Delete behavior**: Deleting this resource resets the configuration to its default value (e.g., `false` for `webserver.api.app_sudo`) rather than removing the setting.
- **Boolean conversion**: String values `"true"` and `"false"` are automatically converted to boolean types for the Pi-hole API.
- **Supported namespaces**: Currently only `webserver.*` configuration keys are supported.

## Related Resources

- [`pihole_config` data source](../data-sources/config.md) - For reading configuration values without managing them
