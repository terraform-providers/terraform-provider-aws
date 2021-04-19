---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_instance"
description: |-
  Provides details about a Lightsail Instance
---

# Data Source: aws_lightsail_instance

Provides details about a Lightsail Instance. Amazon Lightsail is a service to provide easy virtual private servers
with custom software already setup. See [What is Amazon Lightsail?](https://lightsail.aws.amazon.com/ls/docs/getting-started/article/what-is-amazon-lightsail)
for more information.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage

```hcl
# Create a new GitLab Lightsail Instance
data "aws_lightsail_instance" "gitlab_test" {
  name = "custom_gitlab"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Lightsail Instance.

## Availability Zones
Lightsail currently supports the following Availability Zones (e.g. `us-east-1a`):

- `ap-northeast-1{a,c,d}`
- `ap-northeast-2{a,c}`
- `ap-south-1{a,b}`
- `ap-southeast-1{a,b,c}`
- `ap-southeast-2{a,b,c}`
- `ca-central-1{a,b}`
- `eu-central-1{a,b,c}`
- `eu-west-1{a,b,c}`
- `eu-west-2{a,b,c}`
- `eu-west-3{a,b,c}`
- `us-east-1{a,b,c,d,e,f}`
- `us-east-2{a,b,c}`
- `us-west-2{a,b,c}`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the Lightsail instance (matches `name`).
* `availability_zone` - The Availability Zone in which your instance resides.
* `blueprint_id` - The ID for a virtual private server image.
* `bundle_id` - The bundle of specification information.
* `key_pair_name` - The name of your key pair.
* `user_data` - launch script to configure server with additional user data
* `tags` - A map of tags assigned to the resource. A key-only tag will show an empty string as the value.
* `arn` - The ARN of the Lightsail instance (matches `id`).
* `created_at` - The timestamp when the instance was created.
* `ipv6_address` - (**Deprecated**) The first IPv6 address of the Lightsail instance. Use `ipv6_addresses` attribute instead.
* `ipv6_addresses` - List of IPv6 addresses for the Lightsail instance.
* `ip_address_type` - The IP address type of the Lightsail Instance.