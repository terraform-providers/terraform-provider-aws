---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_distribution_configuration"
description: |-
    Provides details about an Image Builder Distribution Configuration
---

# Data Source: aws_imagebuilder_distribution_configuration

Provides details about an Image Builder Distribution Configuration.

## Example Usage

```hcl
data "aws_imagebuilder_distribution_configuration" "example" {
  arn = "arn:aws:imagebuilder:us-west-2:aws:distribution-configuration/example"
}
```

## Argument Reference

* `arn` - (Required) Amazon Resource Name (ARN) of the distribution configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `date_created` - Date the distribution configuration was created.
* `date_updated` - Date the distribution configuration was updated.
* `description` - Description of the distribution configuration.
* `distribution` - Set of distributions.
    * `ami_distribution_configuration` - Nested list of AMI distribution configuration.
        * `ami_tags` - Key-value map of tags to apply to distributed AMI.
        * `description` - Description to apply to distributed AMI.
        * `kms_key_id` - Amazon Resource Name (ARN) of Key Management Service (KMS) Key to encrypt AMI.
        * `launch_permission` - Nested list of EC2 launch permissions.
            * `user_groups` - Set of EC2 launch permission user groups.
            * `user_ids` - Set of AWS Account identifiers.
        * `target_account_ids` - Set of target AWS Account identifiers.
    * `license_configuration_arns` - Set of Amazon Resource Names (ARNs) of License Manager License Configurations.
    * `region` - AWS Region of distribution.
* `name` - Name of the distribution configuration.
* `tags` - Key-value map of resource tags for the distribution configuration.
