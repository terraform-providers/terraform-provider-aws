---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_product_portfolio_association"
description: |-
  Manages a Service Catalog Product Portfolio Association
---

# Resource: aws_servicecatalog_product_portfolio_association

Manages a Service Catalog Product Portfolio Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_product_portfolio_association" "example" {
  portfolio_id = "port-68656c6c6f"
  product_id   = "prod-dnigbtea24ste"
}
```

## Argument Reference

The following arguments are required:

* `portfolio_id` - (Required) Portfolio identifier.
* `product_id` - (Required) Product identifier.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.
* `source_portfolio_id` - (Optional) Identifier of the source portfolio.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier of the association.

## Import

`aws_servicecatalog_product_portfolio_association` can be imported using the accept language, portfolio ID, and product ID, e.g.

```
$ terraform import aws_servicecatalog_product_portfolio_association.example en:port-68656c6c6f:prod-dnigbtea24ste
```
