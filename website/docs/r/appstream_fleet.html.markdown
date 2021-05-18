---
subcategory: "AppStream"
layout: "aws"
page_title: "AWS: aws_appstream_fleet"
description: |-
  Provides an AppStream fleet
---

# Resource: aws_appstream_fleet

Provides an AppStream fleet and associates it with a stack.

## Example Usage

```hcl
resource "aws_appstream_fleet" "test_fleet" {
  name       = "test-fleet"
  stack_name = aws_appstream_stack.appstream_stack.name
  compute_capacity {
    desired_instances = 1
  }
  description                    = "test fleet"
  disconnect_timeout             = 15
  display_name                   = "test-fleet"
  enable_default_internet_access = false
  fleet_type                     = "ON_DEMAND"
  image_name                     = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type                  = "stream.standard.large"
  max_user_duration              = 600
  subnet_ids                     = ["subnet-06e9b13400c225127"]
  security_group_ids             = ["sg-0397cdfe509785903", "sg-0bd2dddff01dee52d"]
  tags = {
    TagName = "tag-value"
  }
  state = "RUNNING"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the AppStream fleet, used as the fleet's identifier.  Only allows alphanumeric, hypen, underscore and period.
* `compute_capacity` - (Required) block to specify `desired_instances`.
* `stack_name` - (Optional) Name of AppStream stack, to be associated with this fleet.
* `description` - (Optional) Longer description for the AppStream fleet.
* `disconnect_timeout` - (Optional) disconnect timeout in minutes.
* `display_name` - (Optional) Human-readable friendly name for the AppStream fleet.
* `domain_info` - (Optional) Nested block to specify `directory_name` and `organizational_unit_distinguished_name`.
* `enable_default_internet_access` - (Optional) specifies whether default internet access is allowed.  Requires being in a public subnet.
* `fleet_type` - (Optional) "ON_DEMAND" or "ALWAYS_ON"
* `image_name` - (Required) Name of the AppStream image to use for this fleet.
* `instance_type` - (Required) Type of instance, e.g., "stream.standard.medium" or "stream.standard.large"
* `max_user_duration` - (Optional) Maximum user session duration in minutes.
* `state` - (Optional) "RUNNING" or "STOPPED".
* `security_group_ids` - Security group IDs to attach to AppStream instances.
* `subnet_ids` - VPC subnet IDs in which to create AppStream instances.
* `tags` - Map of tags to attach to AppStream instances.

## Attributes Reference

* `name` - Identifier for this AppStream fleet, as specified in input
