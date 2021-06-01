---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_component"
description: |-
    Manage an Image Builder Component
---

# Resource: aws_imagebuilder_component

Manages an Image Builder Component.

## Example Usage

### Inline Data Document

```hcl
resource "aws_imagebuilder_component" "example" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = "example"
  platform = "Linux"
  version  = "1.0.0"
}
```

### URI Document

```hcl
resource "aws_imagebuilder_component" "example" {
  name     = "example"
  platform = "Linux"
  uri      = "s3://${aws_s3_bucket_object.example.bucket}/${aws_s3_bucket_object.example.key}"
  version  = "1.0.0"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the component.
* `platform` - (Required) Platform of the component.
* `version` - (Required) Version of the component.

The following attributes are optional:

* `change_description` - (Optional) Change description of the component.
* `data` - (Optional) Inline YAML string with data of the component. Exactly one of `data` and `uri` can be specified. Terraform will only perform drift detection of its value when present in a configuration.
* `description` - (Optional) Description of the component.
* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of the Key Management Service (KMS) Key used to encrypt the component.
* `supported_os_versions` - (Optional) Set of Operating Systems (OS) supported by the component.
* `tags` - (Optional) Key-value map of resource tags for the component.
* `uri` - (Optional) S3 URI with data of the component. Exactly one of `data` and `uri` can be specified.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - (Required) Amazon Resource Name (ARN) of the component.
* `date_created` - Date the component was created.
* `encrypted` - Encryption status of the component.
* `owner` - Owner of the component.
* `type` - Type of the component.

## Import

`aws_imagebuilder_components` resources can be imported by using the Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_imagebuilder_component.example arn:aws:imagebuilder:us-east-1:123456789012:component/example/1.0.0/1
```

Certain resource arguments, such as `uri`, cannot be read via the API and imported into Terraform. Terraform will display a difference for these arguments the first run after import if declared in the Terraform configuration for an imported resource.
