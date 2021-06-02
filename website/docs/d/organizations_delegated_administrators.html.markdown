---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_delegated_administrators"
description: |-
  Get a list the AWS accounts that are designated as delegated administrators in this organization 
---

# Data Source: aws_organizations_delegated_administrators

Get a list the AWS accounts that are designated as delegated administrators in this organization

## Example Usage

```terraform
data "aws_organizations_delegated_administrators" "example" {
  service_principal = "SERVICE PRINCIPAL"
}
```


## Argument Reference

* `service_principal` - (Optional) Specifies a service principal name. If specified, then the operation lists the delegated administrators only for the specified service. If you don't specify a service principal, the operation lists all delegated administrators for all services in your organization.

## Attributes Reference

* `delegated_administrators` - The list of delegated administrators in your organization, which have the following attributes:
    * `arn` - The Amazon Resource Name (ARN) of the delegated administrator's account.
    * `delegation_enabled_date` - The date when the account was made a delegated administrator.
    * `email` - The email address that is associated with the delegated administrator's AWS account.
    * `id` - The unique identifier (ID) of the delegated administrator's account.
    * `joined_method` - The method by which the delegated administrator's account joined the organization.
    * `joined_timestamp` - The date when the delegated administrator's account became a part of the organization.
    * `name` - The friendly name of the delegated administrator's account.
    * `status` - The status of the delegated administrator's account in the organization.
