---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_location_nfs"
description: |-
  Manages an AWS DataSync NFS Location
---

# Resource: aws_datasync_location_nfs

Manages an NFS Location within AWS DataSync.

~> **NOTE:** The DataSync Agents must be available before creating this resource.

## Example Usage

```terraform
resource "aws_datasync_location_nfs" "example" {
  server_hostname = "nfs.example.com"
  subdirectory    = "/exported/path"

  on_prem_config {
    agent_arns = [aws_datasync_agent.example.arn]
  }
}
```

## Argument Reference

The following arguments are supported:

* `mount_options` - (Optional) Configuration block containing mount options used by DataSync to access the NFS Server.
* `on_prem_config` - (Required) Configuration block containing information for connecting to the NFS File System.
* `server_hostname` - (Required) Specifies the IP address or DNS name of the NFS server. The DataSync Agent(s) use this to mount the NFS server.
* `subdirectory` - (Required) Subdirectory to perform actions as source or destination. Should be exported by the NFS server.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### mount_options Argument Reference

The following arguments are supported inside the `mount_options` configuration block:

* `version` - (Optional) The specific NFS version that you want DataSync to use for mounting your NFS share. Valid values: `AUTOMATIC`, `NFS3`, `NFS4_0` and `NFS4_1`. Default: `AUTOMATIC`

### on_prem_config Argument Reference

The following arguments are supported inside the `on_prem_config` configuration block:

* `agent_arns` - (Required) List of Amazon Resource Names (ARNs) of the DataSync Agents used to connect to the NFS server.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the DataSync Location.
* `arn` - Amazon Resource Name (ARN) of the DataSync Location.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_datasync_location_nfs` can be imported by using the DataSync Task Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_datasync_location_nfs.example arn:aws:datasync:us-east-1:123456789012:location/loc-12345678901234567
```
