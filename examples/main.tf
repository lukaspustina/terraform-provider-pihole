# Pi-hole Terraform Provider Examples

terraform {
  required_providers {
    pihole = {
      source  = "registry.terraform.io/lukaspustina/pihole"
      version = "~> 0.1"
    }
  }
}

# Basic provider configuration
provider "pihole" {
  url      = "https://pihole.homelab.local:443"
  password = var.pihole_password
}

# Variables
variable "pihole_password" {
  description = "Pi-hole admin password"
  type        = string
  sensitive   = true
}

variable "base_domain" {
  description = "Base domain for homelab services"
  type        = string
  default     = "homelab.local"
}

# Local values for IP addresses
locals {
  server_ips = {
    docker     = "192.168.1.10"
    nas        = "192.168.1.20"
    router     = "192.168.1.1"
    pihole     = "192.168.1.5"
    k8s_master = "192.168.1.30"
    k8s_node1  = "192.168.1.31"
    k8s_node2  = "192.168.1.32"
  }
}

# Main server DNS records
resource "pihole_dns_record" "docker_host" {
  domain = "docker.${var.base_domain}"
  ip     = local.server_ips.docker
}

resource "pihole_dns_record" "nas" {
  domain = "nas.${var.base_domain}"
  ip     = local.server_ips.nas
}

resource "pihole_dns_record" "router" {
  domain = "router.${var.base_domain}"
  ip     = local.server_ips.router
}

resource "pihole_dns_record" "pihole" {
  domain = "pihole.${var.base_domain}"
  ip     = local.server_ips.pihole
}

# Kubernetes cluster
resource "pihole_dns_record" "k8s_master" {
  domain = "k8s-master.${var.base_domain}"
  ip     = local.server_ips.k8s_master
}

resource "pihole_dns_record" "k8s_node1" {
  domain = "k8s-node1.${var.base_domain}"
  ip     = local.server_ips.k8s_node1
}

resource "pihole_dns_record" "k8s_node2" {
  domain = "k8s-node2.${var.base_domain}"
  ip     = local.server_ips.k8s_node2
}

# Service aliases using CNAME records
resource "pihole_cname_record" "portainer" {
  domain = "portainer.${var.base_domain}"
  target = "docker.${var.base_domain}"
}

resource "pihole_cname_record" "grafana" {
  domain = "grafana.${var.base_domain}"
  target = "docker.${var.base_domain}"
}

resource "pihole_cname_record" "prometheus" {
  domain = "prometheus.${var.base_domain}"
  target = "docker.${var.base_domain}"
}

resource "pihole_cname_record" "nextcloud" {
  domain = "cloud.${var.base_domain}"
  target = "nas.${var.base_domain}"
}

resource "pihole_cname_record" "files" {
  domain = "files.${var.base_domain}"
  target = "nas.${var.base_domain}"
}

resource "pihole_cname_record" "media" {
  domain = "media.${var.base_domain}"
  target = "nas.${var.base_domain}"
}

# Kubernetes service aliases
resource "pihole_cname_record" "k8s" {
  domain = "k8s.${var.base_domain}"
  target = "k8s-master.${var.base_domain}"
}

resource "pihole_cname_record" "kubernetes" {
  domain = "kubernetes.${var.base_domain}"
  target = "k8s-master.${var.base_domain}"
}

# Development aliases
resource "pihole_cname_record" "dev" {
  domain = "dev.${var.base_domain}"
  target = "docker.${var.base_domain}"
}

resource "pihole_cname_record" "test" {
  domain = "test.${var.base_domain}"
  target = "docker.${var.base_domain}"
}

resource "pihole_cname_record" "staging" {
  domain = "staging.${var.base_domain}"
  target = "docker.${var.base_domain}"
}

# Web services
resource "pihole_cname_record" "www" {
  domain = "www.${var.base_domain}"
  target = "docker.${var.base_domain}"
}

resource "pihole_cname_record" "blog" {
  domain = "blog.${var.base_domain}"
  target = "docker.${var.base_domain}"
}

resource "pihole_cname_record" "wiki" {
  domain = "wiki.${var.base_domain}"
  target = "docker.${var.base_domain}"
}

# Outputs
output "dns_records" {
  description = "DNS A records created"
  value = {
    docker     = pihole_dns_record.docker_host
    nas        = pihole_dns_record.nas
    router     = pihole_dns_record.router
    pihole     = pihole_dns_record.pihole
    k8s_master = pihole_dns_record.k8s_master
    k8s_node1  = pihole_dns_record.k8s_node1
    k8s_node2  = pihole_dns_record.k8s_node2
  }
}

output "cname_records" {
  description = "CNAME records created"
  value = {
    services = [
      pihole_cname_record.portainer,
      pihole_cname_record.grafana,
      pihole_cname_record.prometheus,
    ]
    storage = [
      pihole_cname_record.nextcloud,
      pihole_cname_record.files,
      pihole_cname_record.media,
    ]
    web = [
      pihole_cname_record.www,
      pihole_cname_record.blog,
      pihole_cname_record.wiki,
    ]
  }
}