# Data source to read an existing DNS zone
data "technitium_zone" "example" {
  name = "example.com"
}

# Output zone information
output "zone_info" {
  value = {
    name              = data.technitium_zone.example.name
    type              = data.technitium_zone.example.type
    is_active         = data.technitium_zone.example.is_active
    is_expired        = data.technitium_zone.example.is_expired
    records_count     = data.technitium_zone.example.records_count
    last_modified     = data.technitium_zone.example.last_modified
    dnssec_status     = data.technitium_zone.example.dnssec_status
  }
}

# Example using the data source in other resources
resource "technitium_dns_record" "conditional_record" {
  count = data.technitium_zone.example.is_active ? 1 : 0
  
  zone = data.technitium_zone.example.name
  name = "api"
  type = "A"
  ttl  = 300
  data = "192.168.1.200"
}