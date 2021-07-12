---
subcategory: "Elastic Beanstalk"
layout: "aws"
page_title: "AWS: aws_elastic_beanstalk_environment"
description: |-
  Retrieve information about an Elastic Beanstalk Environment
---

# Data Source: aws_elastic_beanstalk_environment

Retrieve information about an Elastic Beanstalk Environment.

## Example Usage

```hcl
data "aws_elastic_beanstalk_environment" "example" {
  application = "my-app"
}

output "name" {
  value = data.aws_elastic_beanstalk_environment.example.name
}
```

## Argument Reference

* `application` - (Required) The name of the application to which the environment belongs.

## Attributes Reference

* `name` - The name of the environment
