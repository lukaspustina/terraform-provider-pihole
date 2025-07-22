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

### PUT DNS Record

```bash
 curl -v -X PUT --header "X-FTL-SID: XXX" --header "X-FTL-CSRF: XXX" "https://XXX:443/api/config/dns/hosts/192.168.0.22%20www.homelab.local" -H 'accept: application/json'
```

### Get CNAME Records

```bash
curl -v -X GET --header "X-FTL-SID: XXX" --header "X-FTL-CSRF: XXX" "https://XXX:443/api/config/dns/cnameRecords" -H 'accept: application/json'
```

### PUT CNAME Record

```bash
curl -v -X PUT --header "X-FTL-SID: XXX" --header "X-FTL-CSRF: XXX" "https://XXX:443/api/config/dns/cnameRecords/web,www" -H 'accept: application/json'
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
resource "pihole_dns_record" "www" {
  domain = "www.homelab.local"
  ip     = "192.168.0.22"
}
```

## Resource - CNAME Record

```hcl
resource "pihole_cname_record" "web" {
  domain = "web.homelab.local"
  target = "www.homelab.local"
}
```
