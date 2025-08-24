---
page_title: "technitium_dns_record Resource"
subcategory: "DNS Records"
description: |-
  Manages DNS records in Technitium DNS Server.
---

# technitium_dns_record Resource

Manages DNS records in Technitium DNS Server. This resource allows you to create, update, and delete various types of DNS records including A, AAAA, CNAME, MX, TXT, PTR, NS, and SRV records.

## Example Usage

### A Record (IPv4)

```terraform
resource "technitium_dns_record" "example_a" {
  zone = "example.com"
  name = "www"
  type = "A"
  ttl  = 300
  data = "192.168.1.100"
}
```

### AAAA Record (IPv6)

```terraform
resource "technitium_dns_record" "example_aaaa" {
  zone = "example.com"
  name = "www"
  type = "AAAA"
  ttl  = 300
  data = "2001:db8::1"
}
```

### CNAME Record

```terraform
resource "technitium_dns_record" "example_cname" {
  zone = "example.com"
  name = "blog"
  type = "CNAME"
  ttl  = 300
  data = "www.example.com"
}
```

### MX Record

```terraform
resource "technitium_dns_record" "example_mx" {
  zone     = "example.com"
  name     = "@"  # Root domain
  type     = "MX"
  ttl      = 300
  data     = "mail.example.com"
  priority = 10
}
```

### TXT Record

```terraform
resource "technitium_dns_record" "example_txt_spf" {
  zone = "example.com"
  name = "@"
  type = "TXT"
  ttl  = 300
  data = "v=spf1 include:_spf.google.com ~all"
}
```

### NS Record

```terraform
resource "technitium_dns_record" "example_ns" {
  zone = "example.com"
  name = "subdomain"
  type = "NS"
  ttl  = 86400
  data = "ns1.subdomain.example.com"
}
```

### SRV Record

```terraform
resource "technitium_dns_record" "example_srv" {
  zone     = "example.com"
  name     = "_sip._tcp"
  type     = "SRV"
  ttl      = 300
  data     = "sip.example.com"
  priority = 10
  weight   = 5
  port     = 5060
}
```

### PTR Record (Reverse DNS)

```terraform
resource "technitium_dns_record" "example_ptr" {
  zone = "1.168.192.in-addr.arpa"
  name = "100"
  type = "PTR"
  ttl  = 300
  data = "www.example.com"
}
```

## Argument Reference

The following arguments are supported:

* `zone` - (Required) The zone in which to create the DNS record.

* `name` - (Required) The record name (e.g., 'www' for <www.example.com>). Use '@' for the zone's root domain.

* `type` - (Required) The DNS record type. Valid values are: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `PTR`, `NS`, `SRV`.

* `ttl` - (Required) Time-to-live value in seconds.

* `data` - (Required) Record data, which varies by record type:
  * For A records: The IPv4 address.
  * For AAAA records: The IPv6 address.
  * For CNAME records: The canonical name (domain).
  * For MX records: The mail exchange server domain name.
  * For TXT records: The text content.
  * For PTR records: The pointer domain name.
  * For NS records: The name server domain.
  * For SRV records: The target domain for the service.

* `priority` - (Optional) Priority value for MX and SRV records. Defaults to 0.

* `weight` - (Optional) Weight value for SRV records. Defaults to 0.

* `port` - (Optional) Port value for SRV records. Defaults to 0.

* `comments` - (Optional) Optional comments for the DNS record.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A unique identifier for the record in the format `zone:name:type[:priority][:data]`.

* `disabled` - Whether the record is disabled.

* `dnssec_status` - DNSSEC status of the record.

* `last_used_on` - When the record was last used.

## Import

DNS records can be imported using the format `zone:name:type[:priority][:data]`, for example:

```
terraform import technitium_dns_record.a_record example.com:www:A:192.0.2.10
```

For MX records with priority:

```
terraform import technitium_dns_record.mx_record example.com:@:MX:10:mail.example.com
```
