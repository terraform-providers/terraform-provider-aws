---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_code_signing_config"
description: |-
  Provides a Lambda Code Signing Config data source.
---

# Data Source: aws_lambda_code_signing_config

Provides information about a Lambda Code Signing Config. A code signing configuration defines a list of allowed signing profiles and defines the code-signing validation policy (action to be taken if deployment validation checks fail).

For information about Lambda code signing configurations and how to use them, see [configuring code signing for Lambda functions][1]

## Example Usage

```hcl
data "aws_lambda_code_signing_config" "existing_csc" {
  arn = "arn:aws:lambda:${var.aws_region}:${var.aws_account}:code-signing-config:csc-0f6c334abcdea4d8b"
}
```

## Argument Reference

The following arguments are supported:

* `arn` - (Required) The Amazon Resource Name (ARN) of the code signing configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `allowed_publishers` - List of allowed publishers as signing profiles for this code signing configuration.
* `config_id` - Unique identifier for the code signing configuration.
* `description` - Code signing configuration description.
* `last_modified` - The date and time that the code signing configuration was last modified.
* `policies` - List of code signing policies that control the validation failure action for signature mismatch or expiry.

`allowed_publishers` is exported with the following attribute:

* `signing_profile_version_arns` - The Amazon Resource Name (ARN) for each of the signing profiles. A signing profile defines a trusted user who can sign a code package.

`policies` is exported with the following attribute:

* `untrusted_artifact_on_deployment` - Code signing configuration policy for deployment validation failure.

[1]: https://docs.aws.amazon.com/lambda/latest/dg/configuration-codesigning.html
