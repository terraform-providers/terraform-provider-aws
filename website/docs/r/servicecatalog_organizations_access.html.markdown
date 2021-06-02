---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_organizations_access"
description: |-
  Manages Service Catalog Organizations Access
---

# Resource: aws_servicecatalog_organizations_access

Manages Service Catalog AWS Organizations Access, a portfolio sharing feature through AWS Organizations. This allows Service Catalog to receive updates on your organization in order to sync your shares with the current structure. This resource will prompt AWS to set `organizations:EnableAWSServiceAccess` on your behalf so that your shares can be in sync with any changes in your AWS Organizations structure.

~> **NOTE:** This resource can only be used by the management account in the organization. In other words, a delegated administrator is not authorized to use the resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_organizations_access" "example" {
  enabled = "true"
}
```

## Argument Reference

The following arguments are required:

* `enabled` - (Required) Whether to enable AWS Organizations access.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Account ID for the account using the resource.
