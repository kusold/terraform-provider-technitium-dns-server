---
page_title: "Getting Started with Technitium DNS Server"
subcategory: "Guides"
description: |-
  A guide to getting started with Technitium DNS Server and the Terraform provider.
---

# Getting Started with Technitium DNS Server

This guide will help you set up Technitium DNS Server and use the Terraform provider to manage your DNS infrastructure.

## Installing Technitium DNS Server

Technitium DNS Server can be installed in several ways:

### Docker (Recommended for Testing)

The quickest way to get started is using Docker:

```bash
docker run -d \
  --name technitium-dns \
  -p 5380:5380 \
  -p 53:53/udp \
  -v technitium_data:/app/config \
  technitium/dns-server
```

### Self-Hosted Installation

You can download and install Technitium DNS Server on various platforms:

1. Visit the [Technitium DNS Server GitHub releases page](https://github.com/TechnitiumSoftware/DnsServer/releases)
2. Download the version appropriate for your operating system
3. Install and configure following the instructions for your platform

## Initial Configuration

After installing Technitium DNS Server:

1. Access the web console at `http://localhost:5380/`
2. Set up the administrator account with a username and secure password
3. Configure your network settings as needed

## Setting Up the Terraform Provider

1. Create a new Terraform project directory
2. Create a `main.tf` file with the provider configuration:

```terraform
terraform {
  required_providers {
    technitium = {
      source = "kusold/technitium-dns-server"
      version = "~> 0.1"
    }
  }
}

provider "technitium" {
  host     = "http://localhost:5380"
  username = "admin"
  password = "password" # Use variables for sensitive data in production
}
```

3. Initialize the Terraform project:

```bash
terraform init
```

## Creating Your First DNS Zone

Add the following to your Terraform configuration:

```terraform
resource "technitium_zone" "example_com" {
  name = "example.com"
  type = "Primary"
}

# Add an A record to the zone
resource "technitium_dns_record" "www" {
  zone = technitium_zone.example_com.name
  name = "www"
  type = "A"
  ttl  = 3600
  data = "192.0.2.10"
}
```

Apply the configuration:

```bash
terraform apply
```

## Verifying the Configuration

You can verify that your configuration has been applied correctly by:

1. Checking the Technitium web console at `http://localhost:5380/`
2. Using the DNS records data source to retrieve records:

```terraform
data "technitium_dns_records" "example_records" {
  zone = "example.com"
}

output "records" {
  value = data.technitium_dns_records.example_records.records
}
```

3. Testing DNS resolution:

```bash
dig @localhost www.example.com
```

## Next Steps

- Create additional zones and records
- Set up DNSSEC
- Configure forwarder zones
- Add advanced record types like MX, SRV, or TXT

## Additional Resources

- [Technitium DNS Server Documentation](https://github.com/TechnitiumSoftware/DnsServer/blob/master/README.md)
- [Technitium DNS Server API Documentation](https://github.com/TechnitiumSoftware/DnsServer/blob/master/APIDOCS.md)
- [Docker Environment Variables](https://github.com/TechnitiumSoftware/DnsServer/blob/master/DockerEnvironmentVariables.md)