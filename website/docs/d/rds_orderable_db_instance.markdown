---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_rds_orderable_db_instance"
description: |-
  Information about RDS orderable DB instances.
---

# Data Source: aws_rds_orderable_db_instance

Information about RDS orderable DB instances and valid parameter combinations.

## Example Usage

```hcl
data "aws_rds_orderable_db_instance" "test" {
  engine         = "mysql"
  engine_version = "5.7.22"
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = ["db.r6.xlarge", "db.m4.large", "db.t3.small"]
}
```

Valid parameter combinations can also be found with `preferred_engine_versions` and/or `preferred_instance_classes`.

```hcl
data "aws_rds_orderable_db_instance" "test" {
  engine        = "mysql"
  license_model = "general-public-license"

  preferred_engine_versions  = ["5.6.35", "5.6.41", "5.6.44"]
  preferred_instance_classes = ["db.t2.small", "db.t3.medium", "db.t3.large"]
}
```


## Argument Reference

The following arguments are supported:

* `availability_zone_group` - (Optional) Availability zone group.
* `engine` - (Required) DB engine. Engine values include `aurora`, `aurora-mysql`, `aurora-postgresql`, `docdb`, `mariadb`, `mysql`, `neptune`, `oracle-ee`, `oracle-se`, `oracle-se1`, `oracle-se2`, `postgres`, `sqlserver-ee`, `sqlserver-ex`, `sqlserver-se`, and `sqlserver-web`.
* `engine_version` - (Optional) Version of the DB engine. If none is provided, the AWS-defined default version will be used.
* `instance_class` - (Optional) DB instance class. Examples of classes are `db.m3.2xlarge`, `db.t2.small`, and `db.m3.medium`.
* `license_model` - (Optional) License model. Examples of license models are `general-public-license`, `bring-your-own-license`, and `amazon-license`.
* `preferred_instance_classes` - (Optional) Ordered list of preferred RDS DB instance classes. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned.
* `preferred_engine_versions` - (Optional) Ordered list of preferred RDS DB instance engine versions. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned.
* `storage_type` - (Optional) Storage types. Examples of storage types are `standard`, `io1`, `gp2`, and `aurora`.
* `supports_enhanced_monitoring` - (Optional) Enable this to ensure a DB instance supports Enhanced Monitoring at intervals from 1 to 60 seconds.
* `supports_global_databases` - (Optional) Enable this to ensure a DB instance supports Aurora global databases with a specific combination of other DB engine attributes.
* `supports_iam_database_authentication` - (Optional) Enable this to ensure a DB instance supports IAM database authentication.
* `supports_iops` - (Optional) Enable this to ensure a DB instance supports provisioned IOPS.
* `supports_kerberos_authentication` - (Optional) Enable this to ensure a DB instance supports Kerberos Authentication.
* `supports_performance_insights` - (Optional) Enable this to ensure a DB instance supports Performance Insights.
* `supports_storage_autoscaling` - (Optional) Enable this to ensure Amazon RDS can automatically scale storage for DB instances that use the specified DB instance class.
* `supports_storage_encryption` - (Optional) Enable this to ensure a DB instance supports encrypted storage.
* `vpc` - (Optional) Boolean that indicates whether to show only VPC or non-VPC offerings.

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
* `outpost_capable` - Whether a DB instance supports RDS on Outposts.
* `read_replica_capable` - Whether a DB instance can have a read replica.
* `supported_engine_modes` - A list of the supported DB engine modes.
