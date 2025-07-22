# SPECS for Pi Hole Terraform Provider

I like to create a Terraform provider for Pi Hole to manage DNS records and CNAME records using the Pi Hole API version 6.

## Pi Hole API Documentation

* <https://ftl.pi-hole.net/master/docs/>

## Example API Calls

### Authentication

```bash
curl -X POST "https://:443/api/auth" -H 'accept: application/json' -H 'content-type: application/json'  -d '{"password":"XXX"}'
```

### Get DNS Records

```bash
curl -v -X GET --header "X-FTL-SID: XXX" --header "X-FTL-CSRF: XXX" "https://XXX:443/api/config/dns/hosts" -H 'accept: application/json'
```

### Get CNAME Records

```bash
curl -v -X GET --header "X-FTL-SID: XXX" --header "X-FTL-CSRF: XXX" "https://XXX:443/api/config/dns/cnameRecords" -H 'accept: application/json'
```

## Provider

```hcl
provider "pihole" {
  url = "https://example.com"
  password = "xxx"
}
```

## Resource - DNS Record

```hcl
resource "pihole_dns_record" "sae" {
  domain = "sae.${local.domain}"
  ip     = "192.168.2.159"
}
```

## Resource - CNAME Record

```hcl
resource "pihole_cname_record" "start" {
  domain = "start.services.${local.domain}"
  target = "docker-priv.${local.domain}"
}
```
