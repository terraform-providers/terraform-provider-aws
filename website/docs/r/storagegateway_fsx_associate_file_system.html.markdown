---
subcategory: "Storage Gateway"
layout: "aws"
page_title: "AWS: storagegateway_fsx_associate_file_system"
description: |-
  Associate an Amazon FSx file system with the FSx File Gateway. After the association process is complete, the file shares on the Amazon FSx file system are available for access through the gateway. This operation only supports the FSx File Gateway type.
---

# Resource: aws_storagegateway_fsx_associate_file_system

Associate an Amazon FSx file system with the FSx File Gateway. After the association process is complete, the file shares on the Amazon FSx file system are available for access through the gateway. This operation only supports the FSx File Gateway type.

[FSx File Gateway requirements](https://docs.aws.amazon.com/filegateway/latest/filefsxw/Requirements.html).

## Example Usage

```terraform
resource "aws_storagegateway_fsx_associate_file_system" "example" {
  gateway_arn = aws_storagegateway_gateway.example.arn
  location_arn = aws_fsx_windows_file_system.example.arn
  username = "Admin"
  password = "avoid-plaintext-passwords"
  audit_destination_arn = aws_s3_bucket.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `gateway_arn` - (Required) The Amazon Resource Name (ARN) of the gateway.
* `location_arn` - (Required) The Amazon Resource Name (ARN) of the Amazon FSx file system to associate with the FSx File Gateway.
* `username` - (Required) The user name of the user credential that has permission to access the root share of the Amazon FSx file system. The user account must belong to the Amazon FSx delegated admin user group.
* `password` - (Required, sensative) The password of the user credential.
* `audit_destination_arn` - (Optional) The Amazon Resource Name (ARN) of the storage used for the audit logs.
* `cache_attributes` - (Optional) Refresh cache information. see [Cache Attributes](#cache_attributes) for more details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### cache_attributes

* `cache_stale_timeout_in_seconds` - (Optional) Refreshes a file share's cache by using Time To Live (TTL).
 TTL is the length of time since the last refresh after which access to the directory would cause the file gateway
  to first refresh that directory's contents from the Amazon S3 bucket. Valid Values: `0` or `300` to `2592000` seconds (5 minutes to 30 days). Defaults to `0`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the NFS File Share.
* `arn` - Amazon Resource Name (ARN) of the newly created file system association.
* `gateway_arn` - Amazon Resource Name (ARN) of the Storage Gateway
* `location_arn` - Amazon Resource Name (ARN) of the FSx File System
* `username` - Username used to connect to Directory Service
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_storagegateway_fsx_associate_file_system` can be imported by using the NFS File Share Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_storagegateway_fsx_associate_file_system.example arn:aws:storagegateway:us-east-1:123456789012:fs-association/fsa-0DA347732FDB40125
```
