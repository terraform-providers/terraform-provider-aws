---
layout: "aws"
page_title: "AWS: aws_lightsail_instance"
sidebar_current: "docs-aws-resource-lightsail-instance"
description: |-
  Provides an Lightsail Instance
---

# aws_lightsail_instance

Provides a Lightsail Instance. Amazon Lightsail is a service to provide easy virtual private servers
with custom software already setup. See [What is Amazon Lightsail?](https://lightsail.aws.amazon.com/ls/docs/getting-started/article/what-is-amazon-lightsail)
for more information.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage

```hcl
# Create a new GitLab Lightsail Instance
resource "aws_lightsail_instance" "gitlab_test" {
  name              = "custom gitlab"
  availability_zone = "us-east-1b"
  blueprint_id      = "string"
  bundle_id         = "string"
  key_pair_name     = "some_key_name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Lightsail Instance
* `availability_zone` - (Required) The Availability Zone in which to create your
instance (see list below)
* `blueprint_id` - (Required) The ID for a virtual private server image
(see list below)
* `bundle_id` - (Required) The bundle of specification information (see list below)
* `key_pair_name` - (Required) The name of your key pair. Created in the
Lightsail console (cannot use `aws_key_pair` at this time)
* `user_data` - (Optional) launch script to configure server with additional user data

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

## Blueprints

Lightsail currently supports the following Blueprint IDs:

### OS Only

- `amazon_linux_2018_03_0_2`
- `centos_7_1805_01`
- `debian_8_7`
- `debian_9_5`
- `freebsd_11_1`
- `opensuse_42_2`
- `ubuntu_16_04_2`
- `ubuntu_18_04`

### Apps and OS

- `drupal_8_5_6`
- `gitlab_11_1_4_1`
- `joomla_3_8_11`
- `lamp_5_6_37_2`
- `lamp_7_1_20_1`
- `magento_2_2_5`
- `mean_4_0_1`
- `nginx_1_14_0_1`
- `nodejs_10_8_0`
- `plesk_ubuntu_17_8_11_1`
- `redmine_3_4_6`
- `wordpress_4_9_8`
- `wordpress_multisite_4_9_8`

## Bundles

Lightsail currently supports the following Bundle IDs (e.g. an instance in `ap-northeast-1` would use `small_2_0`):

### Prefix

A Bundle ID starts with one of the below size prefixes:

- `nano_`
- `micro_`
- `small_`
- `medium_`
- `large_`
- `xlarge_`
- `2xlarge_`

### Suffix

A Bundle ID ends with one of the following suffixes depending on Availability Zone:

- ap-northeast-1: `2_0`
- ap-northeast-2: `2_0`
- ap-south-1: `2_1`
- ap-southeast-1: `2_0`
- ap-southeast-2: `2_2`
- ca-central-1: `2_0`
- eu-central-1: `2_0`
- eu-west-1: `2_0`
- eu-west-2: `2_0`
- eu-west-3: `2_0`
- us-east-1: `2_0`
- us-east-2: `2_0`
- us-west-2: `2_0`

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ARN of the Lightsail instance (matches `arn`).
* `arn` - The ARN of the Lightsail instance (matches `id`).
* `availability_zone`
* `blueprint_id`
* `bundle_id`
* `key_pair_name`
* `user_data`

## Import

Lightsail Instances can be imported using their name, e.g.

```
$ terraform import aws_lightsail_instance.gitlab_test 'custom gitlab'
```
