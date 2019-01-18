---
layout: "aws"
page_title: "AWS: aws_dms_endpoint"
sidebar_current: "docs-aws-resource-dms-endpoint"
description: |-
  Provides a DMS (Data Migration Service) endpoint resource.
---

# aws_dms_endpoint

Provides a DMS (Data Migration Service) endpoint resource. DMS endpoints can be created, updated, deleted, and imported.

~> **Note:** All arguments including the password will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).

## Example Usage

```hcl
# Create a new endpoint
resource "aws_dms_endpoint" "test" {
  certificate_arn             = "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
  database_name               = "test"
  endpoint_id                 = "test-dms-endpoint-tf"
  endpoint_type               = "source"
  engine_name                 = "aurora"
  extra_connection_attributes = ""
  kms_key_arn                 = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
  password                    = "test"
  port                        = 3306
  server_name                 = "test"
  ssl_mode                    = "none"

  tags = {
    Name = "test"
  }

  username = "test"
}
```

## Argument Reference

The following arguments are supported:

* `certificate_arn` - (Optional, Default: empty string) The Amazon Resource Name (ARN) for the certificate.
* `database_name` - (Optional) The name of the endpoint database.
* `endpoint_id` - (Required) The database endpoint identifier.

    - Must contain from 1 to 255 alphanumeric characters or hyphens.
    - Must begin with a letter
    - Must contain only ASCII letters, digits, and hyphens
    - Must not end with a hyphen
    - Must not contain two consecutive hyphens

* `endpoint_type` - (Required) The type of endpoint. Can be one of `source | target`.
* `engine_name` - (Required) The type of engine for the endpoint. Can be one of `mysql | oracle | postgres | mariadb | aurora | redshift | sybase | sqlserver | dynamodb | mongodb | s3 | azuredb`.
* `extra_connection_attributes` - (Optional) Additional attributes associated with the connection. For available attributes see [Using Extra Connection Attributes with AWS Database Migration Service](http://docs.aws.amazon.com/dms/latest/userguide/CHAP_Introduction.ConnectionAttributes.html).
* `kms_key_arn` - (Required when `engine_name` is `mongodb`, optional otherwise) The Amazon Resource Name (ARN) for the KMS key that will be used to encrypt the connection parameters. If you do not specify a value for `kms_key_arn`, then AWS DMS will use your default encryption key. AWS KMS creates the default encryption key for your AWS account. Your AWS account has a different default encryption key for each AWS region.
* `password` - (Optional) The password to be used to login to the endpoint database.
* `port` - (Optional) The port used by the endpoint database.
* `server_name` - (Optional) The host name of the server.
* `ssl_mode` - (Optional, Default: none) The SSL mode to use for the connection. Can be one of `none | require | verify-ca | verify-full`
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `username` - (Optional) The user name to be used to login to the endpoint database.
* `service_access_role` - (Optional) The Amazon Resource Name (ARN) used by the service access IAM role for dynamodb endpoints.
* `mongodb_settings` - (Optional) Settings for the source MongoDB endpoint. Available settings are `auth_type` (default: `PASSWORD`), `auth_mechanism` (default: `DEFAULT`), `nesting_level` (default: `NONE`), `extract_doc_id` (default: `false`), `docs_to_investigate` (default: `1000`) and `auth_source` (default: `admin`). For more details, see [Using MongoDB as a Source for AWS DMS](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.MongoDB.html).
* `s3_settings` - (Optional) Settings for the target S3 endpoint. Available settings are `service_access_role_arn`, `external_table_definition`, `csv_row_delimiter` (default: `\\n`), `csv_delimiter` (default: `,`), `bucket_folder`, `bucket_name` and `compression_type` (default: `NONE`). For more details, see [Using Amazon S3 as a Target for AWS Database Migration Service](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.S3.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `endpoint_arn` - The Amazon Resource Name (ARN) for the endpoint.

## Import

Endpoints can be imported using the `endpoint_id`, e.g.

```
$ terraform import aws_dms_endpoint.test test-dms-endpoint-tf
```
