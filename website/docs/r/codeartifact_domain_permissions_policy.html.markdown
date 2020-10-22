---
subcategory: "CodeArtifact"
layout: "aws"
page_title: "AWS: aws_codeartifact_domain_permissions_policy"
description: |-
  Provides a CodeArtifact Domain Permissions Policy resource.
---

# Resource: aws_codeartifact_domain_permissions_policy

Provides a CodeArtifact Domains Permissions Policy Resource.

## Example Usage

```hcl
resource "aws_kms_key" "example" {
  description = "domain key"
}

resource "aws_codeartifact_domain" "example" {
  domain         = "example.com"
  encryption_key = aws_kms_key.example.arn
}

resource "aws_codeartifact_domain_permissions_policy" "test" {
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

* `domain` - (Required) The name of the domain on which to set the resource policy.
* `policy_document` - (Required) A JSON policy string to be set as the access control resource policy on the provided domain.
* `domain_owner` - (Optional) The account number of the AWS account that owns the domain.
* `policy_revision` - (Optional) The current revision of the resource policy to be set. This revision is used for optimistic locking, which prevents others from overwriting your changes to the domain's resource policy.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Name of Domain.
* `resource_arn` - The ARN of the resource associated with the resource policy.

## Import

CodeArtifact Domain Permissions Policies can be imported using the CodeArtifact Domain ARN, e.g.

```
$ terraform import aws_codeartifact_domain_permissions_policy.example arn:aws:codeartifact:us-west-2:012345678912:domain/tf-acc-test-1928056699409417367
```
