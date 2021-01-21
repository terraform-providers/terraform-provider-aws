---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_orderable_db_instance"
description: |-
  Information about Neptune orderable DB instances.
---

# Data Source: aws_neptune_orderable_db_instance

Information about Neptune orderable DB instances.

## Example Usage

```hcl
data "aws_neptune_orderable_db_instance" "test" {
  engine_version             = "1.0.3.0"
  preferred_instance_classes = ["db.r5.large", "db.r4.large", "db.t3.medium"]
}
```

## Argument Reference

The following arguments are supported:

* `engine` - (Optional) DB engine. (Default: `neptune`)
* `engine_version` - (Optional) Version of the DB engine. For example, `1.0.1.0`, `1.0.1.2`, `1.0.2.2`, and `1.0.3.0`.
* `instance_class` - (Optional) DB instance class. Examples of classes are `db.r5.large`, `db.r5.xlarge`, `db.r4.large`, `db.r5.4xlarge`, `db.r5.12xlarge`, `db.r4.xlarge`, and `db.t3.medium`.
* `license_model` - (Optional) License model. (Default: `amazon-license`)
* `preferred_instance_classes` - (Optional) Ordered list of preferred Neptune DB instance classes. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned.
* `vpc` - (Optional) Enable to show only VPC offerings.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `availability_zones` - Availability zones where the instance is available.
* `max_iops_per_db_instance` - Maximum total provisioned IOPS for a DB instance.
* `max_iops_per_gib` - Maximum provisioned IOPS per GiB for a DB instance.
* `max_storage_size` - Maximum storage size for a DB instance.
* `min_iops_per_db_instance` - Minimum total provisioned IOPS for a DB instance.
* `min_iops_per_gib` - Minimum provisioned IOPS per GiB for a DB instance.
* `min_storage_size` - Minimum storage size for a DB instance.
* `multi_az_capable` - Whether a DB instance is Multi-AZ capable.
* `read_replica_capable` - Whether a DB instance can have a read replica.
* `storage_type` - The storage type for a DB instance.
* `supports_enhanced_monitoring` - Whether a DB instance supports Enhanced Monitoring at intervals from 1 to 60 seconds.
* `supports_iam_database_authentication` - Whether a DB instance supports IAM database authentication.
* `supports_iops` - Whether a DB instance supports provisioned IOPS.
* `supports_performance_insights` - Whether a DB instance supports Performance Insights.
* `supports_storage_encryption` - Whether a DB instance supports encrypted storage.
