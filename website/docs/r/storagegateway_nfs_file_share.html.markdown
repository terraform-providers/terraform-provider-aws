---
layout: "aws"
page_title: "AWS: aws_storagegateway_nfs_file_share"
sidebar_current: "docs-aws-resource-storagegateway-nfs-file-share"
description: |-
  Manages an AWS Storage Gateway NFS File Share
---

# aws_storagegateway_nfs_file_share

Manages an AWS Storage Gateway NFS File Share.

## Example Usage

```hcl
resource "aws_storagegateway_nfs_file_share" "example" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = "${aws_storagegateway_gateway.example.arn}"
  location_arn = "${aws_s3_bucket.example.arn}"
  role_arn     = "${aws_iam_role.example.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `client_list` - (Required) The list of clients that are allowed to access the file gateway. The list must contain either valid IP addresses or valid CIDR blocks. Set to `["0.0.0.0/0"]` to not limit access. Minimum 1 item. Maximum 100 items.
* `gateway_arn` - (Required) Amazon Resource Name (ARN) of the file gateway.
* `location_arn` - (Required) The ARN of the backed storage used for storing file data.
* `role_arn` - (Required) The ARN of the AWS Identity and Access Management (IAM) role that a file gateway assumes when it accesses the underlying storage.
* `default_storage_class` - (Optional) The default storage class for objects put into an Amazon S3 bucket by the file gateway. Defaults to `S3_STANDARD`. Valid values: `S3_STANDARD`, `S3_STANDARD_IA`, `S3_ONEZONE_IA`.
* `guess_mime_type_enabled` - (Optional) Boolean value that enables guessing of the MIME type for uploaded objects based on file extensions. Defaults to `true`.
* `kms_encrypted` - (Optional) Boolean value if `true` to use Amazon S3 server side encryption with your own AWS KMS key, or `false` to use a key managed by Amazon S3. Defaults to `false`.
* `kms_key_arn` - (Optional) Amazon Resource Name (ARN) for KMS key used for Amazon S3 server side encryption. This value can only be set when `kms_encrypted` is true.
* `nfs_file_share_defaults` - (Optional) Nested argument with file share default values. More information below.
* `object_acl` - (Optional) Access Control List permission for S3 bucket objects. Defaults to `private`.
* `read_only` - (Optional) Boolean to indicate write status of file share. File share does not accept writes if `true`. Defaults to `false`.
* `requester_pays` - (Optional) Boolean who pays the cost of the request and the data download from the Amazon S3 bucket. Set this value to `true` if you want the requester to pay instead of the bucket owner. Defaults to `false`.
* `squash` - (Optional) Maps a user to anonymous user. Defaults to `RootSquash`. Valid values: `RootSquash` (only root is mapped to anonymous user), `NoSquash` (no one is mapped to anonymous user), `AllSquash` (everyone is mapped to anonymous user)

### nfs_file_share_defaults

Files and folders stored as Amazon S3 objects in S3 buckets don't, by default, have Unix file permissions assigned to them. Upon discovery in an S3 bucket by Storage Gateway, the S3 objects that represent files and folders are assigned these default Unix permissions.

* `directory_mode` - (Optional) The Unix directory mode in the string form "nnnn". Defaults to `"0777"`.
* `file_mode` - (Optional) The Unix file mode in the string form "nnnn". Defaults to `"0666"`.
* `group_id` - (Optional) The default group ID for the file share (unless the files have another group ID specified). Defaults to `65534` (`nfsnobody`). Valid values: `0` through `4294967294`.
* `owner_id` - (Optional) The default owner ID for the file share (unless the files have another owner ID specified). Defaults to `65534` (`nfsnobody`). Valid values: `0` through `4294967294`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the NFS File Share.
* `arn` - Amazon Resource Name (ARN) of the NFS File Share.
* `fileshare_id` - ID of the NFS File Share.
* `path` - File share path used by the NFS client to identify the mount point.

## Timeouts

`aws_storagegateway_nfs_file_share` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

* `create` - (Default `10m`) How long to wait for file share creation.
* `update` - (Default `10m`) How long to wait for file share updates.
* `delete` - (Default `10m`) How long to wait for file share deletion.

## Import

`aws_storagegateway_nfs_file_share` can be imported by using the NFS File Share Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_storagegateway_nfs_file_share.example arn:aws:storagegateway:us-east-1:123456789012:share/share-12345678
```
