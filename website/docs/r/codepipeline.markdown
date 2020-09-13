---
subcategory: "CodePipeline"
layout: "aws"
page_title: "AWS: aws_codepipeline"
description: |-
  Provides a CodePipeline
---

# Resource: aws_codepipeline

Provides a CodePipeline.

## Example Usage

```hcl
resource "aws_codepipeline" "codepipeline" {
  name     = "tf-test-pipeline"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.codepipeline_bucket.bucket
    type     = "S3"

    encryption_key {
      id   = data.aws_kms_alias.s3kmskey.arn
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["source_output"]

      configuration = {
        Owner      = "my-organization"
        Repo       = "test"
        Branch     = "master"
        OAuthToken = var.github_token
      }
    }
  }

  stage {
    name = "Build"

    action {
      name             = "Build"
      category         = "Build"
      owner            = "AWS"
      provider         = "CodeBuild"
      input_artifacts  = ["source_output"]
      output_artifacts = ["build_output"]
      version          = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }

  stage {
    name = "Deploy"

    action {
      name            = "Deploy"
      category        = "Deploy"
      owner           = "AWS"
      provider        = "CloudFormation"
      input_artifacts = ["build_output"]
      version         = "1"

      configuration = {
        ActionMode     = "REPLACE_ON_FAILURE"
        Capabilities   = "CAPABILITY_AUTO_EXPAND,CAPABILITY_IAM"
        OutputFileName = "CreateStackOutput.json"
        StackName      = "MyStack"
        TemplatePath   = "build_output::sam-templated.yaml"
      }
    }
  }
}

resource "aws_s3_bucket" "codepipeline_bucket" {
  bucket = "test-bucket"
  acl    = "private"
}

resource "aws_iam_role" "codepipeline_role" {
  name = "test-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codepipeline.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "codepipeline_policy" {
  name = "codepipeline_policy"
  role = aws_iam_role.codepipeline_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect":"Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketVersioning",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.codepipeline_bucket.arn}",
        "${aws_s3_bucket.codepipeline_bucket.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "codebuild:BatchGetBuilds",
        "codebuild:StartBuild"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

data "aws_kms_alias" "s3kmskey" {
  name = "alias/myKmsKey"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the pipeline.
* `role_arn` - (Required) A service role Amazon Resource Name (ARN) that grants AWS CodePipeline permission to make calls to AWS services on your behalf.
* `artifact_store` (Required) One or more artifact_store blocks. Artifact stores are documented below.
* `stage` (Minimum of at least two `stage` blocks is required) A stage block. Stages are documented below.
* `tags` - (Optional) A map of tags to assign to the resource.


An `artifact_store` block supports the following arguments:

* `location` - (Required) The location where AWS CodePipeline stores artifacts for a pipeline; currently only `S3` is supported.
* `type` - (Required) The type of the artifact store, such as Amazon S3
* `encryption_key` - (Optional) The encryption key block AWS CodePipeline uses to encrypt the data in the artifact store, such as an AWS Key Management Service (AWS KMS) key. If you don't specify a key, AWS CodePipeline uses the default key for Amazon Simple Storage Service (Amazon S3). An `encryption_key` block is documented below.
* `region` - (Optional) The region where the artifact store is located. Required for a cross-region CodePipeline, do not provide for a single-region CodePipeline.

An `encryption_key` block supports the following arguments:

* `id` - (Required) The KMS key ARN or ID
* `type` - (Required) The type of key; currently only `KMS` is supported

A `stage` block supports the following arguments:

* `name` - (Required) The name of the stage.
* `action` - (Required) The action(s) to include in the stage. Defined as an `action` block below

An `action` block supports the following arguments:

* `category` - (Required) A category defines what kind of action can be taken in the stage, and constrains the provider type for the action. Possible values are `Approval`, `Build`, `Deploy`, `Invoke`, `Source` and `Test`.
* `owner` - (Required) The creator of the action being called. Possible values are `AWS`, `Custom` and `ThirdParty`.
* `name` - (Required) The action declaration's name.
* `provider` - (Required) The provider of the service being called by the action. Valid providers are determined by the action category. For example, an action in the Deploy category type might have a provider of AWS CodeDeploy, which would be specified as CodeDeploy.
* `version` - (Required) A string that identifies the action type.
* `configuration` - (Optional) A map of the action declaration's configuration. Configurations options for action types and providers can be found in the [Pipeline Structure Reference](http://docs.aws.amazon.com/codepipeline/latest/userguide/reference-pipeline-structure.html#action-requirements) and [Action Structure Reference](https://docs.aws.amazon.com/codepipeline/latest/userguide/action-reference.html) documentation.
* `input_artifacts` - (Optional) A list of artifact names to be worked on.
* `output_artifacts` - (Optional) A list of artifact names to output. Output artifact names must be unique within a pipeline.
* `role_arn` - (Optional) The ARN of the IAM service role that will perform the declared action. This is assumed through the roleArn for the pipeline.
* `run_order` - (Optional) The order in which actions are run.
* `region` - (Optional) The region in which to run the action.
* `namespace` - (Optional) The namespace all output variables will be accessed from.

~> **Note:** The input artifact of an action must exactly match the output artifact declared in a preceding action, but the input artifact does not have to be the next action in strict sequence from the action that provided the output artifact. Actions in parallel can declare different output artifacts, which are in turn consumed by different following actions.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The codepipeline ID.
* `arn` - The codepipeline ARN.

## Import

CodePipelines can be imported using the name, e.g.

```
$ terraform import aws_codepipeline.foo example
```
