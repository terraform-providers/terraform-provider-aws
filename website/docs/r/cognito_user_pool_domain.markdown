---
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_domain"
sidebar_current: "docs-aws-resource-cognito-user-pool-domain"
description: |-
  Provides a Cognito User Pool Domain resource.
---

# aws_cognito_user_pool_domain

Provides a Cognito User Pool Domain resource.

## Example Usage

```hcl
resource "aws_cognito_user_pool_domain" "main" {
  domain = "example-domain"
  user_pool_id = "${aws_cognito_user_pool.example.id}"
}

resource "aws_cognito_user_pool" "example" {
  name = "example-pool"
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The domain string.
* `user_pool_id` - (Required) The user pool ID.

## Attribute Reference

The following attributes are exported:

* `aws_account_id` - The AWS account ID for the user pool owner.
* `cloudfront_distribution_arn` - The ARN of the CloudFront distribution.
* `s3_bucket` - The S3 bucket where the static files for this domain are stored.
* `version` - The app version.
