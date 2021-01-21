---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_data_catalog_encryption_settings"
description: |-
  Provides a Glue Data Catalog Encryption Settings resource.
---

# Resource: aws_glue_data_catalog_encryption_settings

Provides a Glue Data Catalog Encryption Settings resource.

## Example Usage

```hcl
resource "aws_glue_data_catalog_encryption_settings" "example" {
  data_catalog_encryption_settings {
    connection_password_encryption {
      aws_kms_key_id                       = aws_kms_key.test.arn
      return_connection_password_encrypted = true
    }

    encryption_at_rest {
      catalog_encryption_mode = "SSE-KMS"
      sse_aws_kms_key_id      = aws_kms_key.test.arn
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `data_catalog_encryption_settings` – (Required) The security configuration to set. see [Data Catalog Encryption Settings](#data_catalog_encryption_settings).
* `catalog_id` – (Optional) The ID of the Data Catalog to set the security configuration for. If none is provided, the AWS account ID is used by default.

### data_catalog_encryption_settings

* `connection_password_encryption` - (Required) When connection password protection is enabled, the Data Catalog uses a customer-provided key to encrypt the password as part of CreateConnection or UpdateConnection and store it in the ENCRYPTED_PASSWORD field in the connection properties. You can enable catalog encryption or only password encryption. see [Connection Password Encryption](#connection_password_encryption).
* `encryption_at_rest` - (Required) Specifies the encryption-at-rest configuration for the Data Catalog. see [Encryption At Rest](#encryption_at_rest).

### connection_password_encryption

* `return_connection_password_encrypted` - (Required) When set to `true`, passwords remain encrypted in the responses of GetConnection and GetConnections. This encryption takes effect independently of the catalog encryption.
* `aws_kms_key_id` - (Optional) A KMS key ARN that is used to encrypt the connection password. If connection password protection is enabled, the caller of CreateConnection and UpdateConnection needs at least `kms:Encrypt` permission on the specified AWS KMS key, to encrypt passwords before storing them in the Data Catalog.

### encryption_at_rest

* `catalog_encryption_mode` - (Required) The encryption-at-rest mode for encrypting Data Catalog data. Valid values are `DISABLED` and `SSE-KMS`.
* `sse_aws_kms_key_id` - (Optional) The ARN of the AWS KMS key to use for encryption at rest.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Data Catalog to set the security configuration for.

## Import

Glue Data Catalog Encryption Settings can be imported using `CATALOG-ID` (AWS account ID if not custom), e.g.

```
$ terraform import aws_glue_data_catalog_encryption_settings.example 123456789012
```
