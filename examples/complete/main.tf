# Complete example showing a typical DNS setup with Technitium
terraform {
  required_providers {
    technitium = {
      source = "kusold/technitium-dns-server"
    }
  }
}

provider "technitium" {
  host     = var.technitium_host
  username = var.technitium_username
  password = var.technitium_password
}

# Variables
variable "technitium_host" {
  description = "Technitium DNS Server URL"
  type        = string
  default     = "http://localhost:5380"
}

variable "technitium_username" {
  description = "Technitium username"
  type        = string
  default     = "admin"
}

variable "technitium_password" {
  description = "Technitium password"
  type        = string
  sensitive   = true
  default     = "admin"
}

variable "domain_name" {
  description = "Primary domain name"
  type        = string
  default     = "mycompany.com"
}

variable "web_server_ip" {
  description = "Web server IP address"
  type        = string
  default     = "192.168.1.100"
}

variable "mail_server_ip" {
  description = "Mail server IP address"
  type        = string
  default     = "192.168.1.200"
}

# Primary DNS Zone
resource "technitium_zone" "primary_domain" {
  name = var.domain_name
  type = "Primary"

  # SOA record configuration
  soa_primary_name_server = "ns1.${var.domain_name}"
  soa_responsible_person  = "admin.${var.domain_name}"
  soa_refresh             = 3600
  soa_retry               = 1800
  soa_expire              = 1209600
  soa_minimum             = 300
}

# Root domain A record
resource "technitium_dns_record" "root_domain" {
  zone = technitium_zone.primary_domain.name
  name = "@"
  type = "A"
  ttl  = 300
  data = var.web_server_ip
}

# WWW subdomain
resource "technitium_dns_record" "www" {
  zone = technitium_zone.primary_domain.name
  name = "www"
  type = "CNAME"
  ttl  = 300
  data = var.domain_name
}

# Mail exchange record
resource "technitium_dns_record" "mx" {
  zone     = technitium_zone.primary_domain.name
  name     = "@"
  type     = "MX"
  ttl      = 300
  data     = "mail.${var.domain_name}"
  priority = 10
}

# Mail server A record
resource "technitium_dns_record" "mail_server" {
  zone = technitium_zone.primary_domain.name
  name = "mail"
  type = "A"
  ttl  = 300
  data = var.mail_server_ip
}

# Name servers
resource "technitium_dns_record" "ns1" {
  zone = technitium_zone.primary_domain.name
  name = "@"
  type = "NS"
  ttl  = 86400
  data = "ns1.${var.domain_name}"
}

resource "technitium_dns_record" "ns2" {
  zone = technitium_zone.primary_domain.name
  name = "@"
  type = "NS"
  ttl  = 86400
  data = "ns2.${var.domain_name}"
}

# Name server A records
resource "technitium_dns_record" "ns1_a" {
  zone = technitium_zone.primary_domain.name
  name = "ns1"
  type = "A"
  ttl  = 86400
  data = var.web_server_ip
}

resource "technitium_dns_record" "ns2_a" {
  zone = technitium_zone.primary_domain.name
  name = "ns2"
  type = "A"
  ttl  = 86400
  data = var.mail_server_ip
}

# SPF record
resource "technitium_dns_record" "spf" {
  zone = technitium_zone.primary_domain.name
  name = "@"
  type = "TXT"
  ttl  = 300
  data = "v=spf1 a mx include:_spf.google.com ~all"
}

# DMARC record
resource "technitium_dns_record" "dmarc" {
  zone = technitium_zone.primary_domain.name
  name = "_dmarc"
  type = "TXT"
  ttl  = 300
  data = "v=DMARC1; p=quarantine; rua=mailto:dmarc@${var.domain_name}; ruf=mailto:dmarc@${var.domain_name}"
}

# Common subdomains
resource "technitium_dns_record" "api" {
  zone = technitium_zone.primary_domain.name
  name = "api"
  type = "A"
  ttl  = 300
  data = var.web_server_ip
}

resource "technitium_dns_record" "ftp" {
  zone = technitium_zone.primary_domain.name
  name = "ftp"
  type = "CNAME"
  ttl  = 300
  data = var.domain_name
}

# Development subdomain zone
resource "technitium_zone" "dev_subdomain" {
  name = "dev.${var.domain_name}"
  type = "Primary"

  soa_primary_name_server = "ns1.${var.domain_name}"
  soa_responsible_person  = "admin.${var.domain_name}"
}

# Development environment records
resource "technitium_dns_record" "dev_www" {
  zone = technitium_zone.dev_subdomain.name
  name = "www"
  type = "A"
  ttl  = 300
  data = "192.168.1.150"
}

resource "technitium_dns_record" "dev_api" {
  zone = technitium_zone.dev_subdomain.name
  name = "api"
  type = "A"
  ttl  = 300
  data = "192.168.1.151"
}

# DNS Apps for enhanced functionality
resource "technitium_dns_app" "wild_ip" {
  name           = "Wild IP"
  install_method = "url"
  url            = "https://download.technitium.com/dns/apps/WildIpApp.zip"

  config = jsonencode({
    "enabled" = true
    "ipv4"    = true
    "ipv6"    = true
  })
}

resource "technitium_dns_app" "geo_country" {
  name           = "Geo Country"
  install_method = "url"
  url            = "https://download.technitium.com/dns/apps/GeoCountryApp.zip"

  config = jsonencode({
    "enabled"        = true
    "defaultCountry" = "US"
  })
}

# Data sources for monitoring
data "technitium_dns_apps" "installed_apps" {}

data "technitium_dns_store_apps" "available_apps" {}

# Outputs
output "primary_zone_name" {
  description = "Primary DNS zone name"
  value       = technitium_zone.primary_domain.name
}

output "zone_records_count" {
  description = "Number of DNS records created"
  value = length([
    technitium_dns_record.root_domain,
    technitium_dns_record.www,
    technitium_dns_record.mx,
    technitium_dns_record.mail_server,
    technitium_dns_record.ns1,
    technitium_dns_record.ns2,
    technitium_dns_record.ns1_a,
    technitium_dns_record.ns2_a,
    technitium_dns_record.spf,
    technitium_dns_record.dmarc,
    technitium_dns_record.api,
    technitium_dns_record.ftp,
  ])
}

output "development_zone_name" {
  description = "Development subdomain zone name"
  value       = technitium_zone.dev_subdomain.name
}

output "installed_dns_apps" {
  description = "List of installed DNS applications"
  value = {
    for app in data.technitium_dns_apps.installed_apps.apps : app.name => {
      version  = app.version
      dns_apps = length(app.dns_apps)
    }
  }
}

output "available_store_apps" {
  description = "Count of available apps in the store"
  value       = length(data.technitium_dns_store_apps.available_apps.store_apps)
}
