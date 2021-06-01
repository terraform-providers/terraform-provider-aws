---
subcategory: "Signer"
layout: "aws"
page_title: "AWS: aws_signer_signing_profile"
description: |-
  Provides a Signer Signing Profile data source.
---

# Data Source: aws_signer_signing_profile

Provides information about a Signer Signing Profile.

## Example Usage

```hcl
data "aws_signer_signing_profile" "production_signing_profile" {
  name = "prod_profile_DdW3Mk1foYL88fajut4mTVFGpuwfd4ACO6ANL0D1uIj7lrn8adK"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the target signing profile.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) for the signing profile.
* `platform_display_name` - A human-readable name for the signing platform associated with the signing profile.
* `platform_id` - The ID of the platform that is used by the target signing profile.
* `revocation_record` - Revocation information for a signing profile.
* `signature_validity_period` - The validity period for a signing job.
* `status` - The status of the target signing profile.
* `tags` - A list of tags associated with the signing profile.
* `version` - The current version of the signing profile.
* `version_arn` - The signing profile ARN, including the profile version.
