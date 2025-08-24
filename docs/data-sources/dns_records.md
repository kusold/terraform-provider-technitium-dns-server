---
page_title: "<no value> <no value>"
subcategory: "DNS Records"
description: |-
  <no value>
---

# <no value> (<no value>)

<no value>

## Example Usage

```terraform
# Data source to list all DNS records in a zone
data "technitium_dns_records" "all_records" {
  zone = "example.com"
}

# Data source to filter records by type
data "technitium_dns_records" "a_records_only" {
  zone         = "example.com"
  record_types = ["A"]
}

# Data source to filter multiple record types
data "technitium_dns_records" "web_records" {
  zone         = "example.com"
  record_types = ["A", "AAAA", "CNAME"]
}

# Output all records information
output "all_records" {
  value = {
    zone          = data.technitium_dns_records.all_records.zone
    total_records = length(data.technitium_dns_records.all_records.records)
    records       = data.technitium_dns_records.all_records.records
  }
}

# Output filtered A records
output "a_records" {
  value = [
    for record in data.technitium_dns_records.a_records_only.records : {
      name = record.name
      data = record.data
      ttl  = record.ttl
    }
  ]
}

# Example: Create monitoring records based on existing records
resource "technitium_dns_record" "monitoring_records" {
  for_each = {
    for record in data.technitium_dns_records.a_records_only.records :
    record.name => record if record.name != "@"
  }

  zone = "monitoring.example.com"
  name = "check-${each.key}"
  type = "CNAME"
  ttl  = 300
  data = "${each.key}.example.com"
}
```

## Argument Reference

The following arguments are supported:

## Attribute Reference

In addition to all arguments above, the following attributes are exported:
