---
subcategory: "Transit Gateway Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_device"
description: |-
  Creates a device in a global network.
---

# Resource: aws_networkmanager_device

Creates a device in a global network. If you specify both a site ID and a location,
the location of the site is used for visualization in the Network Manager console.

## Example Usage

```hcl
resource "aws_networkmanager_global_network" "example" {
}

resource "aws_networkmanager_site" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
}

resource "aws_networkmanager_device" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
  site_id           = aws_networkmanager_site.example.id
}
```

## Argument Reference

The following arguments are supported:

* `global_network_id` - (Required) The ID of the Global Network to create the device in.
* `description` - (Optional) Description of the device.
* `location` - (Optional) The device location as documented below.
* `type` - (Optional) The type of device.
* `model` - (Optional) The model of device.
* `serial_number` - (Optional) The serial number of the device.
* `site_id` - (Optional) The ID of the Site to create device in.
* `vendor` - (Optional) The vendor of the device.
* `tags` - (Optional) Key-value tags for the device.

The `location` object supports the following:

* `address` - (Optional) Address of the location.
* `latitude` - (Optional) Latitude of the location.
* `longitude` - (Optional) Longitude of the location.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Device Amazon Resource Name (ARN)

## Import

`aws_networkmanager_device` can be imported using the device ARN, e.g.

```
$ terraform import aws_networkmanager_device.example arn:aws:networkmanager::123456789012:device/global-network-0d47f6t230mz46dy4/device-07f6fd08867abc123
```
