# Configure a DNS App
resource "technitium_dns_app_config" "wild_ip_config" {
  name   = "Wild IP"
  config = jsonencode({
    "enabled" = true
    "ipv4"    = true
    "ipv6"    = true
  })
}

# Configure a Split Horizon App with network rules
resource "technitium_dns_app_config" "split_horizon_config" {
  name = "Split Horizon"
  config = jsonencode({
    "networks" = [
      {
        "network"      = "10.0.0.0/8"
        "isPrivate"    = true
        "recordType"   = "A"
        "recordValue"  = "192.168.1.100"
      },
      {
        "network"      = "0.0.0.0/0"
        "isPrivate"    = false
        "recordType"   = "A"
        "recordValue"  = "203.0.113.100"
      }
    ]
  })
}

# Configure a Geo Country App with country mappings
resource "technitium_dns_app_config" "geo_country_config" {
  name = "Geo Country"
  config = jsonencode({
    "countries" = [
      {
        "country"     = "US"
        "recordType"  = "A"
        "recordValue" = "192.0.2.100"
      },
      {
        "country"     = "GB"
        "recordType"  = "A"
        "recordValue" = "192.0.2.200"
      }
    ],
    "defaultRecordType"  = "A",
    "defaultRecordValue" = "192.0.2.1"
  })
}
