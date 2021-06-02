---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_constraint"
description: |-
  Manages a Service Catalog Constraint
---

# Resource: aws_servicecatalog_constraint

Manages a Service Catalog Constraint.

~> **NOTE:** This resource does not associate a Service Catalog product and portfolio. However, the product and portfolio must be associated (see the `aws_servicecatalog_product_portfolio_association` resource) prior to creating a constraint or you will receive an error.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_constraint" "example" {
  description  = "Back off, man. I'm a scientist."
  portfolio_id = aws_servicecatalog_portfolio.example.id
  product_id   = aws_servicecatalog_product.example.id
  type         = "LAUNCH"

  parameters = jsonencode({
    "RoleArn" : "arn:aws:iam::123456789012:role/LaunchRole"
  })
}
```

## Argument Reference

The following arguments are required:

* `parameters` - (Required) Constraint parameters in JSON format. The syntax depends on the constraint type. See details below.
* `portfolio_id` - (Required) Portfolio identifier.
* `product_id` - (Required) Product identifier.
* `type` - (Required) Type of constraint. Valid values are `LAUNCH`, `NOTIFICATION`, `RESOURCE_UPDATE`, `STACKSET`, and `TEMPLATE`.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.
* `description` - (Optional) Description of the constraint.

### `parameters`

The `type` you specify determines what must be included in the `parameters` JSON:

* `LAUNCH`: You are required to specify either the RoleArn or the LocalRoleName but can't use both. If you specify the `LocalRoleName` property, when an account uses the launch constraint, the IAM role with that name in the account will be used. This allows launch-role constraints to be account-agnostic so the administrator can create fewer resources per shared account. The given role name must exist in the account used to create the launch constraint and the account of the user who launches a product with this launch constraint. You cannot have both a `LAUNCH` and a `STACKSET` constraint. You also cannot have more than one `LAUNCH` constraint on an `aws_servicecatalog_product` and `aws_servicecatalog_portfolio`. Specify the `RoleArn` and `LocalRoleName` properties as follows:

```json
{ "RoleArn" : "arn:aws:iam::123456789012:role/LaunchRole" }
```

```json
{ "LocalRoleName" : "SCBasicLaunchRole" }
```

* `NOTIFICATION`: Specify the `NotificationArns` property as follows:

```json
{ "NotificationArns" : ["arn:aws:sns:us-east-1:123456789012:Topic"] }
```

* `RESOURCE_UPDATE`: Specify the `TagUpdatesOnProvisionedProduct` property as follows. The `TagUpdatesOnProvisionedProduct` property accepts a string value of `ALLOWED` or `NOT_ALLOWED`.

```json
{ "Version" : "2.0","Properties" :{ "TagUpdateOnProvisionedProduct" : "String" }}
```

* `STACKSET`: Specify the Parameters property as follows. You cannot have both a `LAUNCH` and a `STACKSET` constraint. You also cannot have more than one `STACKSET` constraint on on an `aws_servicecatalog_product` and `aws_servicecatalog_portfolio`. Products with a `STACKSET` constraint will launch an AWS CloudFormation stack set.

```json
{ "Version" : "String", "Properties" : { "AccountList" : [ "String" ], "RegionList" : [ "String" ], "AdminRole" : "String", "ExecutionRole" : "String" }}
```

* `TEMPLATE`: Specify the Rules property. For more information, see [Template Constraint Rules](http://docs.aws.amazon.com/servicecatalog/latest/adminguide/reference-template_constraint_rules.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Constraint identifier.
* `owner` - Owner of the constraint.

## Import

`aws_servicecatalog_constraint` can be imported using the constraint ID, e.g.

```
$ terraform import aws_servicecatalog_constraint.example cons-nmdkb6cgxfcrs
```
