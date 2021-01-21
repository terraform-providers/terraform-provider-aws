---
subcategory: "CodeArtifact"
layout: "aws"
page_title: "AWS: aws_codeartifact_authorization_token"
description: |-
    Provides details about a CodeArtifact Authorization Token
---

# Data Source: aws_codeartifact_authorization_token

The CodeArtifact Authorization Token data source generates a temporary authentication token for accessing repositories in a CodeArtifact domain.

## Example Usage

```hcl
data "aws_codeartifact_authorization_token" "test" {
  domain = aws_codeartifact_domain.test.domain
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The name of the domain that is in scope for the generated authorization token.
* `domain_owner` - (Optional) The account number of the AWS account that owns the domain.
* `duration_seconds` - (Optional) The time, in seconds, that the generated authorization token is valid. Valid values are `0` and between `900` and `43200`.

## Attributes Reference

In addition to the argument above, the following attributes are exported:

* `authorization_token` - Temporary authorization token.
* `expiration` - The time in UTC RFC3339 format when the authorization token expires.
