---
subcategory: "Route53"
layout: "aws"
page_title: "AWS: aws_route53_traffic_policy_instance"
description: |-
  Provides a Route53 traffic policy instance resource.
---

# Resource: aws_route53_traffic_policy_instance

Provides a Route53 traffic policy instance resource.

## Example Usage

```hcl
resource "aws_route53_traffic_policy_instance" "test" {
  name                   = "test.example.com"
  traffic_policy_id      = "b3gb108f-ea6f-45a5-baab-9d112d8b4037"
  traffic_policy_version = 1
  hosted_zone_id         = "Z033120931TAQO548OGJC"
  ttl                    = 360
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The domain name for which Amazon Route 53 responds to DNS queries by using the resource record sets that Route 53 creates for this traffic policy instance.
* `traffic_policy_id` - (Required) The ID of the traffic policy that you want to use to create resource record sets in the specified hosted zone.
* `traffic_policy_version` - (Required) The version of the traffic policy
* `hosted_zone_id` - (Required) The ID of the hosted zone
* `ttl` - (Optional) The TTL that you want Amazon Route 53 to assign to all of the resource record sets that it creates in the specified hosted zone.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Id of created traffic policy instance.

## Import

Route53 traffic policy instance can be imported using its id.

```
$ terraform import aws_route53_traffic_policy_instance.test df579d9a-6396-410e-ac22-e7ad60cf9e7e
```
