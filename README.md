# Terraform Provider for Pi-hole

A Terraform provider for managing Pi-hole DNS records and CNAME records using the Pi-hole API v6.

This provider allows you to manage your Pi-hole DNS configuration as code, enabling Infrastructure as Code (IaC) practices for your local DNS setup.

This work is based on the preliminary work of [Ryan Wholey](https://github.com/ryanwholey/terraform-provider-pihole) and re-uses the HCL resource definitions to allow for easy migration, but does not share any code etc. It has been improved to support the latest Pi-hole API v6, including robust error handling, rate limiting, and TLS support.

## Features

- ✅ **DNS A Records**: Manage custom DNS A records
- ✅ **CNAME Records**: Manage CNAME aliases
- ✅ **Configuration Management**: Manage Pi-hole configuration settings (requires admin password)
- ✅ **Pi-hole API v6**: Full compatibility with modern Pi-hole installations
- ✅ **Robust Error Handling**: Automatic retries with exponential backoff
- ✅ **Rate Limited**: Prevents API overload with built-in delays
- ✅ **TLS Support**: Secure TLS verification by default, with optional bypass for self-signed certificates

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- Pi-hole installation with API v6 support
- Pi-hole admin password

## Usage

### Provider Configuration

#### Basic Configuration

```hcl
provider "pihole" {
  url      = "https://pihole.homelab.local:443"
  password = "your-admin-password-here"
}
```

#### Advanced Configuration

```hcl
provider "pihole" {
  url      = "https://pihole.homelab.local:443"
  password = var.pihole_password
  
  # TLS configuration (optional)
  insecure_tls = false  # Skip TLS verification for self-signed certificates
  
  # Connection management (optional)
  max_connections        = 1       # Limit concurrent connections
  request_delay_ms       = 500     # Slower requests for busy Pi-hole
  retry_attempts         = 5       # More retries for unstable connections
  retry_backoff_base_ms  = 1000    # Longer backoff delays
}
```

#### Configuration Options

- `url` (Required) - Pi-hole server URL
- `password` (Required) - Pi-hole admin password
- `insecure_tls` (Optional) - Skip TLS certificate verification (default: false)
- `max_connections` (Optional) - Maximum concurrent connections (default: 1)
- `request_delay_ms` (Optional) - Delay between requests in milliseconds (default: 300)
- `retry_attempts` (Optional) - Number of retry attempts (default: 3)  
- `retry_backoff_base_ms` (Optional) - Base retry delay in milliseconds (default: 500)

### Full Configuration Example

```hcl
terraform {
  required_providers {
    pihole = {
      source  = "registry.terraform.io/lukaspustina/pihole"
      version = "0.2.0"
    }
  }
}

provider "pihole" {
  url      = "https://pihole.homelab.local:443"  # Your Pi-hole URL
  password = "your-admin-password-here"        # Pi-hole admin password
  
  # Optional: TLS configuration (shown with defaults)
  insecure_tls = false             # Skip TLS certificate verification (default: false)
  
  # Optional: Connection and timing settings (shown with defaults)
  max_connections        = 1       # Maximum concurrent connections to Pi-hole
  request_delay_ms       = 300     # Delay between API requests in milliseconds
  retry_attempts         = 3       # Number of retry attempts for failed requests
  retry_backoff_base_ms  = 500     # Base delay for retry backoff in milliseconds
}
```

### DNS A Records

Create custom DNS A records that resolve domain names to IP addresses:

```hcl
resource "pihole_dns_record" "homelab_server" {
  domain = "server.homelab.local"
  ip     = "192.168.1.100"
}

resource "pihole_dns_record" "nas" {
  domain = "nas.homelab.local"  
  ip     = "192.168.1.101"
}

resource "pihole_dns_record" "docker_host" {
  domain = "docker.homelab.local"
  ip     = "192.168.1.102"
}
```

### CNAME Records

Create CNAME aliases that point to other domain names:

```hcl
resource "pihole_cname_record" "www" {
  domain = "www.homelab.local"
  target = "server.homelab.local"
}

resource "pihole_cname_record" "blog" {
  domain = "blog.homelab.local"
  target = "server.homelab.local"
}

resource "pihole_cname_record" "files" {
  domain = "files.homelab.local"
  target = "nas.homelab.local"
}
```

### Configuration Management

Manage Pi-hole webserver configuration settings such as enabling API access for application passwords:

**⚠️ Important**: Webserver configuration management requires the **admin password**, not an application password. Application passwords cannot modify Pi-hole webserver configuration settings unless `webserver.api.app_sudo` is enabled. This setting can be enabled via the Pi-hole web interface under Settings → API/Web interface → "Permit destructive actions via API".

```hcl
# Enable app_sudo to allow application passwords to modify all Pi-hole settings
resource "pihole_config" "enable_app_sudo" {
  key   = "webserver.api.app_sudo"
  value = "true"
}

# Read current webserver configuration status
data "pihole_config" "app_sudo_status" {
  key = "webserver.api.app_sudo"
}

# Use configuration status in outputs
output "api_app_sudo_enabled" {
  description = "Whether application passwords can modify any Pi-hole settings"
  value       = data.pihole_config.app_sudo_status.value == "true"
}
```

### Complete Example

```hcl
terraform {
  required_providers {
    pihole = {
      source  = "registry.terraform.io/lukaspustina/pihole"
      version = "0.2.0"
    }
  }
}

provider "pihole" {
  url      = "https://pihole.homelab.local:443"
  password = var.pihole_password
}

# Main services
resource "pihole_dns_record" "docker" {
  domain = "docker.homelab.local"
  ip     = "192.168.0.19"
}

resource "pihole_dns_record" "nas" {
  domain = "nas.homelab.local"
  ip     = "192.168.0.20"
}

# Service aliases
resource "pihole_cname_record" "portainer" {
  domain = "portainer.homelab.local"
  target = "docker.homelab.local"
}

resource "pihole_cname_record" "files" {
  domain = "files.homelab.local"
  target = "nas.homelab.local"
}
```

## Resource Reference

### pihole_dns_record

Manages a DNS A record in Pi-hole.

#### Arguments

- `domain` (Required, String) - The domain name to resolve
- `ip` (Required, String) - The IP address to resolve to

#### Attributes

- `id` (String) - The resource identifier (same as domain)

### pihole_cname_record  

Manages a CNAME record in Pi-hole.

#### Arguments

- `domain` (Required, String) - The CNAME alias domain name
- `target` (Required, String) - The target domain name to point to

#### Attributes

- `id` (String) - The resource identifier (same as domain)

## Development

### Building

```bash
go build -o terraform-provider-pihole
```

### Testing

#### Unit Tests

Run unit tests (no Pi-hole instance required):

```bash
go test -v ./internal/provider
```

#### Acceptance Tests

Acceptance tests require a running Pi-hole instance. They can be run in two ways:

##### Option 1: Local Pi-hole Instance

If you have a Pi-hole running locally:

```bash
# Set environment variables for your Pi-hole instance
export PIHOLE_URL="http://your-pihole-ip:80"  # or https://your-pihole:443
export PIHOLE_PASSWORD="your-admin-password"

# Run acceptance tests
TF_ACC=1 go test -v ./internal/provider -run TestAcc -timeout 30m
```

##### Option 2: Docker Pi-hole (Recommended)

Use Docker to run a temporary Pi-hole for testing:

```bash
# Start a Pi-hole container for testing
docker run -d \
  --name pihole-test \
  -p 8080:80 \
  -p 8053:53/tcp \
  -p 8053:53/udp \
  -e WEBPASSWORD=testpass123 \
  -e VIRTUAL_HOST=localhost \
  pihole/pihole:latest

# Wait for Pi-hole to be ready (may take 30-60 seconds)
timeout 60 bash -c 'until curl -f http://localhost:8080/admin/; do sleep 2; done'

# Set environment variables
export PIHOLE_URL="http://localhost:8080"
export PIHOLE_PASSWORD="testpass123"

# Run acceptance tests
TF_ACC=1 go test -v ./internal/provider -run TestAcc -timeout 30m

# Clean up when done
docker stop pihole-test && docker rm pihole-test
```

##### Test Behavior

- **Without PIHOLE_URL**: Acceptance tests automatically skip, only unit tests run
- **With PIHOLE_URL**: Tests connect to the specified Pi-hole instance and create/modify real DNS records
- **Test cleanup**: Tests automatically clean up resources they create
- **Timeout**: Use `-timeout 30m` as acceptance tests may take time due to rate limiting

##### Running Specific Test Suites

```bash
# Run only DNS record tests
TF_ACC=1 go test -v ./internal/provider -run TestAccPiholeDNSRecord -timeout 30m

# Run only CNAME record tests  
TF_ACC=1 go test -v ./internal/provider -run TestAccPiholeCNAMERecord -timeout 30m

# Run only provider configuration tests (no Pi-hole required)
TF_ACC=1 go test -v ./internal/provider -run TestAccPiholeProvider -timeout 5m
```

### Local Development

```bash
# Setup development environment (includes automatic linting on commits)
make setup-dev

# Build and install locally (automatically runs linting)
make install

# Manual formatting and checks
make fmt        # Format code with gofmt and goimports  
make lint       # Run golangci-lint
make vet        # Run go vet
make check      # Run all quality checks (fmt + vet + lint + test-unit)
make check-full # Run comprehensive checks including coverage
```

#### Automatic Linting

The project is configured for automatic linting:

- **On build**: `make build` automatically runs formatting, vet, and linting
- **On commit**: Git pre-commit hooks automatically format and lint code
- **In CI**: GitHub Actions runs comprehensive linting and formatting checks

To enable automatic linting on commits:

```bash
make setup-dev
```

This configures git to run formatting and linting before every commit.

### CI/CD Pipeline

The project includes a comprehensive GitHub Actions pipeline that:

1. **Builds and Tests**: Runs on multiple Go versions (1.22, 1.23) with cross-platform builds
2. **Acceptance Tests**: Automatically sets up a Pi-hole Docker container and runs full acceptance tests
3. **Security Scanning**: Runs vulnerability scans with Trivy and Nancy
4. **Code Quality**: Enforces formatting, linting, and static analysis
5. **Automated Releases**: Uses GoReleaser for cross-platform binary releases

The acceptance tests in CI use the same Docker setup as described in the testing section above.

### Project Structure

```
├── main.go                                    # Provider entry point
├── internal/provider/
│   ├── provider.go                           # Main provider configuration
│   ├── client.go                            # Pi-hole API client
│   ├── dns_record_resource.go               # DNS A record resource
│   └── cname_record_resource.go             # CNAME record resource
├── go.mod                                   # Go module definition
├── Makefile                                 # Build automation
├── SPECS.md                                 # Original project specifications
└── CLAUDE.md                               # Development guidance
```

## API Compatibility

This provider is designed for **Pi-hole API v6** and uses the following endpoints:

- `POST /api/auth` - Authentication
- `GET /api/config/dns/hosts` - Retrieve DNS records  
- `PUT /api/config/dns/hosts/{ip}%20{domain}` - Create/update DNS records
- `DELETE /api/config/dns/hosts/{ip}%20{domain}` - Delete DNS records
- `GET /api/config/dns/cnameRecords` - Retrieve CNAME records
- `PUT /api/config/dns/cnameRecords/{domain},{target}` - Create/update CNAME records  
- `DELETE /api/config/dns/cnameRecords/{domain},{target}` - Delete CNAME records
- `GET /api/config/webserver` - Retrieve webserver configuration settings
- `PUT /api/config/webserver` - Update webserver configuration settings

## Troubleshooting

### Connection Issues

If you experience connection timeouts or "connection refused" errors:

- The provider includes automatic retry logic with exponential backoff
- API calls are rate-limited to prevent overwhelming Pi-hole
- Only one concurrent connection is maintained

### Authentication Issues

- Ensure your Pi-hole admin password is correct
- Check that your Pi-hole URL is accessible and uses the correct protocol (HTTP/HTTPS)
- Verify that API access is enabled in Pi-hole settings

### Application Password vs Admin Password

If you're using a Pi-hole **application password** (not the main admin password), **NO modifications are possible** unless `webserver.api.app_sudo` is enabled:

- **DNS/CNAME records**: Application passwords **cannot** modify DNS or CNAME records unless `webserver.api.app_sudo` is enabled
- **Webserver configuration changes**: Application passwords **cannot** modify Pi-hole webserver configuration settings unless `webserver.api.app_sudo` is enabled
- **Enable all modifications**: Set `webserver.api.app_sudo = true` using an admin password first:

```hcl
# This requires admin password to set initially
resource "pihole_config" "enable_app_sudo" {
  key   = "webserver.api.app_sudo"  
  value = "true"
}
```

- **Alternative**: Enable "Permit destructive actions via API" in Pi-hole web interface (Settings → API/Web interface)

### TLS Certificate Issues

By default, the provider verifies TLS certificates for secure connections. If your Pi-hole uses self-signed certificates, you can disable TLS verification by setting `insecure_tls = true` in your provider configuration:

```hcl
provider "pihole" {
  url          = "https://pihole.homelab.local:443"
  password     = "your-admin-password-here"
  insecure_tls = true  # Allow self-signed certificates
}
```

**Security Note**: Only use `insecure_tls = true` for local Pi-hole installations with self-signed certificates. For production environments or public-facing Pi-hole instances, keep the default `insecure_tls = false` setting.

## Contributing

This provider was created as a collaborative effort between a developer and Claude AI. Contributions are welcome!

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

This Terraform provider was developed with the assistance of [Claude AI](https://claude.ai/code) from Anthropic. The collaborative development process involved:

---

**Note**: This provider is not officially affiliated with Pi-hole. Pi-hole is a registered trademark of Pi-hole LLC.
