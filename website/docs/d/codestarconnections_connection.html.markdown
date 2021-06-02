---
subcategory: "CodeStar Connections"
layout: "aws"
page_title: "AWS: aws_codestarconnections_connection"
description: |-
  Provides details about CodeStar Connection
---

# Data Source: aws_codestarconnections_connection

Provides details about CodeStar Connection.

## Example Usage

```terraform
data "aws_codestarconnections_connection" "example" {
  arn = aws_codestarconnections_connection.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `arn` - (Required) The CodeStar Connection ARN.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `connection_status` - The CodeStar Connection status. Possible values are `PENDING`, `AVAILABLE` and `ERROR`.
* `id` - The CodeStar Connection ARN.
* `host_arn` - The Amazon Resource Name (ARN) of the host associated with the connection.
* `name` - The name of the CodeStar Connection. The name is unique in the calling AWS account.
* `provider_type` - The name of the external provider where your third-party code repository is configured. Possible values are `Bitbucket`, `GitHub`, or `GitHubEnterpriseServer`.
* `tags` - Map of key-value resource tags to associate with the resource.
