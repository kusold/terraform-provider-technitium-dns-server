---
page_title: "technitium_zone Resource"
subcategory: "Zone Management"
description: |-
  Manages a DNS zone in Technitium DNS Server.
---

# technitium_zone Resource

Manages a DNS zone in Technitium DNS Server. Zones are the basic building blocks for DNS management and can be configured as Primary (authoritative), Secondary (slave), Forwarder (conditional forwarding), Stub, or Catalog zones.

## Example Usage

```terraform
# Primary DNS Zone
resource "technitium_zone" "example_primary" {
  name = "example.com"
  type = "Primary"
  use_soa_serial_date_scheme = true
}

# Secondary DNS Zone
resource "technitium_zone" "example_secondary" {
  name = "secondary.example.com"
  type = "Secondary"
  
  # Primary name server for zone transfers
  primary_name_server_addresses = "192.168.1.10,192.168.1.11"
  zone_transfer_protocol = "Tcp"
}

# Conditional Forwarder Zone
resource "technitium_zone" "example_forwarder" {
  name = "internal.company.com"
  type = "Forwarder"
  
  # Forwarder configuration
  forwarder = "10.0.0.10"
  protocol = "Udp"
  
  # Optional: Enable DNSSEC validation
  dnssec_validation = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The domain name for the zone. Can be a valid domain name, IP address, or network address in CIDR format for reverse zones.

* `type` - (Required) The type of zone to create. Valid values are: `Primary`, `Secondary`, `Stub`, `Forwarder`, `SecondaryForwarder`, `Catalog`, `SecondaryCatalog`.

* `catalog` - (Optional) The name of the catalog zone to become its member zone. Valid only for Primary, Stub, and Forwarder zones.

* `use_soa_serial_date_scheme` - (Optional) Set to true to enable using date scheme for SOA serial. Valid for Primary, Forwarder, and Catalog zones. Defaults to false.

* `primary_name_server_addresses` - (Optional) Comma separated list of IP addresses or domain names of the primary name server. Used only with Secondary, SecondaryForwarder, SecondaryCatalog, and Stub zones.

* `zone_transfer_protocol` - (Optional) The zone transfer protocol to be used. Valid values are: `Tcp`, `Tls`, `Quic`. Used by Secondary, SecondaryForwarder, and SecondaryCatalog zones. Defaults to `Tcp`.

* `tsig_key_name` - (Optional) The TSIG key name to be used. Used by Secondary, SecondaryForwarder, and SecondaryCatalog zones.

* `validate_zone` - (Optional) Set to true to enable ZONEMD validation. Valid only for Secondary zones. Defaults to false.

* `initialize_forwarder` - (Optional) Set to true to initialize the Conditional Forwarder zone with an FWD record. Valid for Forwarder zones. Defaults to false.

* `protocol` - (Optional) The DNS transport protocol to be used by the Conditional Forwarder zone. Valid values are: `Udp`, `Tcp`, `Tls`, `Https`, `Quic`. Defaults to `Udp`.

* `forwarder` - (Optional) The address of the DNS server to be used as a forwarder. Use 'this-server' to forward internally. Required for Conditional Forwarder zones.

* `dnssec_validation` - (Optional) Set to true to indicate if DNSSEC validation must be done. Used with Conditional Forwarder zones. Defaults to false.

* `proxy_type` - (Optional) The type of proxy for conditional forwarding. Valid values are: `NoProxy`, `DefaultProxy`, `Http`, `Socks5`. Defaults to `DefaultProxy`.

* `proxy_address` - (Optional) The proxy server address to use when proxy_type is configured.

* `proxy_port` - (Optional) The proxy server port to use when proxy_type is configured.

* `proxy_username` - (Optional) The proxy server username to use when proxy_type is configured.

* `proxy_password` - (Optional) The proxy server password to use when proxy_type is configured.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier for the zone resource (same as the zone name).

* `internal` - Indicates if this is an internal zone.

* `dnssec_status` - The DNSSEC status of the zone.

* `disabled` - Indicates if the zone is disabled.

* `soa_serial` - The SOA serial number of the zone.

## Import

Zones can be imported using the zone name:

```
terraform import technitium_zone.example example.com
```
