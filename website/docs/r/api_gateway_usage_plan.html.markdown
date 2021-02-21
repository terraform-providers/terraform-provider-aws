---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_api_gateway_usage_plan"
description: |-
  Provides an API Gateway Usage Plan.
---

# Resource: aws_api_gateway_usage_plan

Provides an API Gateway Usage Plan.

## Example Usage

```hcl
resource "aws_api_gateway_rest_api" "example" {
  body = jsonencode({
    openapi = "3.0.1"
    info = {
      title   = "example"
      version = "1.0"
    }
    paths = {
      "/path1" = {
        get = {
          x-amazon-apigateway-integration = {
            httpMethod           = "GET"
            payloadFormatVersion = "1.0"
            type                 = "HTTP_PROXY"
            uri                  = "https://ip-ranges.amazonaws.com/ip-ranges.json"
          }
        }
      }
    }
  })

  name = "example"
}

resource "aws_api_gateway_deployment" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id

  triggers = {
    redeployment = sha1(jsonencode(aws_api_gateway_rest_api.example.body))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "development" {
  deployment_id = aws_api_gateway_deployment.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
  stage_name    = "development"
}

resource "aws_api_gateway_stage" "production" {
  deployment_id = aws_api_gateway_deployment.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
  stage_name    = "production"
}

resource "aws_api_gateway_usage_plan" "example" {
  name         = "my-usage-plan"
  description  = "my description"
  product_code = "MYCODE"

  api_stages {
    api_id = aws_api_gateway_rest_api.example.id
    stage  = aws_api_gateway_stage.development.stage_name
  }

  api_stages {
    api_id = aws_api_gateway_rest_api.example.id
    stage  = aws_api_gateway_stage.production.stage_name
  }

  quota_settings {
    limit  = 20
    offset = 2
    period = "WEEK"
  }

  throttle_settings {
    burst_limit = 5
    rate_limit  = 10
  }
}
```

## Argument Reference

The API Gateway Usage Plan argument layout is a structure composed of several sub-resources - these resources are laid out below.

### Top-Level Arguments

* `name` - (Required) The name of the usage plan.
* `description` - (Optional) The description of a usage plan.
* `api_stages` - (Optional) The associated [API stages](#api-stages-arguments) of the usage plan.
* `quota_settings` - (Optional) The [quota settings](#quota-settings-arguments) of the usage plan.
* `throttle_settings` - (Optional) The [throttling limits](#throttling-settings-arguments) of the usage plan.
* `product_code` - (Optional) The AWS Marketplace product identifier to associate with the usage plan as a SaaS product on AWS Marketplace.
* `tags` - (Optional) Key-value map of resource tags

#### Api Stages arguments

* `api_id` (Required) - API Id of the associated API stage in a usage plan.
* `stage` (Required) - API stage name of the associated API stage in a usage plan.

#### Quota Settings Arguments

* `limit` (Optional) - The maximum number of requests that can be made in a given time period.
* `offset` (Optional) - The number of requests subtracted from the given limit in the initial time period.
* `period` (Optional) - The time period in which the limit applies. Valid values are "DAY", "WEEK" or "MONTH".

#### Throttling Settings Arguments

* `burst_limit` (Optional) - The API request burst limit, the maximum rate limit over a time ranging from one to a few seconds, depending upon whether the underlying token bucket is at its full capacity.
* `rate_limit` (Optional) - The API request steady-state rate limit.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the API resource
* `name` - The name of the usage plan.
* `description` - The description of a usage plan.
* `api_stages` - The associated API stages of the usage plan.
* `quota_settings` - The quota of the usage plan.
* `throttle_settings` - The throttling limits of the usage plan.
* `product_code` - The AWS Marketplace product identifier to associate with the usage plan as a SaaS product on AWS Marketplace.
* `arn` - Amazon Resource Name (ARN)

## Import

AWS API Gateway Usage Plan can be imported using the `id`, e.g.

```sh
$ terraform import aws_api_gateway_usage_plan.myusageplan <usage_plan_id>
```
