---
page_title: "technitium_zone Data Source"
subcategory: "Zone Management"
description: |-
  Retrieves information about a DNS zone in Technitium DNS Server.
---

# technitium_zone Data Source

Retrieves information about a DNS zone in Technitium DNS Server. This data source allows you to fetch details about an existing zone including its type, settings, and current state.

## Example Usage

```terraform
# Data source to read an existing DNS zone
data "technitium_zone" "example" {
  name = "example.com"
}

# Output zone information
output "zone_info" {
  value = {
    name          = data.technitium_zone.example.name
    type          = data.technitium_zone.example.type
    internal      = data.technitium_zone.example.internal
    disabled      = data.technitium_zone.example.disabled
    dnssec_status = data.technitium_zone.example.dnssec_status
    soa_serial    = data.technitium_zone.example.soa_serial
  }
}

# Example using the data source in other resources
resource "technitium_dns_record" "conditional_record" {
  count = data.technitium_zone.example.disabled ? 0 : 1
  
  zone = data.technitium_zone.example.name
  name = "api"
  type = "A"
  ttl  = 300
  data = "192.168.1.200"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The domain name for the zone to retrieve.

## Attribute Reference

The following attributes are exported:

* `id` - The unique identifier for the zone (same as the zone name).

* `type` - The type of zone (Primary, Secondary, Stub, Forwarder, SecondaryForwarder, Catalog, SecondaryCatalog).

* `catalog` - The name of the catalog zone this zone is a member of, if any.

* `use_soa_serial_date_scheme` - Whether the zone uses date scheme for SOA serial.

* `primary_name_server_addresses` - Comma separated list of IP addresses or domain names of the primary name server for Secondary, SecondaryForwarder, SecondaryCatalog, and Stub zones.

* `zone_transfer_protocol` - The zone transfer protocol used (Tcp, Tls, Quic).

* `tsig_key_name` - The TSIG key name used, if any.

* `validate_zone` - Whether ZONEMD validation is enabled.

* `initialize_forwarder` - Whether the Conditional Forwarder zone is initialized with an FWD record.

* `protocol` - The DNS transport protocol used by the Conditional Forwarder zone.

* `forwarder` - The address of the DNS server used as a forwarder.

* `dnssec_validation` - Whether DNSSEC validation is enabled.

* `proxy_type` - The type of proxy for conditional forwarding.

* `proxy_address` - The proxy server address used, if any.

* `proxy_port` - The proxy server port used, if any.

* `proxy_username` - The proxy server username used, if any.

* `proxy_password` - The proxy server password used, if any.

* `internal` - Whether this is an internal zone.

* `dnssec_status` - The DNSSEC status of the zone.

* `disabled` - Whether the zone is disabled.

* `soa_serial` - The SOA serial number of the zone.