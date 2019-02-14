---
layout: "aws"
page_title: "AWS: aws_codebuild_webhook"
sidebar_current: "docs-aws-resource-codebuild-webhook"
description: |-
  Provides a CodeBuild Webhook resource.
---

# aws_codebuild_webhook

Manages a CodeBuild webhook, which is an endpoint accepted by the CodeBuild service to trigger builds from source code repositories. Depending on the source type of the CodeBuild project, the CodeBuild service may also automatically create and delete the actual repository webhook as well.

## Example Usage

### Bitbucket and GitHub

When working with [Bitbucket](https://bitbucket.org) and [GitHub](https://github.com) source CodeBuild webhooks, the CodeBuild service will automatically create (on `aws_codebuild_webhook` resource creation) and delete (on `aws_codebuild_webhook` resource deletion) the Bitbucket/GitHub repository webhook using its granted OAuth permissions. This behavior cannot be controlled by Terraform.

~> **Note:** The AWS account that Terraform uses to create this resource *must* have authorized CodeBuild to access Bitbucket/GitHub's OAuth API in each applicable region. This is a manual step that must be done *before* creating webhooks with this resource. If OAuth is not configured, AWS will return an error similar to `ResourceNotFoundException: Could not find access token for server type github`. More information can be found in the CodeBuild User Guide for [Bitbucket](https://docs.aws.amazon.com/codebuild/latest/userguide/sample-bitbucket-pull-request.html) and [GitHub](https://docs.aws.amazon.com/codebuild/latest/userguide/sample-github-pull-request.html).

~> **Note:** Further managing the automatically created Bitbucket/GitHub webhook with the `bitbucket_hook`/`github_repository_webhook` resource is only possible with importing that resource after creation of the `aws_codebuild_webhook` resource. The CodeBuild API does not ever provide the `secret` attribute for the `aws_codebuild_webhook` resource in this scenario.

```hcl
resource "aws_codebuild_webhook" "example" {
  project_name = "${aws_codebuild_project.example.name}"
}
```

### GitHub Enterprise

When working with [GitHub Enterprise](https://enterprise.github.com/) source CodeBuild webhooks, the GHE repository webhook must be separately managed (e.g. manually or with the `github_repository_webhook` resource).

More information creating webhooks with GitHub Enterprise can be found in the [CodeBuild User Guide](https://docs.aws.amazon.com/codebuild/latest/userguide/sample-github-enterprise.html).

```hcl
resource "aws_codebuild_webhook" "example" {
  project_name = "${aws_codebuild_project.example.name}"
}

resource "github_repository_webhook" "example" {
  active     = true
  events     = ["push"]
  name       = "example"
  repository = "${github_repository.example.name}"

  configuration {
    url          = "${aws_codebuild_webhook.example.payload_url}"
    secret       = "${aws_codebuild_webhook.example.secret}"
    content_type = "json"
    insecure_ssl = false
  }
}
```

## Argument Reference

The following arguments are supported:

* `project_name` - (Required) The name of the build project.
* `branch_filter` - (Optional) A regular expression used to determine which branches get built. Default is all branches are built.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the build project.
* `payload_url` - The CodeBuild endpoint where webhook events are sent.
* `secret` - The secret token of the associated repository. Not returned by the CodeBuild API for all source types.
* `url` - The URL to the webhook.

~> **Note:** The `secret` attribute is only set on resource creation, so if the secret is manually rotated, terraform will not pick up the change on subsequent runs.  In that case, the webhook resource should be tainted and re-created to get the secret back in sync.

## Import

CodeBuild Webhooks can be imported using the CodeBuild Project name, e.g.

```
$ terraform import aws_codebuild_webhook.example MyProjectName
```
