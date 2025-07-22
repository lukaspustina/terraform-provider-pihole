# Variables for Pi-hole Terraform Provider Examples

variable "pihole_url" {
  description = "Pi-hole server URL (including protocol and port)"
  type        = string
  default     = "https://pihole.homelab.local:443"
  
  validation {
    condition     = can(regex("^https?://", var.pihole_url))
    error_message = "Pi-hole URL must start with http:// or https://"
  }
}

variable "pihole_password" {
  description = "Pi-hole admin password"
  type        = string
  sensitive   = true
  
  validation {
    condition     = length(var.pihole_password) > 0
    error_message = "Pi-hole password cannot be empty"
  }
}

variable "base_domain" {
  description = "Base domain for homelab services"
  type        = string
  default     = "homelab.local"
  
  validation {
    condition     = can(regex("^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$", var.base_domain))
    error_message = "Base domain must be a valid domain name"
  }
}

variable "network_prefix" {
  description = "Network prefix for IP addresses (e.g., '192.168.1')"
  type        = string
  default     = "192.168.1"
  
  validation {
    condition     = can(regex("^([0-9]{1,3}\\.){2}[0-9]{1,3}$", var.network_prefix))
    error_message = "Network prefix must be in format 'x.y.z' (first 3 octets of IP)"
  }
}

variable "server_hosts" {
  description = "Map of server hostnames to their last IP octet"
  type        = map(number)
  default = {
    docker     = 10
    nas        = 20
    router     = 1
    pihole     = 5
    k8s_master = 30
    k8s_node1  = 31
    k8s_node2  = 32
  }
  
  validation {
    condition = alltrue([
      for host, octet in var.server_hosts : octet >= 1 && octet <= 254
    ])
    error_message = "IP octets must be between 1 and 254"
  }
}

variable "service_aliases" {
  description = "Map of service aliases to their target hosts"
  type        = map(string)
  default = {
    portainer  = "docker"
    grafana    = "docker"
    prometheus = "docker"
    nextcloud  = "nas"
    files      = "nas"
    media      = "nas"
    www        = "docker"
    blog       = "docker"
    wiki       = "docker"
  }
}

variable "enable_development_services" {
  description = "Enable development service aliases (dev, test, staging)"
  type        = bool
  default     = true
}

variable "enable_kubernetes_cluster" {
  description = "Enable Kubernetes cluster DNS records"
  type        = bool
  default     = false
}