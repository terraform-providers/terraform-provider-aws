---
layout: "aws"
page_title: "AWS: aws_ssm_document"
sidebar_current: "docs-aws-datasource-ssm-document"
description: |-
  Provides a SSM Document datasource
---

# Data Source: aws_ssm_document

Gets the contents of the specified Systems Manager document.

## Example Usage

To get the contents of the document owned by AWS.

```hcl
data "aws_ssm_document" "foo" {
  name = "AWS-GatherSoftwareInventory"
  document_format = "YAML"
}

output "content" {
  value = "${data.aws_ssm_document.foo.content}"
}
```
To get the contents of the custom document.

```hcl
data "aws_ssm_document" "test" {
  name = "${aws_ssm_document.test.name}"
  document_format = "JSON"
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Systems Manager document.
* `document_format` - (Optional) Returns the document in the specified format. The document format can be either JSON or YAML. JSON is the default format.
* `document_version` - (Optional) The document version for which you want information.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the document.
* `content` - The contents of the document.
* `document_type` - The type of the document.
