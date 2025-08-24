---
page_title: "Technitium DNS Server Provider"
description: |-
  The Technitium DNS Server provider is used to interact with resources provided by Technitium DNS Server.
---

# Technitium DNS Server Provider

The Technitium DNS Server provider is used to manage resources in a [Technitium DNS Server](https://github.com/TechnitiumSoftware/DnsServer) instance via its REST API. Technitium DNS Server is a free, open source, cross-platform, authoritative and recursive DNS server that can be used for self-hosting a DNS server for privacy & security.

This provider enables Infrastructure as Code practices for managing:

- DNS zones (Primary, Secondary, Forwarder, etc.)
- DNS records (A, AAAA, CNAME, MX, TXT, etc.)
- Domain name resolution settings
- And more

## Example Usage

```terraform
terraform {
  required_providers {
    technitium = {
      source = "kusold/technitium-dns-server"
      version = "~> 0.1"
    }
  }
}

# Configure the Technitium DNS Server Provider
provider "technitium" {
  host     = "http://localhost:5380"
  username = "admin"
  password = "password"

  # Alternative authentication using token
  # token = var.technitium_token
}

# Create a primary DNS zone
resource "technitium_zone" "example_com" {
  name = "example.com"
  type = "Primary"
}

# Create an A record
resource "technitium_dns_record" "www" {
  zone = technitium_zone.example_com.name
  name = "www"
  type = "A"
  ttl  = 3600
  data = "192.0.2.10"
}
```

## Authentication

The Technitium DNS Server provider supports two authentication methods:

- Username and password authentication
- API token authentication

### Username and Password Authentication

```terraform
provider "technitium" {
  host     = "http://localhost:5380"
  username = "admin"
  password = "password"
}
```

### Token Authentication

```terraform
provider "technitium" {
  host  = "http://localhost:5380"
  token = var.technitium_token
}
```

## Argument Reference

- `host` - (Required) The URL of the Technitium DNS Server instance. Must include protocol (http/https) and port number if non-standard.
- `username` - (Optional) The username for authentication. Either username/password or token must be provided.
- `password` - (Optional) The password for authentication. Either username/password or token must be provided.
- `token` - (Optional) The API token for authentication. Either username/password or token must be provided.
- `timeout_seconds` - (Optional) Request timeout in seconds. Defaults to 30.
- `retry_attempts` - (Optional) Number of retry attempts for failed requests. Defaults to 3.
- `insecure_skip_verify` - (Optional) Skip TLS certificate verification. Defaults to false.
