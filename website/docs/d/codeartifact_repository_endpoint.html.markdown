---
subcategory: "CodeArtifact"
layout: "aws"
page_title: "AWS: aws_codeartifact_repository_endpoint"
description: |-
    Provides details about a CodeArtifact Repository Endpoint
---

# Data Source: aws_codeartifact_repository_endpoint

The CodeArtifact Repository Endpoint data source returns the endpoint of a repository for a specific package format.

## Example Usage

```hcl
data "aws_codeartifact_repository_endpoint" "test" {
  domain     = aws_codeartifact_domain.test.domain
  repository = aws_codeartifact_repository.test.repository
  format     = "npm"
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The name of the domain that contains the repository.
* `repository` - (Required) The name of the repository.
* `format` - (Required) Which endpoint of a repository to return. A repository has one endpoint for each package format: `npm`, `pypi`, `maven`, and `nuget`.
* `domain_owner` - (Optional) The account number of the AWS account that owns the domain.

## Attributes Reference

In addition to the argument above, the following attributes are exported:

* `repository_endpoint` - The URL of the returned endpoint.
