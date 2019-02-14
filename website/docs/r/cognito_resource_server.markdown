---
layout: "aws"
page_title: "AWS: aws_cognito_resource_server"
side_bar_current: "docs-aws-resource-cognito-resource-server"
description: |-
  Provides a Cognito Resource Server.
---

# aws_cognito_resource_server

Provides a Cognito Resource Server.

## Example Usage

### Create a basic resource server

```hcl
resource "aws_cognito_user_pool" "pool" {
  name = "pool"
}

resource "aws_cognito_resource_server" "resource" {
  identifier = "https://example.com"
  name       = "example"

  user_pool_id = "${aws_cognito_user_pool.pool.id}"
}
```

### Create a resource server with sample-scope

```hcl
resource "aws_cognito_user_pool" "pool" {
  name = "pool"
}

resource "aws_cognito_resource_server" "resource" {
  identifier = "https://example.com"
  name       = "example"

  scope {
    scope_name        = "sample-scope"
    scope_description = "a Sample Scope Description"
  }

  user_pool_id = "${aws_cognito_user_pool.pool.id}"
}
```

## Argument Reference

The following arguments are supported:

* `identifier` - (Required) An identifier for the resource server.
* `name` - (Required) A name for the resource server.
* `scope` - (Optional) A list of [Authorization Scope](#authorization_scope).

### Authorization Scope

* `scope_name` - (Required) The scope name.
* `scope_description` - (Required) The scope description.

## Attribute Reference

In addition to the arguments, which are exported, the following attributes are exported:

* `scope_identifiers` - A list of all scopes configured for this resource server in the format identifier/scope_name.

## Import

`aws_cognito_resource_server` can be imported using their User Pool ID and Identifier, e.g.

```
$ terraform import aws_cognito_resource_server.example xxx_yyyyy|https://example.com
```
