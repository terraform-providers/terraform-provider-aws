---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_infrastructure_configuration"
description: |-
  Manages an Image Builder Infrastructure Configuration
---

# Resource: aws_imagebuilder_infrastructure_configuration

Manages an Image Builder Infrastructure Configuration.

## Example Usage

```hcl
resource "aws_imagebuilder_infrastructure_configuration" "example" {
  description                   = "example description"
  instance_profile_name         = aws_iam_instance_profile.example.name
  instance_types                = ["t2.nano", "t3.micro"]
  key_pair                      = aws_key_pair.example.key_name
  name                          = "example"
  security_group_ids            = [aws_security_group.example.id]
  sns_topic_arn                 = aws_sns_topic.example.arn
  subnet_id                     = aws_subnet.main.id
  terminate_instance_on_failure = true

  logging {
    s3_logs {
      s3_bucket_name = aws_s3_bucket.example.bucket
      s3_key_prefix  = "logs"
    }
  }

  tags = {
    foo = "bar"
  }
}
```

## Argument Reference

The following arguments are required:

* `instance_profile_name` - (Required) Name of IAM Instance Profile.
* `name` - (Required) Name for the configuration.

The following arguments are optional:

* `description` - (Optional) Description for the configuration.
* `instance_types` - (Optional) Set of EC2 Instance Types.
* `key_pair` - (Optional) Name of EC2 Key Pair.
* `logging` - (Optional) Configuration block with logging settings. Detailed below.
* `resource_tags` - (Optional) Key-value map of resource tags to assign to infrastructure created by the configuration.
* `security_group_ids` - (Optional) Set of EC2 Security Group identifiers.
* `sns_topic_arn` - (Optional) Amazon Resource Name (ARN) of SNS Topic.
* `subnet_id` - (Optional) EC2 Subnet identifier. Also requires `security_group_ids` argument.
* `tags` - (Optional) Key-value map of resource tags to assign to the configuration.
* `terminate_instance_on_failure` - (Optional) Enable if the instance should be terminated when the pipeline fails. Defaults to `false`.

### logging

The following arguments are required:

* `s3_logs` - (Required) Configuration block with S3 logging settings. Detailed below.

### s3_logs

The following arguments are required:

* `s3_bucket_name` - (Required) Name of the S3 Bucket.

The following arguments are optional:

* `s3_key_prefix` - (Optional) Prefix to use for S3 logs. Defaults to `/`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the configuration.
* `arn` - Amazon Resource Name (ARN) of the configuration.
* `date_created` - Date when the configuration was created.
* `date_updated` - Date when the configuration was updated.

## Import

`aws_imagebuilder_infrastructure_configuration` can be imported using the Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_imagebuilder_infrastructure_configuration.example arn:aws:imagebuilder:us-east-1:123456789012:infrastructure-component/example
```
