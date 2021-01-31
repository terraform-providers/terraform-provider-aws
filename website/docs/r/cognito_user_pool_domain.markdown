---
subcategory: "Cognito"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_domain"
description: |-
  Provides a Cognito User Pool Domain resource.
---

# Resource: aws_cognito_user_pool_domain

Provides a Cognito User Pool Domain resource.

## Example Usage

### Amazon Cognito domain

```hcl
resource "aws_cognito_user_pool_domain" "main" {
  domain       = "example-domain"
  user_pool_id = aws_cognito_user_pool.example.id
}

resource "aws_cognito_user_pool" "example" {
  name = "example-pool"
}
```

### Custom Cognito domain

```hcl
resource "aws_cognito_user_pool_domain" "main" {
  domain          = "example-domain.example.com"
  certificate_arn = aws_acm_certificate.cert.arn
  user_pool_id    = aws_cognito_user_pool.example.id
}

resource "aws_cognito_user_pool" "example" {
  name = "example-pool"
}

data "aws_route53_zone" "example" {
  name = "example.com"
}

resource "aws_route53_record" "auth-cognito-A" {
  name    = aws_cognito_user_pool_domain.main.domain
  type    = "A"
  zone_id = data.aws_route53_zone.example.zone_id
  alias {
    evaluate_target_health = false
    name                   = aws_cognito_user_pool_domain.main.cloudfront_distribution_arn
    # This zone_id is fixed
    zone_id = "Z2FDTNDATAQYW2"
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The domain string.
* `user_pool_id` - (Required) The user pool ID.
* `certificate_arn` - (Optional) The ARN of an ISSUED ACM certificate in us-east-1 for a custom domain.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `aws_account_id` - The AWS account ID for the user pool owner.
* `cloudfront_distribution_arn` - The URL of the CloudFront distribution. This is required to generate the ALIAS `aws_route53_record`
* `s3_bucket` - The S3 bucket where the static files for this domain are stored.
* `version` - The app version.

## Import

Cognito User Pool Domains can be imported using the `domain`, e.g.

```
$ terraform import aws_cognito_user_pool_domain.main <domain>
```
