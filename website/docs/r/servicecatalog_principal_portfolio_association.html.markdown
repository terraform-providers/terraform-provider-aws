---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_principal_portfolio_association"
description: |-
  Manages a Service Catalog Principal Portfolio Association
---

# Resource: aws_servicecatalog_principal_portfolio_association

Manages a Service Catalog Principal Portfolio Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_principal_portfolio_association" "example" {
  portfolio_id  = "port-68656c6c6f"
  principal_arn = "arn:aws:iam::123456789012:user/Eleanor"
}
```

## Argument Reference

The following arguments are required:

* `portfolio_id` - (Required) Portfolio identifier.
* `principal_arn` - (Required) Principal ARN.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.
* `principal_type` - (Optional) Principal type. Setting this argument empty (e.g., `principal_type = ""`) will result in an error. Valid value is `IAM`. Default is `IAM`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier of the association.

## Import

`aws_servicecatalog_principal_portfolio_association` can be imported using the accept language, principal ARN, and portfolio ID, separated by a comma, e.g.

```
$ terraform import aws_servicecatalog_principal_portfolio_association.example en,arn:aws:iam::123456789012:user/Eleanor,port-68656c6c6f
```
