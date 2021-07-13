---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_database"
description: |-
  Provides a Lightsail Database
---

# Resource: aws_lightsail_database

Provides a Lightsail Database. Amazon Lightsail is a service to provide easy virtual private servers
with custom software already setup. See [What is Amazon Lightsail?](https://lightsail.aws.amazon.com/ls/docs/getting-started/article/what-is-amazon-lightsail)
for more information.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage

```terraform
resource "aws_lightsail_database" "test" {
  name                 = "test"
  availability_zone    = "us-east-1a"
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to use for your new Lightsail database resource. Names be unique within each AWS Region in your Lightsail account.
* `availability_zone` - (Required) The Availability Zone in which to create your new database. Use the us-east-2a case-sensitive format. (see list below)
* `master_database_name` - (Required) The name of the master database created when the Lightsail database resource is created.
* `master_password` - (Sensitive) The password for the master user of your new database. The password can include any printable ASCII character except "/", """, or "@".
* `master_username` - The master user name for your new database.
* `blueprint_id` - (Required) The blueprint ID for your new database. A blueprint describes the major engine version of a database. You can get a list of database blueprints IDs by using the AWS CLI command: `aws lightsail get-relational-database-blueprints`
* `bundle_id` - (Required)  The bundle ID for your new database. A bundle describes the performance specifications for your database (see list below). You can get a list of database bundle IDs by using the AWS CLI command: `aws lightsail get-relational-database-bundles`.
* `preferred_backup_window` - The daily time range during which automated backups are created for your new database if automated backups are enabled. Must be in the hh24:mi-hh24:mi format. Example: `16:00-16:30`. Specified in Coordinated Universal Time (UTC).
* `preferred_maintenance_window` - The weekly time range during which system maintenance can occur on your new database. Must be in the ddd:hh24:mi-ddd:hh24:mi format. Specified in Coordinated Universal Time (UTC). Example: `Tue:17:00-Tue:17:30`
* `publicly_accessible` - Specifies the accessibility options for your new database. A value of true specifies a database that is available to resources outside of your Lightsail account. A value of false specifies a database that is available only to your Lightsail resources in the same region as your database.
* `apply_immediately` - When true , applies changes immediately. When false , applies changes during the preferred maintenance window. Some changes may cause an outage.
* `backup_retention_enabled` - When true, enables automated backup retention for your database. When false, disables automated backup retention for your database. Disabling backup retention deletes all automated database backups. Before disabling this, you may want to create a snapshot of your database.
* `skip_final_snapshot` - Determines whether a final database snapshot is created before your database is deleted. If true is specified, no database snapshot is created. If false is specified, a database snapshot is created before your database is deleted. You must specify the final relational database snapshot name parameter if the skip final snapshot parameter is false.
* `final_snapshot_name` - (Required unless `skip_final_snapshot = true`) The name of the database snapshot created if skip final snapshot is false, which is the default value for that parameter.
* `tags` - (Optional) A map of tags to assign to the resource. To create a key-only tag, use an empty string as the value.

## Availability Zones
Lightsail currently supports the following Availability Zones (e.g. `us-east-1a`):

- `ap-northeast-1{a,c,d}`
- `ap-northeast-2{a,c}`
- `ap-south-1{a,b}`
- `ap-southeast-1{a,b,c}`
- `ap-southeast-2{a,b,c}`
- `ca-central-1{a,b}`
- `eu-central-1{a,b,c}`
- `eu-west-1{a,b,c}`
- `eu-west-2{a,b,c}`
- `eu-west-3{a,b,c}`
- `us-east-1{a,b,c,d,e,f}`
- `us-east-2{a,b,c}`
- `us-west-2{a,b,c}`

## Bundles

Lightsail currently supports the following Bundle IDs (e.g. an instance in `ap-northeast-1` would use `small_2_0`):

### Prefix

A Bundle ID starts with one of the below size prefixes:

- `micro_`
- `small_`
- `medium_`
- `large_`

### Infixes (Optional for HA Database)

A Bundle Id can have the following infix added in order to use the HA option of the selected bundle.

- `ha_`

### Suffix

A Bundle ID ends with one of the following suffix: `1_0`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the Lightsail instance (matches `arn`).
* `arn` - The ARN of the Lightsail instance (matches `id`).
* `ca_certificate_identifier` - The certificate associated with the database.
* `created_at` - The timestamp when the instance was created.
* `engine` - The database software (for example, MySQL).
* `engine_version` - The database engine version (for example, 5.7.23).
* `cpu_count` - The number of vCPUs for the database.
* `ram_size` - The amount of RAM in GB for the database.
* `disk_size` - The size of the disk for the database.
* `master_endpoint_port` - The master endpoint network port for the database.
* `master_endpoint_address` - The master endpoint fqdn for the database.
* `secondary_availability_zone` - Describes the secondary Availability Zone of a high availability database. The secondary database is used for failover support of a high availability database.
* `support_code` - The support code for the database. Include this code in your email to support when you have questions about a database in Lightsail. This code enables our support team to look up your Lightsail information more easily.

## Import

Lightsail Databases can be imported using their name, e.g.

```
$ terraform import aws_lightsail_database.foo 'bar'
```
