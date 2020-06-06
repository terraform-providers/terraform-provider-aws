---
subcategory: "CodeCommit"
layout: "aws"
page_title: "AWS: aws_codecommit_repository"
description: |-
  Provides a CodeCommit Repository Resource.
---

# Resource: aws_codecommit_repository

Provides a CodeCommit Repository Resource.

## Example Usage

```hcl
resource "aws_codecommit_repository" "test" {
  repository_name = "MyTestRepository"
  description     = "This is the Sample App Repository"
}
```

## Argument Reference

The following arguments are supported:

* `repository_name` - (Required) The name for the repository. This needs to be less than 100 characters.
* `description` - (Optional) The description of the repository. This needs to be less than 1000 characters
* `default_branch` - (Optional) The default branch of the repository. The branch specified here needs to exist.
* `tags` - (Optional) Key-value map of resource tags

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `repository_id` - The ID of the repository
* `arn` - The ARN of the repository
* `clone_url_http` - The URL to use for cloning the repository over HTTPS.
* `clone_url_ssh` - The URL to use for cloning the repository over SSH.

## Import

Codecommit repository can be imported using repository name, e.g.

```
$ terraform import aws_codecommit_repository.imported ExistingRepo
```
