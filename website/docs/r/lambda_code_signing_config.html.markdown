---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_code_signing_config"
description: |-
  Provides a Lambda Code Signing Config resource.
---

# Resource: aws_lambda_code_signing_config

Provides a Lambda Code Signing Config resource. A code signing configuration defines a list of allowed signing profiles and defines the code-signing validation policy (action to be taken if deployment validation checks fail).

For information about Lambda code signing configurations and how to use them, see [configuring code signing for Lambda functions][1]

## Example Usage

```hcl
resource "aws_lambda_code_signing_config" "new_csc" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.example1.arn,
      aws_signer_signing_profile.example2.arn,
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Warn"
  }

  description = "My awesome code signing config."
}
```

## Argument Reference

* `allowed_publishers` (Required) A configuration block of allowed publishers as signing profiles for this code signing configuration. Detailed below.
* `policies` (Optional) A configuration block of code signing policies that define the actions to take if the validation checks fail. Detailed below.
* `description` - (Optional) Descriptive name for this code signing configuration.

The `allowed_publishers` block supports the following argument:

* `signing_profile_version_arns` - (Required) The Amazon Resource Name (ARN) for each of the signing profiles. A signing profile defines a trusted user who can sign a code package.

The `policies` block supports the following argument:

* `untrusted_artifact_on_deployment` - (Required) Code signing configuration policy for deployment validation failure. If you set the policy to Enforce, Lambda blocks the deployment request if code-signing validation checks fail. If you set the policy to Warn, Lambda allows the deployment and creates a CloudWatch log. Valid values: `Warn`, `Enforce`. Default value: `Warn`.


## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) of the code signing configuration.
* `config_id` - Unique identifier for the code signing configuration.
* `last_modified` - The date and time that the code signing configuration was last modified.

[1]: https://docs.aws.amazon.com/lambda/latest/dg/configuration-codesigning.html

## Import

Code Signing Configs can be imported using their ARN, e.g.

```
$ terraform import aws_lambda_code_signing_config.imported_csc arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-0f6c334abcdea4d8b
```
