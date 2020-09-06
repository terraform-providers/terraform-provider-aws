---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_spot_fleet_request"
description: |-
  Provides a Spot Fleet Request resource.
---

# Resource: aws_spot_fleet_request

Provides an EC2 Spot Fleet Request resource. This allows a fleet of Spot
instances to be requested on the Spot market.

## Example Usage

### Using launch specifications

```hcl
# Request a Spot fleet
resource "aws_spot_fleet_request" "cheap_compute" {
  iam_fleet_role      = "arn:aws:iam::12345678:role/spot-fleet"
  spot_price          = "0.03"
  allocation_strategy = "diversified"
  target_capacity     = 6
  valid_until         = "2019-11-04T20:44:20Z"

  launch_specification {
    instance_type            = "m4.10xlarge"
    ami                      = "ami-1234"
    spot_price               = "2.793"
    placement_tenancy        = "dedicated"
    iam_instance_profile_arn = aws_iam_instance_profile.example.arn
  }

  launch_specification {
    instance_type            = "m4.4xlarge"
    ami                      = "ami-5678"
    key_name                 = "my-key"
    spot_price               = "1.117"
    iam_instance_profile_arn = aws_iam_instance_profile.example.arn
    availability_zone        = "us-west-1a"
    subnet_id                = "subnet-1234"
    weighted_capacity        = 35

    root_block_device {
      volume_size = "300"
      volume_type = "gp2"
    }

    tags = {
      Name = "spot-fleet-example"
    }
  }
}
```

### Using launch templates

```hcl
resource "aws_launch_template" "foo" {
  name          = "launch-template"
  image_id      = "ami-516b9131"
  instance_type = "m1.small"
  key_name      = "some-key"
}

resource "aws_spot_fleet_request" "foo" {
  iam_fleet_role  = "arn:aws:iam::12345678:role/spot-fleet"
  spot_price      = "0.005"
  target_capacity = 2
  valid_until     = "2019-11-04T20:44:20Z"

  launch_template_config {
    launch_template_specification {
      id      = aws_launch_template.foo.id
      version = aws_launch_template.foo.latest_version
    }
  }

  depends_on = [aws_iam_policy_attachment.test-attach]
}
```

~> **NOTE:** Terraform does not support the functionality where multiple `subnet_id` or `availability_zone` parameters can be specified in the same
launch configuration block. If you want to specify multiple values, then separate launch configuration blocks should be used or launch template overrides should be configured, one per subnet:

### Using multiple launch specifications

```hcl
resource "aws_spot_fleet_request" "foo" {
  iam_fleet_role  = "arn:aws:iam::12345678:role/spot-fleet"
  spot_price      = "0.005"
  target_capacity = 2
  valid_until     = "2019-11-04T20:44:20Z"

  launch_specification {
    instance_type     = "m1.small"
    ami               = "ami-d06a90b0"
    key_name          = "my-key"
    availability_zone = "us-west-2a"
  }

  launch_specification {
    instance_type     = "m5.large"
    ami               = "ami-d06a90b0"
    key_name          = "my-key"
    availability_zone = "us-west-2a"
  }
}
```


### Using multiple launch configurations

```hcl
data "aws_subnet_ids" "example" {
  vpc_id = var.vpc_id
}

resource "aws_launch_template" "foo" {
  name          = "launch-template"
  image_id      = "ami-516b9131"
  instance_type = "m1.small"
  key_name      = "some-key"
}

resource "aws_spot_fleet_request" "foo" {
  iam_fleet_role  = "arn:aws:iam::12345678:role/spot-fleet"
  spot_price      = "0.005"
  target_capacity = 2
  valid_until     = "2019-11-04T20:44:20Z"

  launch_template_config {
    launch_template_specification {
      id      = aws_launch_template.foo.id
      version = aws_launch_template.foo.latest_version
    }
    overrides {
      subnet_id = data.aws_subnets.example.ids[0]
    }
    overrides {
      subnet_id = data.aws_subnets.example.ids[1]
    }
    overrides {
      subnet_id = data.aws_subnets.example.ids[2]
    }
  }

  depends_on = [aws_iam_policy_attachment.test-attach]
}
```

## Argument Reference

Most of these arguments directly correspond to the
[official API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_SpotFleetRequestConfigData.html).

* `iam_fleet_role` - (Required) Grants the Spot fleet permission to terminate
  Spot instances on your behalf when you cancel its Spot fleet request using
CancelSpotFleetRequests or when the Spot fleet request expires, if you set
terminateInstancesWithExpiration.
* `replace_unhealthy_instances` - (Optional) Indicates whether Spot fleet should replace unhealthy instances. Default `false`.
* `launch_specification` - (Optional) Used to define the launch configuration of the
  spot-fleet request. Can be specified multiple times to define different bids
across different markets and instance types. Conflicts with `launch_template_config`. At least one of `launch_specification` or `launch_template_config` is required.

    **Note:** This takes in similar but not
    identical inputs as [`aws_instance`](instance.html).  There are limitations on
    what you can specify. See the list of officially supported inputs in the
    [reference documentation](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_SpotFleetLaunchSpecification.html). Any normal [`aws_instance`](instance.html) parameter that corresponds to those inputs may be used and it have
    a additional parameter `iam_instance_profile_arn` takes `aws_iam_instance_profile` attribute `arn` as input.

* `launch_template_config` - (Optional) Launch template configuration block. See [Launch Template Configs](#launch-template-configs) below for more details. Conflicts with `launch_specification`. At least one of `launch_specification` or `launch_template_config` is required.
* `spot_price` - (Optional; Default: On-demand price) The maximum bid price per unit hour.
* `wait_for_fulfillment` - (Optional; Default: false) If set, Terraform will
  wait for the Spot Request to be fulfilled, and will throw an error if the
  timeout of 10m is reached.
* `target_capacity` - The number of units to request. You can choose to set the
  target capacity in terms of instances or a performance characteristic that is
  important to your application workload, such as vCPUs, memory, or I/O.
* `allocation_strategy` - Indicates how to allocate the target capacity across
  the Spot pools specified by the Spot fleet request. The default is
  `lowestPrice`.
* `instance_pools_to_use_count` - (Optional; Default: 1)
  The number of Spot pools across which to allocate your target Spot capacity.
  Valid only when `allocation_strategy` is set to `lowestPrice`. Spot Fleet selects
  the cheapest Spot pools and evenly allocates your target Spot capacity across
  the number of Spot pools that you specify.
* `excess_capacity_termination_policy` - Indicates whether running Spot
  instances should be terminated if the target capacity of the Spot fleet
  request is decreased below the current size of the Spot fleet.
* `terminate_instances_with_expiration` - Indicates whether running Spot
  instances should be terminated when the Spot fleet request expires.
* `instance_interruption_behaviour` - (Optional) Indicates whether a Spot
  instance stops or terminates when it is interrupted. Default is
  `terminate`.
* `fleet_type` - (Optional) The type of fleet request. Indicates whether the Spot Fleet only requests the target
  capacity or also attempts to maintain it. Default is `maintain`.
* `valid_until` - (Optional) The end date and time of the request, in UTC [RFC3339](https://tools.ietf.org/html/rfc3339#section-5.8) format(for example, YYYY-MM-DDTHH:MM:SSZ). At this point, no new Spot instance requests are placed or enabled to fulfill the request.
* `valid_from` - (Optional) The start date and time of the request, in UTC [RFC3339](https://tools.ietf.org/html/rfc3339#section-5.8) format(for example, YYYY-MM-DDTHH:MM:SSZ). The default is to start fulfilling the request immediately.
* `load_balancers` (Optional) A list of elastic load balancer names to add to the Spot fleet.
* `target_group_arns` (Optional) A list of `aws_alb_target_group` ARNs, for use with Application Load Balancing.
* `tags` - (Optional) A map of tags to assign to the resource.

### Launch Template Configs

The `launch_template_config` block supports the following:

* `launch_template_specification` - (Required) Launch template specification. See [Launch Template Specification](#launch-template-specification) below for more details.
* `overrides` - (Optional) One or more override configurations. See [Overrides](#overrides) below for more details.

### Launch Template Specification

* `id` - The ID of the launch template. Conflicts with `name`.
* `name` - The name of the launch template. Conflicts with `id`.
* `version` - (Optional) Template version. Unlike the autoscaling equivalent, does not support `$Latest` or `$Default`, so use the launch_template resource's attribute, e.g. `"${aws_launch_template.foo.latest_version}"`. It will use the default version if omitted.

    **Note:** The specified launch template can specify only a subset of the
    inputs of [`aws_launch_template`](launch_template.html).  There are limitations on
    what you can specify as spot fleet does not support all the attributes that are supported by autoscaling groups. [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-launch-templates.html#launch-templates-spot-fleet) is currently sparse, but at least `instance_initiated_shutdown_behavior` is confirmed unsupported.

### Overrides

* `availability_zone` - (Optional) The availability zone in which to place the request.
* `instance_type` - (Optional) The type of instance to request.
* `priority` - (Optional) The priority for the launch template override. The lower the number, the higher the priority. If no number is set, the launch template override has the lowest priority.
* `spot_price` - (Optional) The maximum spot bid for this override request.
* `subnet_id` - (Optional) The subnet in which to launch the requested instance.
* `weighted_capacity` - (Optional) The capacity added to the fleet by a fulfilled request.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 10 mins) Used when requesting the spot instance (only valid if `wait_for_fulfillment = true`)
* `delete` - (Defaults to 15 mins) Used when destroying the spot instance

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Spot fleet request ID
* `spot_request_state` - The state of the Spot fleet request.

## Import

Spot Fleet Requests can be imported using `id`, e.g.

```
$ terraform import aws_spot_fleet_request.fleet sfr-005e9ec8-5546-4c31-b317-31a62325411e
```
