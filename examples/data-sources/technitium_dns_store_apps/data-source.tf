# Get all available DNS apps from the store
data "technitium_dns_store_apps" "all" {}

# Display store apps information
output "store_apps" {
  value = {
    for app in data.technitium_dns_store_apps.all.store_apps : app.name => {
      version           = app.version
      description       = app.description
      size              = app.size
      installed         = app.installed
      installed_version = app.installed_version
      update_available  = app.update_available
    }
  }
}

# Get only uninstalled apps
locals {
  uninstalled_apps = [
    for app in data.technitium_dns_store_apps.all.store_apps : app
    if !app.installed
  ]
}

output "uninstalled_apps" {
  value = local.uninstalled_apps
}

# Get apps that have updates available
locals {
  apps_with_updates = [
    for app in data.technitium_dns_store_apps.all.store_apps : app
    if app.update_available
  ]
}

output "apps_with_updates" {
  value = local.apps_with_updates
}