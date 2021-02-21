---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_api_gateway_deployment"
description: |-
  Manages an API Gateway REST Deployment.
---

# Resource: aws_api_gateway_deployment

Manages an API Gateway REST Deployment. A deployment is a snapshot of the REST API configuration. The deployment can then be published to callable endpoints via the [`aws_api_gateway_stage` resource](api_gateway_stage.html) and optionally managed further with the [`aws_api_gateway_base_path_mapping` resource](api_gateway_base_path_mapping.html), [`aws_api_gateway_domain_name` resource](api_gateway_domain_name.html), and [`aws_api_method_settings` resource](api_gateway_method_settings.html). For more information, see the [API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-deploy-api.html).

To properly capture all REST API configuration in a deployment, this resource must have dependencies on all prior Terraform resources that manage resources/paths, methods, integrations, etc.

* For REST APIs that are configured via OpenAPI specification ([`aws_api_gateway_rest_api` resource](api_gateway_rest_api.html) `body` argument), no special dependency setup is needed beyond referencing the  `id` attribute of that resource unless additional Terraform resources have further customized the REST API.
* When the REST API configuration involves other Terraform resources ([`aws_api_gateway_integration` resource](api_gateway_integration.html), etc.), the dependency setup can be done with implicit resource references in the `triggers` argument or explicit resource references using the [resource `depends_on` meta-argument](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html). The `triggers` argument should be preferred over `depends_on`, since `depends_on` can only capture dependency ordering and will not cause the resource to recreate (redeploy the REST API) with upstream configuration changes.

!> **WARNING:** It is recommended to use the [`aws_api_gateway_stage` resource](api_gateway_stage.html) instead of managing an API Gateway Stage via the `stage_name` argument of this resource. When this resource is recreated (REST API redeployment) with the `stage_name` configured, the stage is deleted and recreated. This will cause a temporary service interruption, increase Terraform plan differences, and can require a second Terraform apply to recreate any downstream stage configuration such as associated `aws_api_method_settings` resources.

~> **NOTE:** It is recommended to enable the [resource `lifecycle` configuration block `create_before_destroy` argument](https://www.terraform.io/docs/configuration/resources.html#create_before_destroy) in this resource configuration to properly order redeployments in Terraform. Without enabling `create_before_destroy`, API Gateway can return errors such as `BadRequestException: Active stages pointing to this deployment must be moved or deleted` on recreation.

## Example Usage

### OpenAPI Specification

An end-to-end example of a REST API configured with OpenAPI can be found in the [`/examples/api-gateway-rest-api-openapi` directory within the GitHub repository](https://github.com/hashicorp/terraform-provider-aws/tree/main/examples/api-gateway-rest-api-openapi).

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

resource "aws_api_gateway_stage" "example" {
  deployment_id = aws_api_gateway_deployment.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
  stage_name    = "example"
}
```

### Terraform Resources

```hcl
resource "aws_api_gateway_rest_api" "example" {
  name = "example"
}

resource "aws_api_gateway_resource" "example" {
  parent_id   = aws_api_gateway_rest_api.example.root_resource_id
  path_part   = "example"
  rest_api_id = aws_api_gateway_rest_api.example.id
}

resource "aws_api_gateway_method" "example" {
  authorization = "NONE"
  http_method   = "GET"
  resource_id   = aws_api_gateway_resource.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
}

resource "aws_api_gateway_integration" "example" {
  http_method = aws_api_gateway_method.example.http_method
  resource_id = aws_api_gateway_resource.example.id
  rest_api_id = aws_api_gateway_rest_api.example.id
  type        = "MOCK"
}

resource "aws_api_gateway_deployment" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id

  triggers = {
    # NOTE: The configuration below will satisfy ordering considerations,
    #       but not pick up all future REST API changes. More advanced patterns
    #       are possible, such as using the filesha1() function against the
    #       Terraform configuration file(s) or removing the .id references to
    #       calculate a hash against whole resources. Be aware that using whole
    #       resources will show a difference after the initial implementation.
    #       It will stabilize to only change when resources change afterwards.
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.example.id,
      aws_api_gateway_method.example.id,
      aws_api_gateway_integration.example.id,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "example" {
  deployment_id = aws_api_gateway_deployment.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
  stage_name    = "example"
}
```

## Argument Reference

The following arguments are supported:

* `rest_api_id` - (Required) REST API identifier.
* `description` - (Optional) Description of the deployment
* `stage_name` - (Optional) Name of the stage to create with this deployment. If the specified stage already exists, it will be updated to point to the new deployment. It is recommended to use the [`aws_api_gateway_stage` resource](api_gateway_stage.html) instead to manage stages.
* `stage_description` - (Optional) Description to set on the stage managed by the `stage_name` argument.
* `triggers` - (Optional) Map of arbitrary keys and values that, when changed, will trigger a redeployment. To force a redeployment without changing these keys/values, use the [`terraform taint` command](https://www.terraform.io/docs/commands/taint.html).
* `variables` - (Optional) Map to set on the stage managed by the `stage_name` argument.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the deployment
* `invoke_url` - The URL to invoke the API pointing to the stage,
  e.g. `https://z4675bid1j.execute-api.eu-west-2.amazonaws.com/prod`
* `execution_arn` - The execution ARN to be used in [`lambda_permission`](/docs/providers/aws/r/lambda_permission.html)'s `source_arn`
  when allowing API Gateway to invoke a Lambda function,
  e.g. `arn:aws:execute-api:eu-west-2:123456789012:z4675bid1j/prod`
* `created_date` - The creation date of the deployment
