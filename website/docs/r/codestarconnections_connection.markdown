---
subcategory: "CodeStar Connections"
layout: "aws"
page_title: "AWS: aws_codestarconnections_connection"
description: |-
  Provides a CodeStar Connection
---

# Resource: aws_codestarconnections_connection

Provides a CodeStar Connection.

~> **NOTE:** The `aws_codestarconnections_connection` resource is created in the state `PENDING`. Authentication with the connection provider must be completed in the AWS Console.

## Example Usage

```hcl
resource "aws_codestarconnections_connection" "example" {
  name          = "example-connection"
  provider_type = "Bitbucket"
}

resource "aws_codepipeline" "example" {
  name     = "tf-test-pipeline"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    # ...
  }

  stage {
    name = "Source"
    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["source_output"]
      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.example.arn
        FullRepositoryId = "my-organization/test"
        BranchName       = "main"
      }
    }
  }

  stage {
    name = "Build"
    action {
      # ...
    }
  }

  stage {
    name = "Deploy"
    action {
      # ...
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the connection to be created. The name must be unique in the calling AWS account. Changing `name` will create a new resource.
* `provider_type` - (Required) The name of the external provider where your third-party code repository is configured. Valid values are `Bitbucket`, `GitHub`, or `GitHubEnterpriseServer`. Changing `provider_type` will create a new resource.
* `tags` - (Optional) Map of key-value resource tags to associate with the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The codestar connection ARN.
* `arn` - The codestar connection ARN.
* `connection_status` - The codestar connection status. Possible values are `PENDING`, `AVAILABLE` and `ERROR`.

## Import

CodeStar connections can be imported using the ARN, e.g.

```
$ terraform import aws_codestarconnections_connection.test-connection arn:aws:codestar-connections:us-west-1:0123456789:connection/79d4d357-a2ee-41e4-b350-2fe39ae59448
```
