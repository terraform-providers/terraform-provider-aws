---
layout: "aws"
page_title: "AWS: aws_elastic_beanstalk_application"
sidebar_current: "docs-aws-datasource-elastic-beanstalk-application"
description: |-
  Retrieve information about an Elastic Beanstalk Application
---

# Data Source: aws_elastic_beanstalk_application

Retrieve information about an Elastic Beanstalk Application.

## Example Usage

```hcl
data "aws_elastic_beanstalk_application" "example" {
  name = "example"
}

output "arn" {
  value = "${data.aws_elastic_beanstalk_application.example.arn}"
}

output "description" {
  value = "${data.aws_elastic_beanstalk_application.example.description}"
}
```

## Argument Reference

* `name` - (Required) The name of the application

## Attributes Reference

* `id` - The name of the application
* `arn` - The Amazon Resource Name (ARN) of the application.
* `description` - Short description of the application

Application version lifecycle (`appversion_lifecycle`) supports the nested attribute containing.

* `service_role` - The ARN of an IAM service role under which the application version is deleted.  Elastic Beanstalk must have permission to assume this role.
* `max_count` - The maximum number of application versions to retain.
* `max_age_in_days` - The number of days to retain an application version.
* `delete_source_from_s3` - Specifies whether delete a version's source bundle from S3 when the application version is deleted.
