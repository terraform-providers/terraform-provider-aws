---
layout: "aws"
page_title: "AWS: aws_ssm_parameter"
sidebar_current: "docs-aws-resource-ssm-parameter"
description: |-
  Provides a SSM Parameter resource
---

# aws_ssm_parameter

Provides an SSM Parameter resource.

## Example Usage

To store a basic string parameter:

```hcl
resource "aws_ssm_parameter" "foo" {
  name  = "foo"
  type  = "String"
  value = "bar"
}
```

To store an encrypted string using the default SSM KMS key:

```hcl
resource "aws_db_instance" "default" {
  allocated_storage    = 10
  storage_type         = "gp2"
  engine               = "mysql"
  engine_version       = "5.7.16"
  instance_class       = "db.t2.micro"
  name                 = "mydb"
  username             = "foo"
  password             = "${var.database_master_password}"
  db_subnet_group_name = "my_database_subnet_group"
  parameter_group_name = "default.mysql5.7"
}

resource "aws_ssm_parameter" "secret" {
  name        = "/${var.environment}/database/password/master"
  description = "The parameter description"
  type        = "SecureString"
  value       = "${var.database_master_password}"

  tags = {
    environment = "${var.environment}"
  }
}
```

~> **Note:** The unencrypted value of a SecureString will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the parameter. If the name contains a path (e.g. any forward slashes (`/`)), it must be fully qualified with a leading forward slash (`/`). For additional requirements and constraints, see the [AWS SSM User Guide](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-parameter-name-constraints.html).
* `type` - (Required) The type of the parameter. Valid types are `String`, `StringList` and `SecureString`.
* `value` - (Required) The value of the parameter.
* `description` - (Optional) The description of the parameter.
* `key_id` - (Optional) The KMS key id or arn for encrypting a SecureString.
* `overwrite` - (Optional) Overwrite an existing parameter. If not specified, will default to `false` if the resource has not been created by terraform to avoid overwrite of existing resource and will default to `true` otherwise (terraform lifecycle rules should then be used to manage the update behavior).
* `allowed_pattern` - (Optional) A regular expression used to validate the parameter value.
* `tags` - (Optional) A mapping of tags to assign to the object.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the parameter.
* `name` - (Required) The name of the parameter.
* `description` - (Required) The description of the parameter.
* `type` - (Required) The type of the parameter. Valid types are `String`, `StringList` and `SecureString`.
* `value` - (Required) The value of the parameter.

## Import

SSM Parameters can be imported using the `parameter store name`, e.g.

```
$ terraform import aws_ssm_parameter.my_param /my_path/my_paramname
```
