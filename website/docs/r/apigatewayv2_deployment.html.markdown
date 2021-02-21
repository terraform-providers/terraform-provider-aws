---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_deployment"
description: |-
  Manages an Amazon API Gateway Version 2 deployment.
---

# Resource: aws_apigatewayv2_deployment

Manages an Amazon API Gateway Version 2 deployment.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api.html).

-> **Note:** Creating a deployment for an API requires at least one `aws_apigatewayv2_route` resource associated with that API. To avoid race conditions when all resources are being created together, you need to add implicit resource references via the `triggers` argument or explicit resource references using the [resource `depends_on` meta-argument](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html).

-> It is recommended to enable the [resource `lifecycle` configuration block `create_before_destroy` argument](https://www.terraform.io/docs/configuration/resources.html#create_before_destroy) in this resource configuration to properly order redeployments in Terraform.

## Example Usage

### Basic

```hcl
resource "aws_apigatewayv2_deployment" "example" {
  api_id      = aws_apigatewayv2_route.example.api_id
  description = "Example deployment"

  lifecycle {
    create_before_destroy = true
  }
}
```

### Redeployment Triggers

-> **NOTE:** This is an optional and Terraform 0.12 (or later) advanced configuration that shows calculating a hash of the API's Terraform resources to determine changes that should trigger a new deployment. This value will change after the first Terraform apply of new resources, triggering an immediate redeployment, however it will stabilize afterwards except for resource changes. The `triggers` map can also be configured in other, more complex ways to fit the environment, avoiding the immediate redeployment issue.

```hcl
resource "aws_apigatewayv2_deployment" "example" {
  api_id      = aws_apigatewayv2_api.example.id
  description = "Example deployment"

  triggers = {
    redeployment = sha1(join(",", list(
      jsonencode(aws_apigatewayv2_integration.example),
      jsonencode(aws_apigatewayv2_route.example),
    )))
  }

  lifecycle {
    create_before_destroy = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `description` - (Optional) The description for the deployment resource. Must be less than or equal to 1024 characters in length.
* `triggers` - (Optional) A map of arbitrary keys and values that, when changed, will trigger a redeployment. To force a redeployment without changing these keys/values, use the [`terraform taint` command](https://www.terraform.io/docs/commands/taint.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The deployment identifier.
* `auto_deployed` - Whether the deployment was automatically released.

## Import

`aws_apigatewayv2_deployment` can be imported by using the API identifier and deployment identifier, e.g.

```
$ terraform import aws_apigatewayv2_deployment.example aabbccddee/1122334
```

The `triggers` argument cannot be imported.
