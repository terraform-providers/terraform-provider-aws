---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_image"
description: |-
  Get information about Workspaces image.
---

# Data Source: aws_workspaces_image

Use this data source to get information about a Workspaces image.

## Example Usage

```hcl
data aws_workspaces_image example {
  image_id = "wsi-ten5h0y19"
}
```

## Argument Reference

The following arguments are supported:

* `image_id` – (Required) The ID of the image.

## Attributes Reference

The following attributes are exported:

* `name` – The name of the image.
* `description` – The description of the image.
* `os` – The operating system that the image is running.
* `required_tenancy` – Specifies whether the image is running on dedicated hardware. When Bring Your Own License (BYOL) is enabled, this value is set to DEDICATED. For more information, see [Bring Your Own Windows Desktop Images](https://docs.aws.amazon.com/workspaces/latest/adminguide/byol-windows-images.html).
* `state` – The status of the image.
