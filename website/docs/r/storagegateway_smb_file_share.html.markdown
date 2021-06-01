---
subcategory: "Storage Gateway"
layout: "aws"
page_title: "AWS: aws_storagegateway_smb_file_share"
description: |-
  Manages an AWS Storage Gateway SMB File Share
---

# Resource: aws_storagegateway_smb_file_share

Manages an AWS Storage Gateway SMB File Share.

## Example Usage

### Active Directory Authentication

~> **NOTE:** The gateway must have already joined the Active Directory domain prior to SMB file share creation. e.g. via "SMB Settings" in the AWS Storage Gateway console or `smb_active_directory_settings` in the [`aws_storagegateway_gateway` resource](/docs/providers/aws/r/storagegateway_gateway.html).

```hcl
resource "aws_storagegateway_smb_file_share" "example" {
  authentication = "ActiveDirectory"
  gateway_arn    = aws_storagegateway_gateway.example.arn
  location_arn   = aws_s3_bucket.example.arn
  role_arn       = aws_iam_role.example.arn
}
```

### Guest Authentication

~> **NOTE:** The gateway must have already had the SMB guest password set prior to SMB file share creation. e.g. via "SMB Settings" in the AWS Storage Gateway console or `smb_guest_password` in the [`aws_storagegateway_gateway` resource](/docs/providers/aws/r/storagegateway_gateway.html).

```hcl
resource "aws_storagegateway_smb_file_share" "example" {
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.example.arn
  location_arn   = aws_s3_bucket.example.arn
  role_arn       = aws_iam_role.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `gateway_arn` - (Required) Amazon Resource Name (ARN) of the file gateway.
* `location_arn` - (Required) The ARN of the backed storage used for storing file data.
* `role_arn` - (Required) The ARN of the AWS Identity and Access Management (IAM) role that a file gateway assumes when it accesses the underlying storage.
* `admin_user_list` - (Optional) A list of users in the Active Directory that have admin access to the file share. Only valid if `authentication` is set to `ActiveDirectory`.
* `authentication` - (Optional) The authentication method that users use to access the file share. Defaults to `ActiveDirectory`. Valid values: `ActiveDirectory`, `GuestAccess`.
* `audit_destination_arn` - (Optional) The Amazon Resource Name (ARN) of the CloudWatch Log Group used for the audit logs.
* `default_storage_class` - (Optional) The default storage class for objects put into an Amazon S3 bucket by the file gateway. Defaults to `S3_STANDARD`. Valid values: `S3_STANDARD`, `S3_STANDARD_IA`, `S3_ONEZONE_IA`.
* `file_share_name` - (Optional) The name of the file share. Must be set if an S3 prefix name is set in `location_arn`.
* `guess_mime_type_enabled` - (Optional) Boolean value that enables guessing of the MIME type for uploaded objects based on file extensions. Defaults to `true`.
* `invalid_user_list` - (Optional) A list of users in the Active Directory that are not allowed to access the file share. Only valid if `authentication` is set to `ActiveDirectory`.
* `kms_encrypted` - (Optional) Boolean value if `true` to use Amazon S3 server side encryption with your own AWS KMS key, or `false` to use a key managed by Amazon S3. Defaults to `false`.
* `kms_key_arn` - (Optional) Amazon Resource Name (ARN) for KMS key used for Amazon S3 server side encryption. This value can only be set when `kms_encrypted` is true.
* `object_acl` - (Optional) Access Control List permission for S3 bucket objects. Defaults to `private`.
* `cache_attributes` - (Optional) Refresh cache information. see [Cache Attributes](#cache_attributes) for more details.
* `read_only` - (Optional) Boolean to indicate write status of file share. File share does not accept writes if `true`. Defaults to `false`.
* `requester_pays` - (Optional) Boolean who pays the cost of the request and the data download from the Amazon S3 bucket. Set this value to `true` if you want the requester to pay instead of the bucket owner. Defaults to `false`.
* `smb_acl_enabled` - (Optional) Set this value to `true` to enable ACL (access control list) on the SMB fileshare. Set it to `false` to map file and directory permissions to the POSIX permissions. This setting applies only to `ActiveDirectory` authentication type.
* `case_sensitivity` - (Optional) The case of an object name in an Amazon S3 bucket. For `ClientSpecified`, the client determines the case sensitivity. For `CaseSensitive`, the gateway determines the case sensitivity. The default value is `ClientSpecified`.
* `valid_user_list` - (Optional) A list of users in the Active Directory that are allowed to access the file share. Only valid if `authentication` is set to `ActiveDirectory`.
* `access_based_enumeration` - (Optional) The files and folders on this share will only be visible to users with read access. Default value is `false`.
* `notification_policy` - (Optional) The notification policy of the file share. For more information see the [AWS Documentation](https://docs.aws.amazon.com/storagegateway/latest/APIReference/API_CreateNFSFileShare.html#StorageGateway-CreateNFSFileShare-request-NotificationPolicy). Default value is `{}`.
* `tags` - (Optional) Key-value map of resource tags

### cache_attributes

* `cache_stale_timeout_in_seconds` - (Optional) Refreshes a file share's cache by using Time To Live (TTL).
 TTL is the length of time since the last refresh after which access to the directory would cause the file gateway
  to first refresh that directory's contents from the Amazon S3 bucket. Valid Values: 300 to 2,592,000 seconds (5 minutes to 30 days)


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the SMB File Share.
* `arn` - Amazon Resource Name (ARN) of the SMB File Share.
* `fileshare_id` - ID of the SMB File Share.
* `path` - File share path used by the NFS client to identify the mount point.

## Timeouts

`aws_storagegateway_smb_file_share` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `10m`) How long to wait for file share creation.
* `update` - (Default `10m`) How long to wait for file share updates.
* `delete` - (Default `15m`) How long to wait for file share deletion.

## Import

`aws_storagegateway_smb_file_share` can be imported by using the SMB File Share Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_storagegateway_smb_file_share.example arn:aws:storagegateway:us-east-1:123456789012:share/share-12345678
```
