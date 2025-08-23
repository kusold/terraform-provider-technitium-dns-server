terraform {
  required_providers {
    technitium = {
      source = "kusold/technitium-dns-server"
    }
  }
}

provider "technitium" {
  # Technitium DNS Server URL
  host = "http://localhost:5380"

  # Authentication using username/password
  username = "admin"
  password = "admin"

  # Alternative: Authentication using API token
  # token = "your-api-token-here"
}