---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_component"
description: |-
    Provides details about an Image Builder Component
---

# Data Source: aws_imagebuilder_component

Provides details about an Image Builder Component.

## Example Usage

```hcl
data "aws_imagebuilder_component" "example" {
  arn = "arn:aws:imagebuilder:us-west-2:aws:component/amazon-cloudwatch-agent-linux/1.0.0"
}
```

## Argument Reference

* `arn` - (Required) Amazon Resource Name (ARN) of the component.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `change_description` - Change description of the component.
* `data` - Data of the component.
* `date_created` - Date the component was created.
* `description` - Description of the component.
* `encrypted` - Encryption status of the component.
* `kms_key_id` - Amazon Resource Name (ARN) of the Key Management Service (KMS) Key used to encrypt the component.
* `name` - Name of the component.
* `owner` - Owner of the component.
* `platform` - Platform of the component.
* `supported_os_versions` - Operating Systems (OSes) supported by the component.
* `tags` - Key-value map of resource tags for the component.
* `type` - Type of the component.
* `version` - Version of the component.
