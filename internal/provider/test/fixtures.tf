# Test fixture configurations for the Pi-hole Terraform provider

# Basic provider configuration for testing
provider "pihole" {
  url      = "https://pihole.test.local"
  password = "test-password-123"
}

# Test DNS records with various scenarios
resource "pihole_dns_record" "simple" {
  domain = "simple.test.local"
  ip     = "192.168.1.100"
}

resource "pihole_dns_record" "server" {
  domain = "server.homelab.test"
  ip     = "10.0.1.50"
}

resource "pihole_dns_record" "nas" {
  domain = "nas.storage.test"
  ip     = "10.0.1.60"
}

resource "pihole_dns_record" "docker_host" {
  domain = "docker.infra.test"
  ip     = "10.0.1.70"
}

# Test CNAME records with various scenarios  
resource "pihole_cname_record" "www" {
  domain = "www.test.local"
  target = "server.homelab.test"
}

resource "pihole_cname_record" "blog" {
  domain = "blog.test.local"
  target = "server.homelab.test"
}

resource "pihole_cname_record" "files" {
  domain = "files.test.local"
  target = "nas.storage.test"
}

resource "pihole_cname_record" "portainer" {
  domain = "portainer.test.local"
  target = "docker.infra.test"
}

resource "pihole_cname_record" "grafana" {
  domain = "grafana.monitoring.test"
  target = "docker.infra.test"
}

# Test complex domain names
resource "pihole_dns_record" "complex_subdomain" {
  domain = "multi-level.sub-domain.complex.test"
  ip     = "192.168.1.200"
}

resource "pihole_cname_record" "complex_cname" {
  domain = "service.api.v1.test"
  target = "multi-level.sub-domain.complex.test"
}

# Test IPv4 edge cases
resource "pihole_dns_record" "ipv4_edge_cases" {
  domain = "edge.test.local"
  ip     = "192.168.255.255"
}

resource "pihole_dns_record" "localhost_alias" {
  domain = "local.test.local"
  ip     = "127.0.0.1"
}

# Test chained CNAME records
resource "pihole_cname_record" "chain_level1" {
  domain = "app.test.local"
  target = "server.homelab.test"
}

resource "pihole_cname_record" "chain_level2" {
  domain = "service.test.local"
  target = "app.test.local"
}

resource "pihole_cname_record" "chain_level3" {
  domain = "frontend.test.local" 
  target = "service.test.local"
}