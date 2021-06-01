---
subcategory: "Firewall Manager (FMS)"
layout: "aws"
page_title: "AWS: aws_fms_admin_account"
description: |-
  Provides a resource to associate/disassociate an AWS Firewall Manager administrator account
---

# Resource: aws_fms_admin_account

Provides a resource to associate/disassociate an AWS Firewall Manager administrator account. This operation must be performed in the `us-east-1` region.

## Example Usage

```hcl
resource "aws_fms_admin_account" "example" {}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Optional) The AWS account ID to associate with AWS Firewall Manager as the AWS Firewall Manager administrator account. This can be an AWS Organizations master account or a member account. Defaults to the current account. Must be configured to perform drift detection.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The AWS account ID of the AWS Firewall Manager administrator account.

## Import

Firewall Manager administrator account association can be imported using the account ID, e.g.

```
$ terraform import aws_fms_admin_account.example 123456789012
```
