---
layout: "aws"
page_title: "AWS: aws_ecr_repository"
sidebar_current: "docs-aws-resource-ecr-repository"
description: |-
  Provides an EC2 Container Registry Repository.
---

# aws_ecr_repository

Provides an EC2 Container Registry Repository.

~> **NOTE on ECR Availability**: The EC2 Container Registry is not yet rolled out
in all regions - available regions are listed
[the AWS Docs](https://docs.aws.amazon.com/general/latest/gr/rande.html#ecr_region).

## Example Usage

```hcl
resource "aws_ecr_repository" "foo" {
  name = "bar"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the repository.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Full ARN of the repository.
* `name` - The name of the repository.
* `registry_id` - The registry ID where the repository was created.
* `repository_url` - The URL of the repository (in the form `aws_account_id.dkr.ecr.region.amazonaws.com/repositoryName`

## Timeouts

`aws_ecr_repository` provides the following [Timeouts](/docs/configuration/resources.html#timeouts)
configuration options:

- `delete` - (Default `20 minutes`) How long to wait for a repository to be deleted.

## Import

ECR Repositories can be imported using the `name`, e.g.

```
$ terraform import aws_ecr_repository.service test-service
```
