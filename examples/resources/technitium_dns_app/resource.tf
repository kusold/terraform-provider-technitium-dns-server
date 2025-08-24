# Install DNS App from URL
resource "technitium_dns_app" "wild_ip" {
  name           = "Wild IP"
  install_method = "url"
  url            = "https://download.technitium.com/dns/apps/WildIpApp.zip"

  # Optional JSON configuration for the app
  config = jsonencode({
    "enabled" = true
    "ipv4"    = true
    "ipv6"    = true
  })
}

# Install DNS App from file content (base64 encoded zip file)
resource "technitium_dns_app" "custom_app" {
  name           = "Custom App"
  install_method = "file"
  file_content   = filebase64("${path.module}/custom-app.zip")

  config = jsonencode({
    "setting1" = "value1"
    "setting2" = 42
  })
}

# Install store app (using URL method with store URL)
resource "technitium_dns_app" "geo_country" {
  name           = "Geo Country"
  install_method = "url"
  url            = "https://download.technitium.com/dns/apps/GeoCountryApp.zip"
}
