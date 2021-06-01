---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_aggregate_authorization"
description: |-
  Manages an AWS Config Aggregate Authorization.
---

# Resource: aws_config_aggregate_authorization

Manages an AWS Config Aggregate Authorization

## Example Usage

```hcl
resource "aws_config_aggregate_authorization" "example" {
  account_id = "123456789012"
  region     = "eu-west-2"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) Account ID
* `region` - (Required) Region
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the authorization

## Import

Config aggregate authorizations can be imported using `account_id:region`, e.g.

```
$ terraform import aws_config_aggregate_authorization.example 123456789012:us-east-1
```
