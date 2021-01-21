---
subcategory: "ECR"
layout: "aws"
page_title: "AWS: aws_ecr_repository_policy"
description: |-
  Provides an Elastic Container Registry Repository Policy.
---

# Resource: aws_ecr_repository_policy

Provides an Elastic Container Registry Repository Policy.

Note that currently only one policy may be applied to a repository.

## Example Usage

```hcl
resource "aws_ecr_repository" "foo" {
  name = "bar"
}

resource "aws_ecr_repository_policy" "foopolicy" {
  repository = aws_ecr_repository.foo.name

  policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "new policy",
            "Effect": "Allow",
            "Principal": "*",
            "Action": [
                "ecr:GetDownloadUrlForLayer",
                "ecr:BatchGetImage",
                "ecr:BatchCheckLayerAvailability",
                "ecr:PutImage",
                "ecr:InitiateLayerUpload",
                "ecr:UploadLayerPart",
                "ecr:CompleteLayerUpload",
                "ecr:DescribeRepositories",
                "ecr:GetRepositoryPolicy",
                "ecr:ListImages",
                "ecr:DeleteRepository",
                "ecr:BatchDeleteImage",
                "ecr:SetRepositoryPolicy",
                "ecr:DeleteRepositoryPolicy"
            ]
        }
    ]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `repository` - (Required) Name of the repository to apply the policy.
* `policy` - (Required) The policy document. This is a JSON formatted string. For more information about building IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `repository` - The name of the repository.
* `registry_id` - The registry ID where the repository was created.

## Import

ECR Repository Policy can be imported using the repository name, e.g.

```
$ terraform import aws_ecr_repository_policy.example example
```
