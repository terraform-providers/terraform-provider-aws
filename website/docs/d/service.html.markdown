---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_service"
description: |-
  Compose and decompose AWS service DNS names
---

# Data Source: aws_service

Use this data source to compose and decompose AWS service DNS names.

## Example Usage

### Get Service DNS Name

```hcl
data "aws_region" "current" {}

data "aws_service" "test" {
  region     = data.aws_region.current.name
  service_id = "ec2"
}
```

### Use Service Reverse DNS Name to Get Components

```hcl
data "aws_service" "s3" {
  reverse_dns_name = "cn.com.amazonaws.cn-north-1.s3"
}
```

### Determine Regional Support for a Service

```hcl
data "aws_service" "s3" {
  reverse_dns_name = "com.amazonaws.us-gov-west-1.waf"
}
```

## Argument Reference

~> **Note:** If you provide `reverse_dns_name` or `dns_name` arguments, `region`, `service_id`, and `prefix` values are ignored. If you provide both `reverse_dns_name` and `dns_name`, `dns_name` takes precedence.

All arguments are optional:

* `dns_name` - DNS name of the service (e.g. `rds.us-east-1.amazonaws.com`).
* `partition` - Partition corresponding to the region.
* `region` - Region of the service (e.g. `us-west-2`, `ap-northeast-1`).
* `reverse_dns_name` - Reverse DNS name of the service (e.g. `com.amazonaws.us-west-2.s3`).
* `reverse_dns_prefix` - Prefix of the service (e.g. `com.amazonaws` in AWS Commercial, `cn.com.amazonaws` in AWS China).
* `service_id` - Service (e.g. `s3`, `rds`, `ec2`). Defaults to `ec2`.

## Attributes Reference

Besides the arguments above, the following attribute is exported.

* `supported` - Whether the service is supported in the region's partition. New services may not be listed immediately as supported.
