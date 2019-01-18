---
layout: "aws"
page_title: "AWS: aws_media_package_channel"
sidebar_current: "docs-aws-resource-media-package-channel"
description: |-
  Provides an AWS Elemental MediaPackage Channel.
---

# aws_media_package_channel

Provides an AWS Elemental MediaPackage Channel.

## Example Usage

```hcl
resource "aws_media_package_channel" "kittens" {
  channel_id  = "kitten-channel"
  description = "A channel dedicated to amusing videos of kittens."
}
```

## Argument Reference

The following arguments are supported:

* `channel_id` - (Required) A unique identifier describing the channel
* `description` - (Optional) A description of the channel

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The same as `channel_id`
* `arn` - The ARN of the channel
* `hls_ingest` - A single item list of HLS ingest information
  * `ingest_endpoints` - A list of the ingest endpoints
    * `password` - The password
    * `url` - The URL
    * `username` - The username

## Import

Media Package Channels can be imported via the channel ID, e.g.

```
$ terraform import aws_media_package_channel.kittens kittens-channel
```
