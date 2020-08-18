---
layout: "aws"
page_title: "AWS: aws_lightsail_disk"
sidebar_current: "docs-aws-resource-lightsail-disk"
description: |-
  Provides a Lightsail Disk
---

# Resource: aws_lightsail_instance

Provides a Lightsail Disk. Amazon Lightsail is a service to provide easy virtual private servers
with custom software already setup. See [What is Amazon Lightsail?](https://lightsail.aws.amazon.com/ls/docs/getting-started/article/what-is-amazon-lightsail)
for more information.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage

```hcl
# Create a new Lightsail Disk
resource "aws_lightsail_disk" "test" {
    availability_zone = "eu-west-1a"
    name = "test"
    size = 8
    tags = {
       foo = "bar"
    }
}

```

## Argument Reference

The following arguments are supported:
* `name` - (Required) The name of the Lightsail disk
* `availability_zone` - (Required) The Availability Zone in which to create your disk
* `size` - (Required) Size in gigabytes
* `tags` - (Optional) A mapping of tags to assign to the resource.

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			
## Attributes Reference

In addition to all arguments above, the following attributes are exported:
* `created_at` - The timestamp when disk was created