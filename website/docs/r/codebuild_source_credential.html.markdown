---
subcategory: "CodeBuild"
layout: "aws"
page_title: "AWS: aws_codebuild_source_credential"
description: |-
  Provides a CodeBuild Source Credential resource.
---

# Resource: aws_codebuild_source_credential

Provides a CodeBuild Source Credentials Resource.

## Example Usage

```hcl
resource "aws_codebuild_source_credential" "example" {
  auth_type   = "PERSONAL_ACCESS_TOKEN"
  server_type = "GITHUB"
  token       = "example"
}
```

### Bitbucket Server Usage

```hcl
resource "aws_codebuild_source_credential" "example" {
  auth_type   = "BASIC_AUTH"
  server_type = "BITBUCKET"
  token       = "example"
  user_name   = "test-user"
}
```

## Argument Reference

The following arguments are supported:

* `auth_type` - (Required) The type of authentication used to connect to a GitHub, GitHub Enterprise, or Bitbucket repository. An OAUTH connection is not supported by the API.
* `server_type` - (Required) The source provider used for this project.
* `token` - (Required) For `GitHub` or `GitHub Enterprise`, this is the personal access token. For `Bitbucket`, this is the app password.
* `user_name` - (Optional) The Bitbucket username when the authType is `BASIC_AUTH`. This parameter is not valid for other types of source providers or connections.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of Source Credential.
* `arn` - The ARN of Source Credential.

## Import

CodeBuild Source Credential can be imported using the CodeBuild Source Credential arn, e.g.

```
$ terraform import aws_codebuild_source_credential.example arn:aws:codebuild:us-west-2:123456789:token:github
```
