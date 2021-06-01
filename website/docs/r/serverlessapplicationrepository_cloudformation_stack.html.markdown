---
subcategory: "Serverless Application Repository"
layout: "aws"
page_title: "AWS: aws_serverlessapplicationrepository_cloudformation_stack"
description: |-
  Deploys an Application CloudFormation Stack from the Serverless Application Repository.
---

# Resource: aws_serverlessapplicationrepository_cloudformation_stack

Deploys an Application CloudFormation Stack from the Serverless Application Repository.

## Example Usage

```hcl
resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = "postgres-rotator"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "func-postgres-rotator"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the stack to create. The resource deployed in AWS will be prefixed with `serverlessrepo-`
* `application_id` - (Required) The ARN of the application from the Serverless Application Repository.
* `capabilities` - (Required) A list of capabilities. Valid values are `CAPABILITY_IAM`, `CAPABILITY_NAMED_IAM`, `CAPABILITY_RESOURCE_POLICY`, or `CAPABILITY_AUTO_EXPAND`
* `parameters` - (Optional) A map of Parameter structures that specify input parameters for the stack.
* `semantic_version` - (Optional) The version of the application to deploy. If not supplied, deploys the latest version.
* `tags` - (Optional) A list of tags to associate with this stack.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A unique identifier of the stack.
* `outputs` - A map of outputs from the stack.

## Import

Serverless Application Repository Stack can be imported using the CloudFormation Stack name (with or without the `serverlessrepo-` prefix) or the CloudFormation Stack ID, e.g.

```
$ terraform import aws_serverlessapplicationrepository_cloudformation_stack.example serverlessrepo-postgres-rotator
```
