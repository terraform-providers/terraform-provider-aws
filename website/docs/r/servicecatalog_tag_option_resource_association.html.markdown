---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_tag_option_resource_association"
description: |-
  Manages a Service Catalog Tag Option Resource Association
---

# Resource: aws_servicecatalog_tag_option_resource_association

Manages a Service Catalog Tag Option Resource Association.

-> **Tip:** A "resource" is either a Service Catalog portfolio or product.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_tag_option_resource_association" "example" {
  resource_id   = "prod-dnigbtea24ste"
  tag_option_id = "tag-pjtvyakdlyo3m"
}
```

## Argument Reference

The following arguments are required:

* `resource_id` - (Required) Resource identifier.
* `tag_option_id` - (Required) Tag Option identifier.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier of the association.
* `resource_arn` - ARN of the resource.
* `resource_created_time` - Creation time of the resource.
* `resource_description` - Description of the resource.
* `resource_name` - Description of the resource.

## Import

`aws_servicecatalog_tag_option_resource_association` can be imported using the tag option ID and resource ID, e.g.

```
$ terraform import aws_servicecatalog_tag_option_resource_association.example tag-pjtvyakdlyo3m:prod-dnigbtea24ste
```
