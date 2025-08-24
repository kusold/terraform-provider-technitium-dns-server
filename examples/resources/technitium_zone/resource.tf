# Primary DNS Zone
resource "technitium_zone" "example_primary" {
  name = "example.com"
  type = "Primary"

  # Optional: SOA record settings
  soa_primary_name_server = "ns1.example.com"
  soa_responsible_person  = "admin.example.com"
  soa_serial              = 2024082301
  soa_refresh             = 3600
  soa_retry               = 1800
  soa_expire              = 1209600
  soa_minimum             = 300
}

# Secondary DNS Zone
resource "technitium_zone" "example_secondary" {
  name = "secondary.example.com"
  type = "Secondary"

  # Primary name server for zone transfers
  primary_name_server_addresses = ["192.168.1.10", "192.168.1.11"]

  # Optional: TSIG key for secure zone transfers
  # tsig_key_name = "example-key"
}

# Conditional Forwarder Zone
resource "technitium_zone" "example_forwarder" {
  name = "internal.company.com"
  type = "Forwarder"

  # Forwarder configuration
  forwarder_addresses = ["10.0.0.10", "10.0.0.11"]
  forwarder_protocol  = "Udp" # Options: Udp, Tcp, Tls, Https, Quic

  # Optional: Enable DNSSEC validation
  dnssec_validation = true
}
