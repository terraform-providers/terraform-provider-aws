---
subcategory: "Elastic Load Balancing v2 (ALB/NLB)"
layout: "aws"
page_title: "AWS: aws_lb"
description: |-
  Provides a Load Balancer resource.
---

# Resource: aws_lb

Provides a Load Balancer resource.

~> **Note:** `aws_alb` is known as `aws_lb`. The functionality is identical.

## Example Usage

### Application Load Balancer

```hcl
resource "aws_lb" "test" {
  name               = "test-lb-tf"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.lb_sg.id]
  subnets            = aws_subnet.public.*.id

  enable_deletion_protection = true

  access_logs {
    bucket  = aws_s3_bucket.lb_logs.bucket
    prefix  = "test-lb"
    enabled = true
  }

  tags = {
    Environment = "production"
  }
}
```

### Network Load Balancer

```hcl
resource "aws_lb" "test" {
  name               = "test-lb-tf"
  internal           = false
  load_balancer_type = "network"
  subnets            = aws_subnet.public.*.id

  enable_deletion_protection = true

  tags = {
    Environment = "production"
  }
}
```

### Specifying Elastic IPs

```hcl
resource "aws_lb" "example" {
  name               = "example"
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id     = aws_subnet.example1.id
    allocation_id = aws_eip.example1.id
  }

  subnet_mapping {
    subnet_id     = aws_subnet.example2.id
    allocation_id = aws_eip.example2.id
  }
}
```

### Specifying private IP addresses for an internal-facing load balancer

```hcl
resource "aws_lb" "example" {
  name               = "example"
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id            = aws_subnet.example1.id
    private_ipv4_address = "10.0.1.15"
  }

  subnet_mapping {
    subnet_id            = aws_subnet.example2.id
    private_ipv4_address = "10.0.2.15"
  }
}
```

## Argument Reference

~> **NOTE:** Please note that internal LBs can only use `ipv4` as the ip_address_type. You can only change to `dualstack` ip_address_type if the selected subnets are IPv6 enabled.

~> **NOTE:** Please note that one of either `subnets` or `subnet_mapping` is required.

The following arguments are supported:

* `name` - (Optional) The name of the LB. This name must be unique within your AWS account, can have a maximum of 32 characters,
must contain only alphanumeric characters or hyphens, and must not begin or end with a hyphen. If not specified,
Terraform will autogenerate a name beginning with `tf-lb`.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `internal` - (Optional) If true, the LB will be internal.
* `load_balancer_type` - (Optional) The type of load balancer to create. Possible values are `application`, `gateway`, or `network`. The default value is `application`.
* `security_groups` - (Optional) A list of security group IDs to assign to the LB. Only valid for Load Balancers of type `application`.
* `drop_invalid_header_fields` - (Optional) Indicates whether HTTP headers with header fields that are not valid are removed by the load balancer (true) or routed to targets (false). The default is false. Elastic Load Balancing requires that message header names contain only alphanumeric characters and hyphens. Only valid for Load Balancers of type `application`.
* `access_logs` - (Optional) An Access Logs block. Access Logs documented below.
* `subnets` - (Optional) A list of subnet IDs to attach to the LB. Subnets
cannot be updated for Load Balancers of type `network`. Changing this value
for load balancers of type `network` will force a recreation of the resource.
* `subnet_mapping` - (Optional) A subnet mapping block as documented below.
* `idle_timeout` - (Optional) The time in seconds that the connection is allowed to be idle. Only valid for Load Balancers of type `application`. Default: 60.
* `enable_deletion_protection` - (Optional) If true, deletion of the load balancer will be disabled via
   the AWS API. This will prevent Terraform from deleting the load balancer. Defaults to `false`.
* `enable_cross_zone_load_balancing` - (Optional) If true, cross-zone load balancing of the load balancer will be enabled.
   This is a `network` load balancer feature. Defaults to `false`.
* `enable_http2` - (Optional) Indicates whether HTTP/2 is enabled in `application` load balancers. Defaults to `true`.
* `customer_owned_ipv4_pool` - (Optional) The ID of the customer owned ipv4 pool to use for this load balancer.
* `ip_address_type` - (Optional) The type of IP addresses used by the subnets for your load balancer. The possible values are `ipv4` and `dualstack`
* `tags` - (Optional) A map of tags to assign to the resource.

Access Logs (`access_logs`) support the following:

* `bucket` - (Required) The S3 bucket name to store the logs in.
* `prefix` - (Optional) The S3 bucket prefix. Logs are stored in the root if not configured.
* `enabled` - (Optional) Boolean to enable / disable `access_logs`. Defaults to `false`, even when `bucket` is specified.

Subnet Mapping (`subnet_mapping`) blocks support the following:

* `subnet_id` - (Required) The id of the subnet of which to attach to the load balancer. You can specify only one subnet per Availability Zone.
* `allocation_id` - (Optional) The allocation ID of the Elastic IP address.
* `private_ipv4_address` - (Optional) A private ipv4 address within the subnet to assign to the internal-facing load balancer.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the load balancer (matches `arn`).
* `arn` - The ARN of the load balancer (matches `id`).
* `arn_suffix` - The ARN suffix for use with CloudWatch Metrics.
* `dns_name` - The DNS name of the load balancer.
* `zone_id` - The canonical hosted zone ID of the load balancer (to be used in a Route 53 Alias record).
* `subnet_mapping.*.outpost_id` - ID of the Outpost containing the load balancer.

## Timeouts

`aws_lb` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `create` - (Default `10 minutes`) Used for Creating LB
- `update` - (Default `10 minutes`) Used for LB modifications
- `delete` - (Default `10 minutes`) Used for destroying LB

## Import

LBs can be imported using their ARN, e.g.

```
$ terraform import aws_lb.bar arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188
```
