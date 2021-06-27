---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: cloudfront_field_level_encryption_config"
description: |-
  Provides a CloudFront Field Level Encryption Config.
---

# Resource: aws_cloudfront_field_level_encryption_config

## Example Usage

The following example below creates a CloudFront Field Level Encryption Config.

```hcl
resource "aws_cloudfront_field_level_encryption_config" "test" {
  comment = "test comment"

  content_type_profile_config {
    forward_when_content_type_is_unknown = true

    content_type_profile {
      content_type = "application/x-www-form-urlencoded"
      format       = "URLEncoded"
    }
  }

  query_arg_profile_config {
    forward_when_query_arg_is_unknown = true

    query_arg_profile {
      profile_id = aws_cloudfront_field_level_encryption_profile.test.id
      query_arg  = "URLEncoded"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Field Level Encryption Config.
* `comment` - (Optional) An optional comment about the Field Level Encryption Config.
* `content_type_profile_config` - (Required) [Content Type Profile Config](#content-type-profile-config) specifies when to forward content if a content type isn't recognized and profiles to use as by default in a request if a query argument doesn't specify a profile to use.
* `query_arg_profile_config` - (Required) [Content Type Profile](#query-arg-profile-config) that specifies when to forward content if a profile isn't found and the profile that can be provided as a query argument in a request.


### Content Type Profile Config

* `forward_when_content_type_is_unknown` - (Required) specifies what to do when an unknown content type is provided for the profile. If true, content is forwarded without being encrypted when the content type is unknown. If false (the default), an error is returned when the content type is unknown.
* `content_type_profile` - (Required) The configuration for a field-level encryption content type-profile. see [Content Type Profile](#content-type-profile).

### Content Type Profile

* `content_type` - (Required) he content type for a field-level encryption content type-profile mapping. valid value is `application/x-www-form-urlencoded`.
* `format` - (Required) The format for a field-level encryption content type-profile mapping. valid value is `URLEncoded`.
* `profile_id` - (Optional) The profile ID for a field-level encryption content type-profile mapping.

### Query Arg Profile Config

* `forward_when_query_arg_is_unknown` - (Required) Flag to set if you want a request to be forwarded to the origin even if the profile specified by the field-level encryption query argument, fle-profile, is unknown.
* `query_arg_profile` - (Required) Profiles specified for query argument-profile mapping for field-level encryption. see [Query Arg Profile](#query-arg-profile).

### Query Arg Profile

* `profile_id` - (Required) ID of profile to use for field-level encryption query argument-profile mapping
* `query_arg` - (Required) Query argument for field-level encryption query argument-profile mapping.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `caller_reference` - Internal value used by CloudFront to allow future updates to the Field Level Encryption Config.
* `etag` - The current version of the Field Level Encryption Config. For example: `E2QWRUHAPOMQZL`.
* `id` - The identifier for the Field Level Encryption Config. For example: `K3D5EWEUDCCXON`.

## Import

Cloudfront Field Level Encryption Config can be imported using the `id`, e.g.

```
$ terraform import aws_cloudfront_field_level_encryption_config.config E74FTE3AEXAMPLE
```