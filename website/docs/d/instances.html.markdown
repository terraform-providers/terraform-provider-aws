---
layout: "aws"
page_title: "AWS: aws_instances"
sidebar_current: "docs-aws-datasource-instances"
description: |-
  Get information on an Amazon EC2 instances.
---

# Data Source: aws_instances

Use this data source to get IDs or IPs of Amazon EC2 instances to be referenced elsewhere,
e.g. to allow easier migration from another management solution
or to make it easier for an operator to connect through bastion host(s).

-> **Note:** It's a best practice to expose instance details via [outputs](https://www.terraform.io/docs/configuration/outputs.html)
and [remote state](https://www.terraform.io/docs/state/remote.html) and
**use [`terraform_remote_state`](https://www.terraform.io/docs/providers/terraform/d/remote_state.html)
data source instead** if you manage referenced instances via Terraform.

~> **Note:** It's strongly discouraged to use this data source for querying ephemeral
instances (e.g. managed via autoscaling group), as the output may change at any time
and you'd need to re-run `apply` every time an instance comes up or dies.

## Example Usage

```hcl
data "aws_instances" "test" {
  instance_tags = {
    Role = "HardWorker"
  }

  filter {
    name   = "instance.group-id"
    values = ["sg-12345678"]
  }

  instance_state_names = ["running", "stopped"]
}

resource "aws_eip" "test" {
  count    = "${length(data.aws_instances.test.ids)}"
  instance = "${data.aws_instances.test.ids[count.index]}"
}
```

## Argument Reference

* `instance_tags` - (Optional) A mapping of tags, each pair of which must
exactly match a pair on desired instances.

* `instance_state_names` - (Optional) A list of instance states that should be applicable to the desired instances. The permitted values are: `pending, running, shutting-down, stopped, stopping, terminated`. The default value is `running`.

* `filter` - (Optional) One or more name/value pairs to use as filters. There are
several valid keys, for a full reference, check out
[describe-instances in the AWS CLI reference][1].

## Attributes Reference

* `ids` - IDs of instances found through the filter
* `private_ips` - Private IP addresses of instances found through the filter
* `public_ips` - Public IP addresses of instances found through the filter


[1]: http://docs.aws.amazon.com/cli/latest/reference/ec2/describe-instances.html
