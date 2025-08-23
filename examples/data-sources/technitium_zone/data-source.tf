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