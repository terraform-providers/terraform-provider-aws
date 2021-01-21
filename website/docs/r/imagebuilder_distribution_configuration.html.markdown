---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_distribution_configuration"
description: |-
    Manage an Image Builder Distribution Configuration
---

# Resource: aws_imagebuilder_distribution_configuration

Manages an Image Builder Distribution Configuration.

## Example Usage

```hcl
resource "aws_imagebuilder_distribution_configuration" "example" {
  name = "example"

  distribution {
    ami_distribution_configuration {
      ami_tags = {
        CostCenter = "IT"
      }

      name = "example-{{ imagebuilder:buildDate }}"

      launch_permission {
        user_ids = ["123456789012"]
      }
    }

    region = "us-east-1"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the distribution configuration.
* `distribution` - (Required) One or more configuration blocks with distribution settings. Detailed below.

The following arguments are optional:

* `description` - (Optional) Description of the distribution configuration.
* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of the Key Management Service (KMS) Key used to encrypt the distribution configuration.
* `tags` - (Optional) Key-value map of resource tags for the distribution configuration.

### distribution

The following arguments are required:

* `region` - (Required) AWS Region for the distribution.

The following arguments are optional:

* `ami_distribution_configuration` - (Optional) Configuration block with Amazon Machine Image (AMI) distribution settings. Detailed below.
* `license_configuration_arns` - (Optional) Set of Amazon Resource Names (ARNs) of License Manager License Configurations.

### ami_distribution_configuration

The following arguments are optional:

* `ami_tags` - (Optional) Key-value map of tags to apply to the distributed AMI.
* `description` - (Optional) Description to apply to the distributed AMI.
* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of the Key Management Service (KMS) Key to encrypt the distributed AMI.
* `launch_permission` - (Optional) Configuration block of EC2 launch permissions to apply to the distributed AMI. Detailed below.
* `name` - (Optional) Name to apply to the distributed AMI.
* `target_account_ids` - (Optional) Set of AWS Account identifiers to distribute the AMI.

### launch_permission

The following arguments are optional:

* `user_groups` - (Optional) Set of EC2 launch permission user groups to assign. Use `all` to distribute a public AMI.
* `user_ids` - (Optional) Set of AWS Account identifiers to assign.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - (Required) Amazon Resource Name (ARN) of the distribution configuration.
* `date_created` - Date the distribution configuration was created.
* `date_updated` - Date the distribution configuration was updated.

## Import

`aws_imagebuilder_distribution_configurations` resources can be imported by using the Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_imagebuilder_distribution_configuration.example arn:aws:imagebuilder:us-east-1:123456789012:distribution-configuration/example
```
