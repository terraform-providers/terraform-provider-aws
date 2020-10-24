---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_key_pair"
description: |-
    Provides details about a specific EC2 Key Pair
---

# Data Source: aws_key_pair

`aws_key_pair` provides details about a specific EC2 Key Pair.

This data source allows to find a EC2 Key Pair given name and certain search criteria.

## Example Usage

The following example shows how to get a EC2 Key Pair from its name.


```hcl
data "aws_key_pair" "example" {
  key_name = "test"
  filter {
    name   = "tag:Component"
    values = ["web"]
  }
}

output "fingerprint" {
  value = data.aws_key_pair.example.fingerprint
}

output "name" {
  value = data.aws_key_pair.example.key_name
}

output "id" {
  value = data.aws_key_pair.example.id
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
Key Pairs. The given filter must match exactly one Key Pairs.

* `key_name` - (Required) The Key Pair name.

* `filter` -  (Optional) Custom filter block as described below.

## Attributes Reference

All of the argument attributes except `filter` blocks are also exported as
result attributes. This data source will complete the data by populating
any fields that are not included in the configuration with the data for
the selected Key Pair.

The following attributes are additionally exported:

* `id` - Amazon Resource Name (ARN) of Key Pair
* `fingerprint` - The SHA-1 digest of the DER encoded private key..

