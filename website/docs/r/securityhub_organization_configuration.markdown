---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_organization_configuration"
description: |-
Auto Enables Security Hub in AWS organization member accounts.
---

# Resource: aws_securityhub_organization_configuration

Auto Enables Security Hub in AWS organization member accounts. 

[aws_securityhub_organization_admin_account]: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/securityhub_organization_admin_account

[Managing administrator and member accounts]: https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-accounts.html

~> **NOTE:** This resource requires an [aws_securityhub_organization_admin_account] to be configured (not necessarily with Terraform). More information about managing Security Hub in an organization can be found in the [Managing administrator and member accounts] documentation

~> **NOTE:** Destroying this resource will disable auto enable Security Hub in AWS organization member accounts.

## Example Usage

```terraform
resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["securityhub.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_securityhub_organization_admin_account" "example" {
  depends_on = [aws_organizations_organization.example]

  admin_account_id = "123456789012"
}

resource "aws_securityhub_organization_configuration" "example" {}
```

## Argument Reference

The resource does not support any arguments.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS Account ID.

## Import

An existing Security Hub enabled account can be imported using the AWS account ID, e.g.

```
$ terraform import aws_securityhub_organization_configuration.example 123456789012
```
