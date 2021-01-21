---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_region_settings"
description: |-
  Provides an AWS Backup Region Settings resource.
---

# Resource: aws_backup_region_settings

Provides an AWS Backup Region Settings resource.

## Example Usage

```hcl
resource "aws_backup_region_settings" "test" {
  resource_type_opt_in_preference = {
    "DynamoDB"        = true
    "Aurora"          = true
    "EBS"             = true
    "EC2"             = true
    "EFS"             = true
    "FSx"             = true
    "RDS"             = true
    "Storage Gateway" = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `resource_type_opt_in_preference` - (Required) A map of services along with the opt-in preferences for the Region.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The AWS region.

## Import

Backup Region Settings can be imported using the `region`, e.g.

```
$ terraform import aws_backup_region_settings.test us-west-2
```
