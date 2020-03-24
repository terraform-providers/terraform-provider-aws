---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: cloudfront_field_level_encryption_profile"
description: |-
  Provides a CloudFront Field Level Encryption Profile.
---

# Resource: aws_cloudfront_field_level_encryption_profile

## Example Usage

The following example below creates a CloudFront Field Level Encryption Profile.

```hcl
resource "aws_cloudfront_public_key" "example" {
  comment     = "test public key"
  encoded_key = "${file("public_key.pem")}"
  name        = "test_key"
}

resource "aws_cloudfront_field_level_encryption_profile" "test" {
  comment = "test comment"
  name    = "test profile"

  encryption_entities {
    public_key_id  = "${aws_cloudfront_public_key.example.id}"
    provider_id    = "test profile"
    field_patterns = ["DateOfBirth"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Field Level Encryption Profile.
* `comment` - (Optional) An optional comment about the Field Level Encryption Profile.
* `encryption_entities` - (Required) The [encryption entities](#encryption-entities) config block for field-level encryption profiles that includes the encryption key and field pattern specifications.

### Encryption Entities

* `public_key_id` - (Required) The public key associated with a set of field-level encryption patterns, to be used when encrypting the fields that match the patterns.
* `provider_id` - (Required) The provider associated with the public key being used for encryption.
* `field_patterns` - (Required) The list of Field patterns in a field-level encryption content type profile specify the fields that you want to be encrypted.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `caller_reference` - Internal value used by CloudFront to allow future updates to the Field Level Encryption Profile.
* `etag` - The current version of the Field Level Encryption Profile. For example: `E2QWRUHAPOMQZL`.
* `id` - The identifier for the Field Level Encryption Profile. For example: `K3D5EWEUDCCXON`.
