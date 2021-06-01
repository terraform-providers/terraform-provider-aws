---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_db_instance_role_association"
description: |-
  Manages an RDS DB Instance association with an IAM Role.
---

# Resource: aws_db_instance_role_association

Manages an RDS DB Instance association with an IAM Role. Example use cases:

* [Amazon RDS Oracle integration with Amazon S3](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/oracle-s3-integration.html)
* [Importing Amazon S3 Data into an RDS PostgreSQL DB Instance](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PostgreSQL.S3Import.html)

-> To manage the RDS DB Instance IAM Role for [Enhanced Monitoring](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Monitoring.OS.html), see the `aws_db_instance` resource `monitoring_role_arn` argument instead.

## Example Usage

```hcl
resource "aws_db_instance_role_association" "example" {
  db_instance_identifier = aws_db_instance.example.id
  feature_name           = "S3_INTEGRATION"
  role_arn               = aws_iam_role.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `db_instance_identifier` - (Required) DB Instance Identifier to associate with the IAM Role.
* `feature_name` - (Required) Name of the feature for association. This can be found in the AWS documentation relevant to the integration or a full list is available in the `SupportedFeatureNames` list returned by [AWS CLI rds describe-db-engine-versions](https://docs.aws.amazon.com/cli/latest/reference/rds/describe-db-engine-versions.html).
* `role_arn` - (Required) Amazon Resource Name (ARN) of the IAM Role to associate with the DB Instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - DB Instance Identifier and IAM Role ARN separated by a comma (`,`)

## Import

`aws_db_instance_role_association` can be imported using the DB Instance Identifier and IAM Role ARN separated by a comma (`,`), e.g.

```
$ terraform import aws_db_instance_role_association.example my-db-instance,arn:aws:iam::123456789012:role/my-role
```
