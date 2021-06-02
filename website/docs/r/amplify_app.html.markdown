---
subcategory: "Amplify Console"
layout: "aws"
page_title: "AWS: aws_amplify_app"
description: |-
  Provides an Amplify App resource.
---

# Resource: aws_amplify_app

Provides an Amplify App resource, a fullstack serverless app hosted on the [AWS Amplify Console](https://docs.aws.amazon.com/amplify/latest/userguide/welcome.html).

~> **Note:** When you create/update an Amplify App from Terraform, you may end up with the error "BadRequestException: You should at least provide one valid token" because of authentication issues. See the section "Repository with Tokens" below.

## Example Usage

```terraform
resource "aws_amplify_app" "example" {
  name       = "example"
  repository = "https://github.com/example/app"

  # The default build_spec added by the Amplify Console for React.
  build_spec = <<-EOT
    version: 0.1
    frontend:
      phases:
        preBuild:
          commands:
            - yarn install
        build:
          commands:
            - yarn run build
      artifacts:
        baseDirectory: build
        files:
          - '**/*'
      cache:
        paths:
          - node_modules/**/*
  EOT

  # The default rewrites and redirects added by the Amplify Console.
  custom_rule {
    source = "/<*>"
    status = "404"
    target = "/index.html"
  }

  environment_variables = {
    ENV = "test"
  }
}
```

### Repository with Tokens

If you create a new Amplify App with the `repository` argument, you also need to set `oauth_token` or `access_token` for authentication. For GitHub, get a [personal access token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line) and set `access_token` as follows:

```terraform
resource "aws_amplify_app" "example" {
  name       = "example"
  repository = "https://github.com/example/app"

  # GitHub personal access token
  access_token = "..."
}
```

You can omit `access_token` if you import an existing Amplify App created by the Amplify Console (using OAuth for authentication).

### Auto Branch Creation

```terraform
resource "aws_amplify_app" "example" {
  name = "example"

  enable_auto_branch_creation = true

  # The default patterns added by the Amplify Console.
  auto_branch_creation_patterns = [
    "*",
    "*/**",
  ]

  auto_branch_creation_config {
    # Enable auto build for the created branch.
    enable_auto_build = true
  }
}
```

### Basic Authorization

```terraform
resource "aws_amplify_app" "example" {
  name = "example"

  enable_basic_auth      = true
  basic_auth_credentials = base64encode("username1:password1")
}
```

### Rewrites and Redirects

```terraform
resource "aws_amplify_app" "example" {
  name = "example"

  # Reverse Proxy Rewrite for API requests
  # https://docs.aws.amazon.com/amplify/latest/userguide/redirects.html#reverse-proxy-rewrite
  custom_rule {
    source = "/api/<*>"
    status = "200"
    target = "https://api.example.com/api/<*>"
  }

  # Redirects for Single Page Web Apps (SPA)
  # https://docs.aws.amazon.com/amplify/latest/userguide/redirects.html#redirects-for-single-page-web-apps-spa
  custom_rule {
    source = "</^[^.]+$|\\.(?!(css|gif|ico|jpg|js|png|txt|svg|woff|ttf|map|json)$)([^.]+$)/>"
    status = "200"
    target = "/index.html"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name for an Amplify app.
* `access_token` - (Optional) The personal access token for a third-party source control system for an Amplify app. The personal access token is used to create a webhook and a read-only deploy key. The token is not stored.
* `auto_branch_creation_config` - (Optional) The automated branch creation configuration for an Amplify app. An `auto_branch_creation_config` block is documented below.
* `auto_branch_creation_patterns` - (Optional) The automated branch creation glob patterns for an Amplify app.
* `basic_auth_credentials` - (Optional) The credentials for basic authorization for an Amplify app.
* `build_spec` - (Optional) The [build specification](https://docs.aws.amazon.com/amplify/latest/userguide/build-settings.html) (build spec) for an Amplify app.
* `custom_rule` - (Optional) The custom rewrite and redirect rules for an Amplify app. A `custom_rule` block is documented below.
* `description` - (Optional) The description for an Amplify app.
* `enable_auto_branch_creation` - (Optional) Enables automated branch creation for an Amplify app.
* `enable_basic_auth` - (Optional) Enables basic authorization for an Amplify app. This will apply to all branches that are part of this app.
* `enable_branch_auto_build` - (Optional) Enables auto-building of branches for the Amplify App.
* `enable_branch_auto_deletion` - (Optional) Automatically disconnects a branch in the Amplify Console when you delete a branch from your Git repository.
* `environment_variables` - (Optional) The environment variables map for an Amplify app.
* `iam_service_role_arn` - (Optional) The AWS Identity and Access Management (IAM) service role for an Amplify app.
* `oauth_token` - (Optional) The OAuth token for a third-party source control system for an Amplify app. The OAuth token is used to create a webhook and a read-only deploy key. The OAuth token is not stored.
* `platform` - (Optional) The platform or framework for an Amplify app. Valid values: `WEB`.
* `repository` - (Optional) The repository for an Amplify app.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.


An `auto_branch_creation_config` block supports the following arguments:

* `basic_auth_credentials` - (Optional) The basic authorization credentials for the autocreated branch.
* `build_spec` - (Optional) The build specification (build spec) for the autocreated branch.
* `enable_auto_build` - (Optional) Enables auto building for the autocreated branch.
* `enable_basic_auth` - (Optional) Enables basic authorization for the autocreated branch.
* `enable_performance_mode` - (Optional) Enables performance mode for the branch.
* `enable_pull_request_preview` - (Optional) Enables pull request previews for the autocreated branch.
* `environment_variables` - (Optional) The environment variables for the autocreated branch.
* `framework` - (Optional) The framework for the autocreated branch.
* `pull_request_environment_name` - (Optional) The Amplify environment name for the pull request.
* `stage` - (Optional) Describes the current stage for the autocreated branch. Valid values: `PRODUCTION`, `BETA`, `DEVELOPMENT`, `EXPERIMENTAL`, `PULL_REQUEST`.

A `custom_rule` block supports the following arguments:

* `condition` - (Optional) The condition for a URL rewrite or redirect rule, such as a country code.
* `source` - (Required) The source pattern for a URL rewrite or redirect rule.
* `status` - (Optional) The status code for a URL rewrite or redirect rule. Valid values: `200`, `301`, `302`, `404`, `404-200`.
* `target` - (Required) The target pattern for a URL rewrite or redirect rule.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the Amplify app.
* `default_domain` - The default domain for the Amplify app.
* `id` - The unique ID of the Amplify app.
* `production_branch` - Describes the information about a production branch for an Amplify app. A `production_branch` block is documented below.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

A `production_branch` block supports the following attributes:

* `branch_name` - The branch name for the production branch.
* `last_deploy_time` - The last deploy time of the production branch.
* `status` - The status of the production branch.
* `thumbnail_url` - The thumbnail URL for the production branch.

## Import

Amplify App can be imported using Amplify App ID (appId), e.g.

```
$ terraform import aws_amplify_app.example d2ypk4k47z8u6
```

App ID can be obtained from App ARN (e.g. `arn:aws:amplify:us-east-1:12345678:apps/d2ypk4k47z8u6`).
