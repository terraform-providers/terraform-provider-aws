---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_parameter_group"
description: |-
  Provides a Redshift Cluster parameter group resource.
---

# Resource: aws_redshift_parameter_group

Provides a Redshift Cluster parameter group resource.

## Example Usage

```terraform
resource "aws_redshift_parameter_group" "bar" {
  name   = "parameter-group-test-terraform"
  family = "redshift-1.0"

  parameter {
    name  = "require_ssl"
    value = "true"
  }

  parameter {
    name  = "query_group"
    value = "example"
  }

  parameter {
    name  = "enable_user_activity_logging"
    value = "true"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Redshift parameter group.
* `family` - (Required) The family of the Redshift parameter group.
* `description` - (Optional) The description of the Redshift parameter group. Defaults to "Managed by Terraform".
* `parameter` - (Optional) A list of Redshift parameters to apply.

Parameter blocks support the following:

* `name` - (Required) The name of the Redshift parameter.
* `value` - (Required) The value of the Redshift parameter.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

You can read more about the parameters that Redshift supports in the [documentation](http://docs.aws.amazon.com/redshift/latest/mgmt/working-with-parameter-groups.html)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of parameter group
* `id` - The Redshift parameter group name.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Redshift Parameter Groups can be imported using the `name`, e.g.

```
$ terraform import aws_redshift_parameter_group.paramgroup1 parameter-group-test-terraform
```
