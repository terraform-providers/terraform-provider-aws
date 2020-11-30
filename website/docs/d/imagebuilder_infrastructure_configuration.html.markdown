---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_infrastructure_configuration"
description: |-
    Provides details about an Image Builder Infrastructure Configuration
---

# Data Source: aws_imagebuilder_infrastructure_configuration

Provides details about an Image Builder Infrastructure Configuration.

## Example Usage

```hcl
data "aws_imagebuilder_infrastructure_configuration" "example" {
  arn = "arn:aws:imagebuilder:us-west-2:aws:infrastructure-configuration/example"
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) Amazon Resource Name (ARN) of the infrastructure configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `date_created` - Date the infrastructure configuration was created.
* `date_created` - Date the infrastructure configuration was updated.
* `description` - Description of the infrastructure configuration.
* `instance_profile_name` - Name of the IAM Instance Profile associated with the configuration.
* `instance_types` - Set of EC2 Instance Types associated with the configuration.
* `key_pair` - Name of the EC2 Key Pair associated with the configuration.
* `logging` - Nested list of logging settings.
    * `s3_logs` - Nested list of S3 logs settings.
        * `s3_bucket_name` - Name of the S3 Bucket for logging.
        * `s3_key_prefix` - Key prefix for S3 Bucket logging.
* `name` - Name of the infrastructure configuration.
* `resource_tags` - Key-value map of resource tags for the infrastructure created by the infrastructure configuration.
* `security_group_ids` - Set of EC2 Security Group identifiers associated with the configuration.
* `sns_topic_arn` - Amazon Resource Name (ARN) of the SNS Topic associated with the configuration.
* `subnet_id` - Identifier of the EC2 Subnet associated with the configuration.
* `tags` - Key-value map of resource tags for the infrastructure configuration.
* `terminate_instance_on_failure` - Whether instances are terminated on failure.
