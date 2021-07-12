---
subcategory: "Database Migration Service (DMS)"
layout: "aws"
page_title: "AWS: aws_dms_replication_instance"
description: |-
  Get information on an a DMS (Data Migration Service) replication instance.
---

# Data Source: aws_dms_replication_instance

Use this data source to get the ID of an Amazon DMS (Data Migration Service) replication instance for use in other resources.

## Example Usage

```hcl
data "aws_dms_replication_instance" "foo" {
  filter {
    name   = "instance-id"
    values = ["i-instanceid"]
  }
}
```

## Argument Reference

* `filter` - One or more name/value pairs to use as filters. There are
several valid keys, for a full reference, check out
[describe-instances in the AWS CLI reference][1].

~> **NOTE:** If anything other than a single match is returned by the search,
Terraform will fail. Ensure that your search is specific enough to return
a single Instance ID only.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `replication_instance_arn` - The ARN of the instance.
* `replication_instance_private_ips` - One or more private IP addresses for the replication instance.
* `replication_instance_public_ips` - One or more public IP addresses for the replication instance.

[1]: https://docs.aws.amazon.com/cli/latest/reference/dms/describe-replication-instances.html
