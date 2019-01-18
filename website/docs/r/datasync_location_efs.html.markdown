---
layout: "aws"
page_title: "AWS: aws_datasync_location_efs"
sidebar_current: "docs-aws-resource-datasync-location-efs"
description: |-
  Manages an EFS Location within AWS DataSync.
---

# aws_datasync_location_efs

Manages an AWS DataSync EFS Location.

~> **NOTE:** The EFS File System must have a mounted EFS Mount Target before creating this resource.

## Example Usage

```hcl
resource "aws_datasync_location_efs" "example" {
  # The below example uses aws_efs_mount_target as a reference to ensure a mount target already exists when resource creation occurs.
  # You can accomplish the same behavior with depends_on or an aws_efs_mount_target data source reference.
  efs_file_system_arn = "${aws_efs_mount_target.example.file_system_arn}"

  ec2_config {
    security_group_arns = ["${aws_security_group.example.arn}"]
    subnet_arn          = "${aws_subnet.example.arn}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `ec2_config` - (Required) Configuration block containing EC2 configurations for connecting to the EFS File System.
* `efs_file_system_arn` - (Required) Amazon Resource Name (ARN) of EFS File System.
* `subdirectory` - (Optional) Subdirectory to perform actions as source or destination. Default `/`.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location.

### ec2_config Argument Reference

The following arguments are supported inside the `ec2_config` configuration block:

* `security_group_arns` - (Required) List of Amazon Resource Names (ARNs) of the EC2 Security Groups that are associated with the EFS Mount Target.
* `subnet_arn` - (Required) Amazon Resource Name (ARN) of the EC2 Subnet that is associated with the EFS Mount Target.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the DataSync Location.
* `arn` - Amazon Resource Name (ARN) of the DataSync Location.

## Import

`aws_datasync_location_efs` can be imported by using the DataSync Task Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_datasync_location_efs.example arn:aws:datasync:us-east-1:123456789012:location/loc-12345678901234567
```
