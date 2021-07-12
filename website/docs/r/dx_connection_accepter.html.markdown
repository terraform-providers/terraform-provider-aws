---
layout: "aws"
page_title: "AWS: aws_dx_connection_accepter"
sidebar_current: "docs-aws-resource-dx-connection-accepter"
description: |-
  Provides a resource to manage the accepter's side of a Direct Connect connection.
---

# Resource: aws_dx_connection_accepter

Provides a resource to manage the accepter's side of a Direct Connect connection.
This resource accepts ownership of a connection created by another AWS account.

## Example Usage

```hcl
provider "aws" {
  # Creator's credentials.
}

provider "aws" {
  alias = "accepter"
  region = "us-east-1"

  # Accepter's credentials.
}

resource "aws_dx_connection" "main" {
  name      = "tf-dx-connection"
  bandwidth = "1Gbps"
  location  = "EqDC2"
}

resource "aws_dx_connection_accepter" "accepter_primary" {
  provider         = "aws.accepter"
  dx_connection_id = "${aws_dx_connection.main.id}"

  tags = {
    Side = "Accepter"
  }
}
```

## Argument Reference

The following arguments are supported:

* `dx_connection_id` - (Required) The ID of the Direct Connect connection to accept.
* `tags` - (Optional) A mapping of tags to assign to the resource.

### Removing `aws_dx_connection_accepter` from your configuration

AWS allows a Direct Connect connection to be deleted from either the allocator's or accepter's side.
However, Terraform only allows the Direct Connect connection to be deleted from the allocator's side
by removing the corresponding `aws_dx_connection` resource from your configuration.
Removing a `aws_dx_connection_accepter` resource from your configuration will remove it
from your statefile and management, **but will not delete the Direct Connect connection.**

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the connection.
* `arn` - The ARN of the connection.

## Timeouts

`aws_dx_connection_accepter` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `10 minutes`) Used for creating connection
- `delete` - (Default `10 minutes`) Used for destroying connection

## Import

Direct Connect connections can be imported using the `connection id`, e.g.

```
$ terraform import aws_dx_connection_accepter.test dxcon-33cc44dd
```
