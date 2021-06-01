---
subcategory: "WAFv2"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl"
description: |-
  Retrieves the summary of a WAFv2 Web ACL.
---

# Data Source: aws_wafv2_web_acl

Retrieves the summary of a WAFv2 Web ACL.

## Example Usage

```hcl
data "aws_wafv2_web_acl" "example" {
  name  = "some-web-acl"
  scope = "REGIONAL"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the WAFv2 Web ACL.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1` (N. Virginia) on the AWS provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the entity.
* `description` - The description of the WebACL that helps with identification.
* `id` - The unique identifier of the WebACL.
