---
subcategory: "SES"
layout: "aws"
page_title: "AWS: aws_ses_configuration_set"
description: |-
  Provides an SES configuration set
---

# Resource: aws_ses_configuration_set

Provides an SES configuration set resource

## Example Usage

```hcl
resource "aws_ses_configuration_set" "test" {
  name = "some-configuration-set-test"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the configuration set

## Import

SES Configuration Sets can be imported using their `name`, e.g.

```
$ terraform import aws_ses_configuration_set.test some-configuration-set-test
```
