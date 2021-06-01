---
subcategory: "Elastic Load Balancing (ELB Classic)"
layout: "aws"
page_title: "AWS: aws_load_balancer_listener_policy"
description: |-
  Attaches a load balancer policy to an ELB Listener.
---

# Resource: aws_load_balancer_listener_policy

Attaches a load balancer policy to an ELB Listener.


## Example Usage for Custom Policy

```hcl
resource "aws_elb" "wu-tang" {
  name               = "wu-tang"
  availability_zones = ["us-east-1a"]

  listener {
    instance_port      = 443
    instance_protocol  = "http"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = "arn:aws:iam::000000000000:server-certificate/wu-tang.net"
  }

  tags = {
    Name = "wu-tang"
  }
}

resource "aws_load_balancer_policy" "wu-tang-ssl" {
  load_balancer_name = aws_elb.wu-tang.name
  policy_name        = "wu-tang-ssl"
  policy_type_name   = "SSLNegotiationPolicyType"

  policy_attribute {
    name  = "ECDHE-ECDSA-AES128-GCM-SHA256"
    value = "true"
  }

  policy_attribute {
    name  = "Protocol-TLSv1.2"
    value = "true"
  }
}

resource "aws_load_balancer_listener_policy" "wu-tang-listener-policies-443" {
  load_balancer_name = aws_elb.wu-tang.name
  load_balancer_port = 443

  policy_names = [
    aws_load_balancer_policy.wu-tang-ssl.policy_name,
  ]
}
```

This example shows how to customize the TLS settings of an HTTPS listener.

## Example Usage for AWS Predefined Security Policy

```hcl
resource "aws_elb" "wu-tang" {
  name               = "wu-tang"
  availability_zones = ["us-east-1a"]

  listener {
    instance_port      = 443
    instance_protocol  = "http"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = "arn:aws:iam::000000000000:server-certificate/wu-tang.net"
  }

  tags = {
    Name = "wu-tang"
  }
}

resource "aws_load_balancer_policy" "wu-tang-ssl-tls-1-1" {
  load_balancer_name = aws_elb.wu-tang.name
  policy_name        = "wu-tang-ssl"
  policy_type_name   = "SSLNegotiationPolicyType"

  policy_attribute {
    name  = "Reference-Security-Policy"
    value = "ELBSecurityPolicy-TLS-1-1-2017-01"
  }
}

resource "aws_load_balancer_listener_policy" "wu-tang-listener-policies-443" {
  load_balancer_name = aws_elb.wu-tang.name
  load_balancer_port = 443

  policy_names = [
    aws_load_balancer_policy.wu-tang-ssl-tls-1-1.policy_name,
  ]
}
```

This example shows how to add a [Predefined Security Policy for ELBs](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-security-policy-table.html)

## Argument Reference

The following arguments are supported:

* `load_balancer_name` - (Required) The load balancer to attach the policy to.
* `load_balancer_port` - (Required) The load balancer listener port to apply the policy to.
* `policy_names` - (Required) List of Policy Names to apply to the backend server.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the policy.
* `load_balancer_name` - The load balancer on which the policy is defined.
* `load_balancer_port` - The load balancer listener port the policies are applied to
