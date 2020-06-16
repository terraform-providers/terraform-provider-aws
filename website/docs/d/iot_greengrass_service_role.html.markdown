---
subcategory: "IoT"
layout: "aws"
page_title: "AWS: aws_iot_greengrass_service_role"
description: |-
  Retrieves the greengrass service role that is attached to your account
---

# Data Source: aws_iot_greengrass_service_role

Returns the greengrass service role that is attached to your account. See also https://docs.aws.amazon.com/greengrass/latest/apireference/-greengrass-servicerole.html

## Example Usage

```hcl
data "aws_iot_greengrass_service_role" "example" {}

output "greengrass_service_role" {
  value = data.aws_iot_greengrass_service_role.example.role_arn
}
```

## Attributes Reference

* `role_arn` - The ARN of the role which is associated with the account.
