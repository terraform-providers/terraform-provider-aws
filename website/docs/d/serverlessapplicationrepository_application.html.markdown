---
subcategory: "Serverless Application Repository"
layout: "aws"
page_title: "AWS: aws_serverlessapplicationrepository_application"
description: |-
  Get information on a AWS Serverless Application Repository application
---

# Data Source: aws_serverlessapplicationrepository_application

Use this data source to get information about an AWS Serverless Application Repository application. For example, this can be used to determine the required `capabilities` for an application.

## Example Usage

```hcl
data "aws_serverlessapplicationrepository_application" "example" {
  application_id = "arn:aws:serverlessrepo:us-east-1:123456789012:applications/ExampleApplication"
}

resource "aws_serverlessapplicationrepository_cloudformation_stack" "example" {
  name             = "Example"
  application_id   = data.aws_serverlessapplicationrepository_application.example.application_id
  semantic_version = data.aws_serverlessapplicationrepository_application.example.semantic_version
  capabilities     = data.aws_serverlessapplicationrepository_application.example.required_capabilities
}
```

## Argument Reference

* `application_id` - (Required) The ARN of the application.
* `semantic_version` - (Optional) The requested version of the application. By default, retrieves the latest version.

## Attributes Reference

* `application_id` - The ARN of the application.
* `name` - The name of the application.
* `required_capabilities` - A list of capabilities describing the permissions needed to deploy the application.
* `source_code_url` - A URL pointing to the source code of the application version.
* `template_url` - A URL pointing to the Cloud Formation template for the application version.
