---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ami_launch_permission"
description: |-
  Adds launch permission to Amazon Machine Image (AMI).
---

# Resource: aws_ami_launch_permission

Adds launch permission to Amazon Machine Image (AMI) from another AWS account.

## Example Usage

```hcl
resource "aws_ami_launch_permission" "example" {
  image_id   = "ami-12345678"
  account_id = "123456789012"
}
```

## Argument Reference

The following arguments are supported:

  * `image_id` - (required) A region-unique name for the AMI.
  * `account_id` - (required) An AWS Account ID to add launch permissions.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

  * `id` - A combination of "`image_id`-`account_id`".

## Import

AWS AMI Launch Permission can be imported using the `ACCOUNT-ID/IMAGE-ID`, e.g.

```sh
$ terraform import aws_ami_launch_permission.example 123456789012/ami-12345678
```
