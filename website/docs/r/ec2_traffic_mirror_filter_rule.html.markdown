---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_traffic_mirror_filter_rule"
description: |-
  Provides an Traffic mirror filter rule
---

# Resource: aws_ec2_traffic_mirror_filter_rule

Provides an Traffic mirror filter rule.  
Read [limits and considerations](https://docs.aws.amazon.com/vpc/latest/mirroring/traffic-mirroring-considerations.html) for traffic mirroring

## Example Usage

To create a basic traffic mirror session

```hcl
resource "aws_ec2_traffic_mirror_filter" "filter" {
  description      = "traffic mirror filter - terraform example"
  network_services = ["amazon-dns"]
}

resource "aws_ec2_traffic_mirror_filter_rule" "ruleout" {
  description              = "test rule"
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.filter.id
  destination_cidr_block   = "10.0.0.0/8"
  source_cidr_block        = "10.0.0.0/8"
  rule_number              = 1
  rule_action              = "accept"
  traffic_direction        = "egress"
}

resource "aws_ec2_traffic_mirror_filter_rule" "rulein" {
  description              = "test rule"
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.filter.id
  destination_cidr_block   = "10.0.0.0/8"
  source_cidr_block        = "10.0.0.0/8"
  rule_number              = 1
  rule_action              = "accept"
  traffic_direction        = "ingress"
  protocol                 = 6

  destination_port_range {
    from_port = 22
    to_port   = 53
  }

  source_port_range {
    from_port = 0
    to_port   = 10
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) Description of the traffic mirror filter rule.
* `traffic_mirror_filter_id`  - (Required) ID of the traffic mirror filter to which this rule should be added
* `destination_cidr_block` - (Required) Destination CIDR block to assign to the Traffic Mirror rule.
* `destination_port_range` - (Optional) Destination port range. Supported only when the protocol is set to TCP(6) or UDP(17). See Traffic mirror port range documented below
* `protocol` - (Optional) Protocol number, for example 17 (UDP), to assign to the Traffic Mirror rule. For information about the protocol value, see [Protocol Numbers](https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml) on the Internet Assigned Numbers Authority (IANA) website.
* `rule_action` - (Required) Action to take (accept | reject) on the filtered traffic. Valid values are `accept` and `reject`
* `rule_number` - (Required) Number of the Traffic Mirror rule. This number must be unique for each Traffic Mirror rule in a given direction. The rules are processed in ascending order by rule number.
* `source_cidr_block` - (Required) Source CIDR block to assign to the Traffic Mirror rule.
* `source_port_range` - (Optional) Source port range. Supported only when the protocol is set to TCP(6) or UDP(17). See Traffic mirror port range documented below
* `traffic_direction` - (Required) Direction of traffic to be captured. Valid values are `ingress` and `egress`

Traffic mirror port range support following attributes:

* `from_port` - (Optional) Starting port of the range
* `to_port` - (Optional) Ending port of the range

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the traffic mirror filter rule.
* `id` - Name of the traffic mirror filter rule.

## Import

Traffic mirror rules can be imported using the `traffic_mirror_filter_id` and `id` separated by `:` e.g.

```
$ terraform import aws_ec2_traffic_mirror_filter_rule.rule tmf-0fbb93ddf38198f64:tmfr-05a458f06445d0aee
```
