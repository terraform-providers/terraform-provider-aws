---
layout: "aws"
page_title: "AWS: aws_datasync_location_nfs"
sidebar_current: "docs-aws-resource-datasync-location-nfs"
description: |-
  Manages an AWS DataSync NFS Location
---

# aws_datasync_location_nfs

Manages an NFS Location within AWS DataSync.

~> **NOTE:** The DataSync Agents must be available before creating this resource.

## Example Usage

```hcl
resource "aws_datasync_location_nfs" "example" {
  server_hostname = "nfs.example.com"
  subdirectory    = "/exported/path"

  on_prem_config {
    agent_arns = ["${aws_datasync_agent.example.arn}"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `on_prem_config` - (Required) Configuration block containing information for connecting to the NFS File System.
* `server_hostname` - (Required) Specifies the IP address or DNS name of the NFS server. The DataSync Agent(s) use this to mount the NFS server.
* `subdirectory` - (Required) Subdirectory to perform actions as source or destination. Should be exported by the NFS server.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location.

### on_prem_config Argument Reference

The following arguments are supported inside the `on_prem_config` configuration block:

* `agent_arns` - (Required) List of Amazon Resource Names (ARNs) of the DataSync Agents used to connect to the NFS server.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the DataSync Location.
* `arn` - Amazon Resource Name (ARN) of the DataSync Location.

## Import

`aws_datasync_location_nfs` can be imported by using the DataSync Task Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_datasync_location_nfs.example arn:aws:datasync:us-east-1:123456789012:location/loc-12345678901234567
```
