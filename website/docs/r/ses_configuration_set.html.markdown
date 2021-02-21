---
subcategory: "SES"
layout: "aws"
page_title: "AWS: aws_ses_configuration_set"
description: |-
  Provides an SES configuration set
---

# Resource: aws_ses_configuration_set

Provides an SES configuration set resource.

## Example Usage

```hcl
resource "aws_ses_configuration_set" "test" {
  name = "some-configuration-set-test"
}
```

### Require TLS Connections

```hcl
resource "aws_ses_configuration_set" "test" {
  name = "some-configuration-set-test"

  delivery_options {
    tls_policy = "Require"
  }
}
```

## Argument Reference

The following argument is required:

* `name` - (Required) Name of the configuration set.

The following argument is optional:

* `delivery_options` - (Optional) Configuration block. Detailed below.

### delivery_options

* `tls_policy` - (Optional) Specifies whether messages that use the configuration set are required to use Transport Layer Security (TLS). If the value is `Require`, messages are only delivered if a TLS connection can be established. If the value is `Optional`, messages can be delivered in plain text if a TLS connection can't be established. Valid values: `Require` or `Optional`. Defaults to `Optional`.

## Attributes Reference

In addition to the arguments, which are exported, the following attributes are exported:

* `arn` - SES configuration set ARN.
* `id` - SES configuration set name.

## Import

SES Configuration Sets can be imported using their `name`, e.g.

```
$ terraform import aws_ses_configuration_set.test some-configuration-set-test
```
