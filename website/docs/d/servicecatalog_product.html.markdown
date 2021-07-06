---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_product"
description: |-
  Provides information on a Service Catalog Product
---

# Data source: aws_servicecatalog_product

Provides information on a Service Catalog Product.

-> **Tip:** A "provisioning artifact" is also referred to as a "version." A "distributor" is also referred to as a "vendor."

## Example Usage

### Basic Usage

```terraform
data "aws_servicecatalog_product" "example" {
  id = "prod-dnigbtea24ste"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Product ID.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the product.
* `created_time` - Time when the product was created.
* `description` - Description of the product.
* `distributor` - Distributor (i.e., vendor) of the product.
* `has_default_path` - Whether the product has a default path.
* `name` - Name of the product.
* `owner` - Owner of the product.
* `status` - Status of the product.
* `support_description` - Support information about the product.
* `support_email` - Contact email for product support.
* `support_url` - Contact URL for product support.
* `tags` - Tags to apply to the product.
* `type` - Type of product.
