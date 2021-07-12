---
subcategory: "IoT"
layout: "aws"
page_title: "AWS: aws_iotwireless_device_profile"
description: |-
    Creates and manages an AWS IoT Wireless Device Profile.
---

# Resource: aws_iotwireless_device_profile

Creates and manages an AWS IoT Wireless Device Profile.

## Example Usage

```terraform
resource "aws_iotwireless_device_profile" "my_profile" {
  name = "my_profile"

  lorawan {
    mac_version         = "1.0.3"
    reg_params_revision = "Regional Parameters v1.0.3rA"
    max_eirp            = 15
    max_duty_cycle      = 10
    rf_region           = "AU915"
    supports_join       = true
    supports_32bit_fcnt = true
  }

  tags = {
    Terraform = true
  }
}
```

## Argument Reference

* `name` - (Required) The name of the device profile.
* `lorawan` - (Required) LoRaWAN configuration. (documented below)
* `tags` - (Optional) Specifies tags key and value.

The `lorawan` object supports the following:

* `class_b_timeout` - (Optional) Class B timeout.
* `class_c_timeout` - (Optional) Class C timeout.
* `factory_preset_freqs_list` - (Optional) Factory preset frequencies list.
* `mac_version` - (Optional) MAC Version.
* `max_duty_cycle` - (Optional) Max duty cycle.
* `max_eirp` - (Optional) Max EIRP.
* `ping_slot_dr` - (Optional) Ping slot data rate
* `ping_slot_freq` - (Optional) Ping slot frequency
* `ping_slot_period` - (Optional) Ping slot period
* `reg_params_revision` - (Optional) Regional parameters revision
* `rf_region` - (Optional) RF region.
* `rx_data_rate_2` - (Optional) RX data rate 2.
* `rx_delay_1` - (Optional) RX delay 1.
* `rx_dr_offset_1` - (Optional) RX data rate offset 1.
* `rx_freq_2` - (Optional) RX frequency 2.
* `supports_32bit_fcnt` - (Optional) Supports 32bit frame count.
* `supports_class_b` - (Optional) Supports class B.
* `supports_class_c` - (Optional) Supports class C.
* `supports_join` - (Optional) Supports join.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the device profile

## Import

IOT Wireless Device Profiles can be imported using the id, e.g.

```
$ terraform import aws_iotwireless_device_profile.my_profile 39b8940a-6d12-49b8-a320-bc42293dd0e1
```
