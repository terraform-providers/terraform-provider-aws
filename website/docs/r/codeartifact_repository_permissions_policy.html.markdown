---
subcategory: "CodeArtifact"
layout: "aws"
page_title: "AWS: aws_codeartifact_repository_permissions_policy"
description: |-
  Provides a CodeArtifact Repository Permissions Policy resource.
---

# Resource: aws_codeartifact_repository_permissions_policy

Provides a CodeArtifact Repostory Permissions Policy Resource.

## Example Usage

```hcl
resource "aws_kms_key" "example" {
  description = "domain key"
}

resource "aws_codeartifact_domain" "example" {
  domain         = "example.com"
  encryption_key = aws_kms_key.example.arn
}

resource "aws_codeartifact_repository" "example" {
  repository = "example"
  domain     = aws_codeartifact_domain.example.domain
}

resource "aws_codeartifact_repository_permissions_policy" "example" {
  repository      = aws_codeartifact_repository.example.repsitory
  domain          = aws_codeartifact_domain.example.domain
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "codeartifact:CreateRepository",
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.example.arn}"
        }
    ]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `repository` - (Required) The name of the repository to set the resource policy on.
* `domain` - (Required) The name of the domain on which to set the resource policy.
* `policy_document` - (Required) A JSON policy string to be set as the access control resource policy on the provided domain.
* `domain_owner` - (Optional) The account number of the AWS account that owns the domain.
* `policy_revision` - (Optional) The current revision of the resource policy to be set. This revision is used for optimistic locking, which prevents others from overwriting your changes to the domain's resource policy.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the resource associated with the resource policy.
* `resource_arn` - The ARN of the resource associated with the resource policy.

## Import

CodeArtifact Repository Permissions Policies can be imported using the CodeArtifact Repository ARN, e.g.

```
$ terraform import aws_codeartifact_repository_permissions_policy.example arn:aws:codeartifact:us-west-2:012345678912:repository/tf-acc-test-6968272603913957763/tf-acc-test-6968272603913957763
```
