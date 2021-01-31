---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_api_gateway_client_certificate"
description: |-
  Provides an API Gateway Client Certificate.
---

# Resource: aws_api_gateway_client_certificate

Provides an API Gateway Client Certificate.

## Example Usage

```hcl
resource "aws_api_gateway_client_certificate" "demo" {
  description = "My client certificate"
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) The description of the client certificate.
* `tags` - (Optional) Key-value map of resource tags


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the client certificate.
* `created_date` - The date when the client certificate was created.
* `expiration_date` - The date when the client certificate will expire.
* `pem_encoded_certificate` - The PEM-encoded public key of the client certificate.
* `arn` - Amazon Resource Name (ARN)

## Import

API Gateway Client Certificates can be imported using the id, e.g.

```
$ terraform import aws_api_gateway_client_certificate.demo ab1cqe
```
