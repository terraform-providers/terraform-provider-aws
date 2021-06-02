---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_connection"
description: |-
  Manages an App Runner Connection.
---

# Resource: aws_apprunner_connection

Manages an App Runner Connection.

~> **NOTE:** After creation, you must complete the authentication handshake using the App Runner console.

## Example Usage

```terraform
resource "aws_apprunner_connection" "example" {
  connection_name = "example"
  provider_type   = "GITHUB"

  tags = {
    Name = "example-apprunner-connection"
  }
}
```

## Argument Reference

The following arguments supported:

* `connection_name` - (Required) Name of the connection.
* `provider_type` - (Required) The source repository provider. Valid values: `GITHUB`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the connection.
* `status` - The current state of the App Runner connection. When the state is `AVAILABLE`, you can use the connection to create an [`aws_apprunner_service` resource](apprunner_service.html).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

App Runner Connections can be imported by using the `connection_name`, e.g.

```
$ terraform import aws_apprunner_connection.example example
```
