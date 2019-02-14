---
layout: "aws"
page_title: "AWS: aws_media_store_container_policy"
sidebar_current: "docs-aws-resource-media-store-container-policy"
description: |-
  Provides a MediaStore Container Policy.
---

# aws_media_store_container_policy

Provides a MediaStore Container Policy.

## Example Usage

```hcl
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_media_store_container" "example" {
  name = "example"
}

resource "aws_media_store_container_policy" "example" {
  container_name = "${aws_media_store_container.example.name}"

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [{
		"Sid": "MediaStoreFullAccess",
		"Action": [ "mediastore:*" ],
		"Principal": {"AWS" : "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"},
		"Effect": "Allow",
		"Resource": "arn:aws:mediastore:${data.aws_caller_identity.current.account_id}:${data.aws_region.current.name}:container/${aws_media_store_container.example.name}/*",
		"Condition": {
			"Bool": { "aws:SecureTransport": "true" }
		}
	}]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `container_name` - (Required) The name of the container.
* `policy` - (Required) The contents of the policy. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](/docs/providers/aws/guides/iam-policy-documents.html).

## Import

MediaStore Container Policy can be imported using the MediaStore Container Name, e.g.

```
$ terraform import aws_media_store_container_policy.example example
```
