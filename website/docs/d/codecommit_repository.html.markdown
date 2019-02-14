---
layout: "aws"
page_title: "AWS: aws_codecommit_repository"
sidebar_current: "docs-aws-datasource-codecommit-repository"
description: |-
  Provides details about CodeCommit Repository.
---

# Data Source: aws_codecommit_repository

The CodeCommit Repository data source allows the ARN, Repository ID, Repository URL for HTTP and Repository URL for SSH to be retrieved for an CodeCommit repository.

## Example Usage

```hcl
data "aws_codecommit_repository" "test" {
  repository_name = "MyTestRepository"
}
```

## Argument Reference

The following arguments are supported:

* `repository_name` - (Required) The name for the repository. This needs to be less than 100 characters.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `repository_id` - The ID of the repository
* `arn` - The ARN of the repository
* `clone_url_http` - The URL to use for cloning the repository over HTTPS.
* `clone_url_ssh` - The URL to use for cloning the repository over SSH.
