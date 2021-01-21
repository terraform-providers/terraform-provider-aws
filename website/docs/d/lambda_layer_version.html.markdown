---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_layer_version"
description: |-
  Provides a Lambda Layer Version data source.
---

# aws_lambda_layer_version

Provides information about a Lambda Layer Version.

## Example Usage

```hcl
variable "layer_name" {
  type = string
}

data "aws_lambda_layer_version" "existing" {
  layer_name = var.layer_name
}
```

## Argument Reference

The following arguments are supported:

* `layer_name` - (Required) Name of the lambda layer.
* `version` - (Optional) Specific layer version. Conflicts with `compatible_runtime`. If omitted, the latest available layer version will be used.
* `compatible_runtime` (Optional) Specific runtime the layer version must support. Conflicts with `version`. If specified, the latest available layer version supporting the provided runtime will be used.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `description` - Description of the specific Lambda Layer version.
* `license_info` - License info associated with the specific Lambda Layer version.
* `compatible_runtimes` - A list of [Runtimes][1] the specific Lambda Layer version is compatible with.
* `arn` - The Amazon Resource Name (ARN) of the Lambda Layer with version.
* `layer_arn` - The Amazon Resource Name (ARN) of the Lambda Layer without version.
* `created_date` - The date this resource was created.
* `signing_job_arn` - The Amazon Resource Name (ARN) of a signing job.
* `signing_profile_version_arn` - The Amazon Resource Name (ARN) for a signing profile version.
* `source_code_hash` - Base64-encoded representation of raw SHA-256 sum of the zip file.
* `source_code_size` - The size in bytes of the function .zip file.
* `version` - This Lamba Layer version.

[1]: https://docs.aws.amazon.com/lambda/latest/dg/API_GetLayerVersion.html#SSS-GetLayerVersion-response-CompatibleRuntimes

