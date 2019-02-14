---
layout: "aws"
page_title: "AWS: aws_organizations_organization"
sidebar_current: "docs-aws-resource-organizations-organization"
description: |-
  Provides a resource to create an organization.
---

# aws_organizations_organization

Provides a resource to create an organization.

## Example Usage:

```hcl
resource "aws_organizations_organization" "org" {
  aws_service_access_principals = [
    "cloudtrail.amazonaws.com",
    "config.amazonaws.com",
  ]

  feature_set = "ALL"
}
```

## Argument Reference

The following arguments are supported:

* `aws_service_access_principals` - (Optional) List of AWS service principal names for which you want to enable integration with your organization. This is typically in the form of a URL, such as service-abbreviation.amazonaws.com. Organization must have `feature_set` set to `ALL`. For additional information, see the [AWS Organizations User Guide](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_integrate_services.html).
* `feature_set` - (Optional) Specify "ALL" (default) or "CONSOLIDATED_BILLING".

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the organization
* `id` - Identifier of the organization
* `master_account_arn` - ARN of the master account
* `master_account_email` - Email address of the master account
* `master_account_id` - Identifier of the master account

## Import

The AWS organization can be imported by using the `id`, e.g.

```
$ terraform import aws_organizations_organization.my_org o-1234567
```
