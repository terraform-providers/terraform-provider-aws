---
layout: "aws"
page_title: "AWS: aws_datasync_agent"
sidebar_current: "docs-aws-resource-datasync-agent"
description: |-
  Manages an AWS DataSync Agent in the provider region
---

# aws_datasync_agent

Manages an AWS DataSync Agent deployed on premises.

~> **NOTE:** One of `activation_key` or `ip_address` must be provided for resource creation (agent activation). Neither is required for resource import. If using `ip_address`, Terraform must be able to make an HTTP (port 80) GET request to the specified IP address from where it is running. The agent will turn off that HTTP server after activation.

## Example Usage

```hcl
resource "aws_datasync_agent" "example" {
  ip_address = "1.2.3.4"
  name       = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the DataSync Agent.
* `activation_key` - (Optional) DataSync Agent activation key during resource creation. Conflicts with `ip_address`. If an `ip_address` is provided instead, Terraform will retrieve the `activation_key` as part of the resource creation.
* `ip_address` - (Optional) DataSync Agent IP address to retrieve activation key during resource creation. Conflicts with `activation_key`. DataSync Agent must be accessible on port 80 from where Terraform is running.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Agent.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the DataSync Agent.
* `arn` - Amazon Resource Name (ARN) of the DataSync Agent.

## Timeouts

`aws_datasync_agent` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

* `create` - (Default `10m`) How long to wait for agent activation and connection to DataSync.

## Import

`aws_datasync_agent` can be imported by using the DataSync Agent Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_datasync_agent.example arn:aws:datasync:us-east-1:123456789012:agent/agent-12345678901234567
```
