# A Record (IPv4)
resource "technitium_dns_record" "example_a" {
  zone = "example.com"
  name = "www"
  type = "A"
  ttl  = 300
  data = "192.168.1.100"
}

# AAAA Record (IPv6)
resource "technitium_dns_record" "example_aaaa" {
  zone = "example.com"
  name = "www"
  type = "AAAA"
  ttl  = 300
  data = "2001:db8::1"
}

# CNAME Record
resource "technitium_dns_record" "example_cname" {
  zone = "example.com"
  name = "blog"
  type = "CNAME"
  ttl  = 300
  data = "www.example.com"
}

# MX Record (Mail Exchange)
resource "technitium_dns_record" "example_mx" {
  zone     = "example.com"
  name     = "@" # Root domain
  type     = "MX"
  ttl      = 300
  data     = "mail.example.com"
  priority = 10
}

# TXT Record
resource "technitium_dns_record" "example_txt_spf" {
  zone = "example.com"
  name = "@"
  type = "TXT"
  ttl  = 300
  data = "v=spf1 include:_spf.google.com ~all"
}

# TXT Record for Domain Verification
resource "technitium_dns_record" "example_txt_verification" {
  zone = "example.com"
  name = "_dmarc"
  type = "TXT"
  ttl  = 300
  data = "v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"
}

# NS Record (Name Server)
resource "technitium_dns_record" "example_ns" {
  zone = "example.com"
  name = "subdomain"
  type = "NS"
  ttl  = 86400
  data = "ns1.subdomain.example.com"
}

# SRV Record (Service Location)
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

# PTR Record (Reverse DNS)
resource "technitium_dns_record" "example_ptr" {
  zone = "1.168.192.in-addr.arpa"
  name = "100"
  type = "PTR"
  ttl  = 300
  data = "www.example.com"
}

# TXT Record with Comments
resource "technitium_dns_record" "example_txt_with_comments" {
  zone     = "example.com"
  name     = "info"
  type     = "TXT"
  ttl      = 300
  data     = "This is a text record with some information"
  comments = "This record contains company information"
}