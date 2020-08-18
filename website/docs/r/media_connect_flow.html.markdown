---
subcategory: "MediaConnect"
layout: "aws"
page_title: "AWS: aws_media_connect_flow"
description: |-
  Provides an AWS Elemental MediaConnect Flow.
---

# Resource: aws_media_connect_flow

Provides an AWS Elemental MediaConnect Flow.

## Example Usage

```hcl
resource "aws_media_connect_flow" "test" {
  name = "tfflow"

  source {
    name           = "tfsource"
    protocol       = "rtp"
    ingest_port    = 3010
    whitelist_cidr = "10.24.34.0/23"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique identifier describing the channel
* `source` - (Required) Configuration block with source settings. Detailed below.
* `availability_zone` - (Optional) The Availability Zone that you want to create the flow in. These options are limited to the Availability Zones within the current AWS Region.

The `source` object supports the following:

* `name` - (Required) The type of encryption that is used on the content ingested from the source. Defined below.
* `decryption` - (Optional) Configuration block with type of encryption that is used on the content ingested from the source. Defined below.
* `description` - (Optional) A description of the source. This description is not visible outside of the current AWS account.
* `entitlement_arn` - (Optional) The ARN of the entitlement that allows you to subscribe to the flow. The content originator grants the entitlement, and the ARN is auto-generated as part of the originator's flow.
* `ingest_port` - (Optional) The port that the flow listens on for incoming content. If the protocol of the source is Zixi, the port must be set to `2088`.
* `max_bitrate` - (Optional) The smoothing max bitrate for RTP and RTP-FEC streams.
* `max_latency` - (Optional) The maximum latency in milliseconds for Zixi-based streams.
* `protocol` - (Optional) The protocol that the source uses to deliver the content to AWS Elemental MediaConnect.
* `stream_id` - (Optional) The stream ID that you want to use for the transport. This parameter applies only to Zixi-based streams.
* `whitelist_cidr` - (Optional) The range of IP addresses that are allowed to contribute content to your source. Use the form of a Classless Inter-Domain Routing (CIDR) block; for example, 10.0.0.0/16.

The `decryption` object supports the following:

* `algorithm` - (Required) The type of algorithm that is used for the encryption (such as aes128, aes192, or aes256).
* `role_arn` - (Required) The Amazon Resource Name (ARN) of the role that you created during setup (when you set up AWS Elemental MediaConnect as a trusted entity).
* `secret_arn` - (Required) The ARN of the secret that you created in AWS Secrets Manager to store the encryption key.
* `key_type` - (Optional) The type of key that is used for the encryption. If you don't specify a keyType value, the service uses the default setting (static-key).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The same as `arn`
* `arn` - The ARN of the flow
* `availability_zone` - The Availability Zone that the flow was created in.
* `description` - A description of the flow. This description appears only on the AWS Elemental MediaConnect console and is not visible outside of the current AWS account.
* `egress_ip` - The outgoing IP address that AWS Elemental MediaConnect uses to send video from the flow.
* `source` - 
  * `arn` - The ARN of the source.
  * `ingest_ip` - The IP address that the flow listens on for incoming content.
  * `max_bitrate` - The smoothing max bitrate for RTP and RTP-FEC streams.
  * `smoothing_latency` - The smoothing latency in milliseconds for RTP and RTP-FEC streams.

## Import

Media Connect Flow can be imported via the flow ARN, e.g.

```
$ terraform import aws_media_connect_flow.test arn:aws:mediaconnect:us-west-2:123456789012:flow:1-7e7a28d2-163f-4b8f:tfflow
```
