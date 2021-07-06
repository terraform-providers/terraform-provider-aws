---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_portfolio_constraints"
description: |-
  Provides information on Service Catalog Portfolio Constraints
---

# Data source: aws_servicecatalog_portfolio_constraints

Provides information on Service Catalog Portfolio Constraints.

## Example Usage

### Basic Usage

```terraform
data "aws_servicecatalog_portfolio_constraints" "example" {
  portfolio_id = "port-3lli3b3an"
}
```

## Argument Reference

The following arguments are required:

* `portfolio_id` - (Required) Portfolio identifier.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.
* `product_id` - (Optional) Product identifier.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `details` - List of information about the constraints. See details below.

### details

* `constraint_id` - Identifier of the constraint.
* `description` - Description of the constraint.
* `portfolio_id` - Identifier of the portfolio the product resides in. The constraint applies only to the instance of the product that lives within this portfolio.
* `product_id` - Identifier of the product the constraint applies to. A constraint applies to a specific instance of a product within a certain portfolio.
* `type` - Type of constraint. Valid values are `LAUNCH`, `NOTIFICATION`, `STACKSET`, and `TEMPLATE`.
