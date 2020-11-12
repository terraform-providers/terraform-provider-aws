---
subcategory: "Transit Gateway Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_device"
description: |-
  Retrieve information about a device.
---

# Data Source: aws_networkmanager_device

Retrieve information about a device.

## Example Usage

```hcl
data "aws_networkmanager_device" "example" {
  id = var.global_network_id
}
```

## Argument Reference

* `id` - (Optional) The id of the specific device to retrieve.

* `global_network_id` - (Required) The ID of the Global Network of the device to retrieve.

* `site_id` - (Optional) The ID of the Site of the device to retrieve.

* `tags` - (Optional) A map of tags, each pair of which must exactly match
  a pair on the desired device.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `description` - Description of the device.
* `location` - The device location as documented below.
* `type` - The type of device.
* `model` - The model of device.
* `serial_number` - The serial number of the device.
* `vendor` - The vendor of the device.

The `location` object supports the following:

* `address` - Address of the location.
* `latitude` - Latitude of the location.
* `longitude` - Longitude of the location.
