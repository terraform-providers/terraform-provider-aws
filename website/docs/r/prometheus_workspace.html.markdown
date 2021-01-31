---
subcategory: "Amazon Managed Service for Prometheus (AMP)"
layout: "aws"
page_title: "AWS: aws_prometheus_workspace"
description: |-
  Manages an Amazon Managed Service for Prometheus (AMP) Workspace
---

# Resource: aws_prometheus_workspace

Manages an Amazon Managed Service for Prometheus (AMP) Workspace.

~> **NOTE:** This AWS functionality is in Preview and may change before General Availability release. Backwards compatibility is not guaranteed between Terraform AWS Provider releases.

## Example Usage

```hcl
resource "aws_prometheus_workspace" "demo" {
  alias = "prometheus-test"
}
```

## Argument Reference

The following argument is supported:

* `alias` - (Optional) The alias of the prometheus workspace. See more [in AWS Docs](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-onboard-create-workspace.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the workspace.
* `id` - Identifier of the workspace
* `prometheus_endpoint` - Prometheus endpoint available for this workspace.

## Import

AMP Workspaces can be imported using the identifier, e.g.

```
$ terraform import aws_prometheus_workspace.demo ws-C6DCB907-F2D7-4D96-957B-66691F865D8B
```
