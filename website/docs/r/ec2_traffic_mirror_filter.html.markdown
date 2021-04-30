---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_traffic_mirror_filter"
description: |-
  Provides an Traffic mirror filter
---

# Resource: aws_ec2_traffic_mirror_filter

Provides an Traffic mirror filter.  
Read [limits and considerations](https://docs.aws.amazon.com/vpc/latest/mirroring/traffic-mirroring-considerations.html) for traffic mirroring

## Example Usage

To create a basic traffic mirror filter

```terraform
resource "aws_ec2_traffic_mirror_filter" "foo" {
  description      = "traffic mirror filter - terraform example"
  network_services = ["amazon-dns"]
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional, Forces new resource) A description of the filter.
* `network_services` - (Optional) List of amazon network services that should be mirrored. Valid values: `amazon-dns`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the traffic mirror filter.
* `id` - The name of the filter.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Traffic mirror filter can be imported using the `id`, e.g.

```
$ terraform import aws_ec2_traffic_mirror_filter.foo tmf-0fbb93ddf38198f64
```
