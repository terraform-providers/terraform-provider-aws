---
subcategory: "IoTAnalytics"
layout: "aws"
page_title: "AWS: aws_iotanalytics_channel"
sidebar_current: "docs-aws-resource-iotanalytics-channel"
description: |-
    Creates and manages an AWS IoTAnalytics Channel
---

# Resource: aws_iotanalytics_channel

## Example Usage

```hcl
resource "aws_iotanalytics_channel" "channel_example" {
  name        = "ChannelName"

  storage {
      service_managed_s3 {}
  }
}
```

## Argument Reference

* `name` - (Required) The name of the input.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the channel.

The `storage` - (Optional) The definition of the input. Object takes the following arguments:

* `service_managed_s3` - Object (Optional) Used to store channel data in an S3 bucket managed by the AWS IoT Analytics service. Takes
* `customer_managed_s3` - Object (Optional) Used to store channel data in an S3 bucket that you manage.
    * `bucket` - (Required) The name of the Amazon S3 bucket in which channel data is stored.
    * `key_prefix` - (Optional) The prefix used to create the keys of the channel data objects. Each object in an Amazon S3 bucket has a key that is its unique identifier within the bucket (each object in a bucket has exactly one key).
    * `role_arn` - (Required) The ARN of the role which grants AWS IoT Analytics permission to interact with your Amazon S3 resources

`Note:` You can use only one parameter for `storage` (`service_managed_s3` or `customer_managed_s3`), otherwise you will get an error.
If you don't choose any parameter or don't define `storage` object, `service_managed_s3` will be chosen by AWS.

The `retention_period` - (Optional) How long, in days, message data is kept for the channel. Object takes following arguments.

* `number_of_days` - (Optional) The number of days that message data is kept.
* `unlimited` - (Optional) If true, message data is kept indefinitely.

`Note:` You can use only one parameter for `retention_period` (`number_of_days` or `unlimited`), otherwise you will get an error.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the input
* `arn` - The ARN of the channel.

## Import

IoTAnalytics Channel can be imported using the `name`, e.g.

```
$ terraform import aws_iotanalytics_channel.channel <name>
```
