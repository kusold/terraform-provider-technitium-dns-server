# Get all installed DNS apps
data "technitium_dns_apps" "all" {}

# Display installed apps information
output "installed_apps" {
  value = {
    for app in data.technitium_dns_apps.all.apps : app.name => {
      version      = app.version
      dns_apps     = app.dns_apps
    }
  }
}

# Check if a specific app is installed
locals {
  wild_ip_installed = contains([for app in data.technitium_dns_apps.all.apps : app.name], "Wild IP")
}

output "wild_ip_status" {
  value = local.wild_ip_installed ? "Wild IP is installed" : "Wild IP is not installed"
}
