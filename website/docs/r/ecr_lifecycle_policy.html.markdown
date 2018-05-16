---
layout: "aws"
page_title: "AWS: aws_ecr_lifecycle_policy"
sidebar_current: "docs-aws-resource-ecr-lifecycle-policy"
description: |-
  Provides an ECR Lifecycle Policy.
---

# aws_ecr_lifecycle_policy

Provides an ECR lifecycle policy.

## Example Usage

### Policy on untagged image

```hcl
resource "aws_ecr_repository" "foo" {
  name = "bar"
}

resource "aws_ecr_lifecycle_policy" "foopolicy" {
  repository = "${aws_ecr_repository.foo.name}"

  policy = <<EOF
{
    "rules": [
        {
            "rulePriority": 1,
            "description": "Expire images older than 14 days",
            "selection": {
                "tagStatus": "untagged",
                "countType": "sinceImagePushed",
                "countUnit": "days",
                "countNumber": 14
            },
            "action": {
                "type": "expire"
            }
        }
    ]
}
EOF
}
```

### Policy on tagged image

```hcl
resource "aws_ecr_repository" "foo" {
  name = "bar"
}

resource "aws_ecr_lifecycle_policy" "foopolicy" {
  repository = "${aws_ecr_repository.foo.name}"

  policy = <<EOF
{
    "rules": [
        {
            "rulePriority": 1,
            "description": "Keep last 30 images",
            "selection": {
                "tagStatus": "tagged",
                "tagPrefixList": ["v"],
                "countType": "imageCountMoreThan",
                "countNumber": 30
            },
            "action": {
                "type": "expire"
            }
        }
    ]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `repository` - (Required) Name of the repository to apply the policy.
* `policy` - (Required) The policy document. This is a JSON formatted string. See more details about [Policy Parameters](http://docs.aws.amazon.com/AmazonECR/latest/userguide/LifecyclePolicies.html#lifecycle_policy_parameters) in the official AWS docs.

## Attributes Reference

The following attributes are exported:

* `repository` - The name of the repository.
* `registry_id` - The registry ID where the repository was created.

## Import

ECR Lifecycle Policy can be imported using the name of the repository, e.g.

```
$ terraform import aws_ecr_lifecycle_policy.example tf-example
```
