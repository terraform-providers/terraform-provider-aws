---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_rds_engine_version"
description: |-
  Information about an RDS engine version.
---

# Data Source: aws_rds_engine_version

Information about an RDS engine version.

## Example Usage

```hcl
data "aws_rds_engine_version" "test" {
  engine             = "mysql"
  preferred_versions = ["5.7.42", "5.7.19", "5.7.17"]
}
```

## Argument Reference

The following arguments are supported:

* `engine` - (Required) DB engine. Engine values include `aurora`, `aurora-mysql`, `aurora-postgresql`, `docdb`, `mariadb`, `mysql`, `neptune`, `oracle-ee`, `oracle-se`, `oracle-se1`, `oracle-se2`, `postgres`, `sqlserver-ee`, `sqlserver-ex`, `sqlserver-se`, and `sqlserver-web`.
* `parameter_group_family` - (Optional) The name of a specific DB parameter group family. Examples of parameter group families are `mysql8.0`, `mariadb10.4`, and `postgres12`.
* `preferred_versions` - (Optional) Ordered list of preferred engine versions. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned. If both the `version` and `preferred_versions` arguments are not configured, the data source will return the default version for the engine.
* `version` - (Optional) Version of the DB engine. For example, `5.7.22`, `10.1.34`, and `12.3`. If both the `version` and `preferred_versions` arguments are not configured, the data source will return the default version for the engine.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `default_character_set` - The default character set for new instances of this engine version.
* `engine_description` - The description of the database engine.
* `exportable_log_types` - Set of log types that the database engine has available for export to CloudWatch Logs.
* `status` - The status of the DB engine version, either available or deprecated.
* `supported_character_sets` - Set of the character sets supported by this engine.
* `supported_feature_names` - Set of features supported by the DB engine.
* `supported_modes` - Set of the supported DB engine modes.
* `supported_timezones` - Set of the time zones supported by this engine.
* `supports_global_databases` - Indicates whether you can use Aurora global databases with a specific DB engine version.
* `supports_log_exports_to_cloudwatch` - Indicates whether the engine version supports exporting the log types specified by `exportable_log_types` to CloudWatch Logs.
* `supports_parallel_query` - Indicates whether you can use Aurora parallel query with a specific DB engine version.
* `supports_read_replica` - Indicates whether the database engine version supports read replicas.
* `valid_upgrade_targets` - Set of engine versions that this database engine version can be upgraded to.
* `version_description` - The description of the database engine version.
