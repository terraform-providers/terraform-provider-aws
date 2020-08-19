---
layout: "aws"
page_title: "AWS: aws_lightsail_domain_entry"
sidebar_current: "docs-aws-resource-lightsail-domain-entry"
description: |-
  Provides an Lightsail Domain Entry
---

# Resource: aws_lightsail_domain_entry

Creates a domain entry resource

## Example Usage, creating a new domain

```hcl
resource "aws_lightsail_domain" "domain_test" {
  domain_name = "mydomain.com"
}

resource "aws_lightsail_domain_entry" "entry_test" {
  domain_name = "${aws_lightsail_domain.domain_test.domain_name}"
  name = "a3.${aws_lightsail_domain.domain_test.domain_name}"
  type = "A"
  target = "127.0.0.1"
  depends_on = ["aws_lightsail_domain.domain_test"]
}

```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) The name of the Lightsail domain in which to create the entry
* `name` - (Required) Name of the entry record
* `type` - (Required) Type of record
* `target` - (Required) Target of the domain entry
* `is_alias` - (Optional) If the entry should be an alias Defaults to `false`
