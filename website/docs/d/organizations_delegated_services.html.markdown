---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_delegated_services"
description: |-
  Get a list the AWS services for which the specified account is a delegated administrator 
---

# Data Source: aws_organizations_delegated_services

Get a list the AWS services for which the specified account is a delegated administrator

## Example Usage

```terraform
data "aws_organizations_delegated_services" "example" {
  account_id = "AWS ACCOUNT ID"
}
```


## Argument Reference

* `account_id` - (Required) The account ID number of a delegated administrator account in the organization.

## Attributes Reference

* `delegated_services` - The services for which the account is a delegated administrator, which have the following attributes:
    * `delegation_enabled_date` - The date that the account became a delegated administrator for this service.
    * `service_principal` - The name of an AWS service that can request an operation for the specified service.
