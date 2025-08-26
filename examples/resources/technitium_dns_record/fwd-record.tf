# Example FWD (Forwarder) DNS record configurations

# Basic FWD record using UDP protocol
resource "technitium_dns_record" "example_fwd_basic" {
  zone = "example.com"
  name = "forward"
  type = "FWD"
  ttl  = 3600
  data = "8.8.8.8"  # Google DNS as forwarder
}

# FWD record with custom protocol
resource "technitium_dns_record" "example_fwd_https" {
  zone      = "example.com"
  name      = "secure-forward"
  type      = "FWD"
  ttl       = 3600
  forwarder = "1.1.1.1"  # Cloudflare DNS
  protocol  = "Https"
}

# FWD record with DNSSEC validation
resource "technitium_dns_record" "example_fwd_dnssec" {
  zone             = "example.com"
  name             = "dnssec-forward"
  type             = "FWD"
  ttl              = 3600
  forwarder        = "9.9.9.9"  # Quad9 DNS
  protocol         = "Tls"
  dnssec_validation = true
}

# FWD record with proxy configuration
resource "technitium_dns_record" "example_fwd_proxy" {
  zone             = "example.com"
  name             = "proxy-forward"
  type             = "FWD"
  ttl              = 3600
  forwarder        = "8.8.8.8"
  protocol         = "Tcp"
  proxy_type       = "Http"
  proxy_address    = "proxy.company.com"
  proxy_port       = 8080
  proxy_username   = "proxyuser"
  proxy_password   = "proxypass"
}

# FWD record with priority and advanced settings
resource "technitium_dns_record" "example_fwd_advanced" {
  zone              = "example.com"
  name              = "advanced-forward"
  type              = "FWD"
  ttl               = 1800
  forwarder         = "1.1.1.1"
  protocol          = "Quic"
  forwarder_priority = 10  # Higher priority (lower number gets tried first)
  dnssec_validation = true
  comments          = "Advanced forwarder with QUIC protocol"
}

# FWD record pointing to this-server (internal forwarding)
resource "technitium_dns_record" "example_fwd_internal" {
  zone      = "example.com"
  name      = "internal-forward"
  type      = "FWD"
  ttl       = 3600
  forwarder = "this-server"  # Special value for internal forwarding
  protocol  = "Udp"
  comments  = "Forward to this DNS server internally"
}
