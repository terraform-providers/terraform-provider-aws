---
layout: "aws"
page_title: "AWS: aws_ecr_repository"
sidebar_current: "docs-aws-datasource-ecr-repository"
description: |-
    Provides details about an ECR Repository
---

# Data Source: aws_ecr_repository

The ECR Repository data source allows the ARN, Repository URI and Registry ID to be retrieved for an ECR repository.

## Example Usage

```hcl
data "aws_ecr_repository" "service" {
  name = "ecr-repository"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the ECR Repository.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Full ARN of the repository.
* `registry_id` - The registry ID where the repository was created.
* `repository_url` - The URL of the repository (in the form `aws_account_id.dkr.ecr.region.amazonaws.com/repositoryName`).
* `tags` - A mapping of tags assigned to the resource.
