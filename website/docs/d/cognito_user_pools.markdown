---
layout: "aws"
page_title: "AWS: aws_cognito_user_pools"
sidebar_current: "docs-aws-cognito-user-pools"
description: |-
  Get list of cognito user pools.
---

# Data Source: aws_cognito_user_pools

Use this data source to get a list of cognito user pools.

## Example Usage

```hcl
data "aws_api_gateway_rest_api" "selected" {
  name = "${var.api_gateway_name}"
}

data "aws_cognito_user_pools" "selected" {
  name = "${var.cognito_user_pool_name}"
}

resource "aws_api_gateway_authorizer" "cognito" {
  name          = "cognito"
  type          = "COGNITO_USER_POOLS"
  rest_api_id   = "${data.aws_api_gateway_rest_api.selected.id}"
  provider_arns = ["${data.aws_cognito_user_pools.selected.arns}"]
}
```

## Argument Reference

* `name` - (required) Name of the cognito user pools. Name is not a unique attribute for cognito user pool, so multiple pools might be returned with given name.


## Attributes Reference

* `ids` - The list of cognito user pool ids.
