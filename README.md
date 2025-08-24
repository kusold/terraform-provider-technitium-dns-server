# Terraform Provider for Technitium DNS Server

[![Go Report Card](https://goreportcard.com/badge/github.com/kusold/terraform-provider-technitium-dns-server)](https://goreportcard.com/report/github.com/kusold/terraform-provider-technitium-dns-server)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This Terraform provider enables you to manage resources in [Technitium DNS Server](https://github.com/TechnitiumSoftware/DnsServer) through Infrastructure as Code. Technitium DNS Server is a free, open-source, cross-platform authoritative and recursive DNS server that can be used for self-hosting a DNS server for privacy and security.

## Features

- Create and manage DNS zones (Primary, Secondary, Forwarder, etc.)
- Manage DNS records (A, AAAA, CNAME, MX, TXT, PTR, NS, SRV)
- Query existing zones and records via data sources
- Support for both username/password and token-based authentication

## Documentation

- [Provider Documentation](./docs/index.md)
- [Getting Started Guide](./docs/guides/getting_started.md)

### Resources

- [`technitium_zone`](./docs/resources/zone.md) - Manage DNS zones
- [`technitium_dns_record`](./docs/resources/dns_record.md) - Manage DNS records

### Data Sources

- [`technitium_zone`](./docs/data-sources/zone.md) - Query DNS zone information
- [`technitium_dns_records`](./docs/data-sources/dns_records.md) - Query DNS records

## Usage Example

```terraform
terraform {
  required_providers {
    technitium = {
      source  = "kusold/technitium-dns-server"
      version = "~> 0.1"
    }
  }
}

provider "technitium" {
  host     = "http://localhost:5380"
  username = "admin"
  password = "password"
}

# Create a primary zone
resource "technitium_zone" "example" {
  name = "example.com"
  type = "Primary"
}

# Add A record to the zone
resource "technitium_dns_record" "www" {
  zone = technitium_zone.example.name
  name = "www"
  type = "A"
  ttl  = 3600
  data = "192.168.1.100"
}

# Add MX record
resource "technitium_dns_record" "mx" {
  zone     = technitium_zone.example.name
  name     = "@"
  type     = "MX"
  ttl      = 3600
  data     = "mail.example.com"
  priority = 10
}
```

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.12.x
- [Go](https://golang.org/doc/install) >= 1.18 (to build the provider plugin)
- [Technitium DNS Server](https://github.com/TechnitiumSoftware/DnsServer) >= 6.0

## Building the Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

## Development Setup

This project uses [nix flakes](https://nixos.wiki/wiki/Flakes) and [direnv](https://direnv.net/) for development environment consistency.

After cloning the repository:

1. Install nix with flakes support
2. Install direnv
3. Run `direnv allow` in the project directory
4. The development environment will be automatically set up

## Testing

To run the tests, use the provided task commands:

```shell
# Run unit tests
task test

# Run acceptance tests (requires a running Technitium DNS Server)
task test:acc
```

## License

This provider is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Technitium DNS Server](https://github.com/TechnitiumSoftware/DnsServer) for their excellent open-source DNS server
- The [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) for making provider development easier