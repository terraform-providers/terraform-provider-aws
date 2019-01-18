---
layout: "aws"
page_title: "AWS: aws_canonical_user_id"
sidebar_current: "docs-aws-datasource-canonical-user-id"
description: |-
  Provides the canonical user ID for the AWS account associated with the provider
  connection to AWS.
---

# Data Source: aws_canonical_user_id

The Canonical User ID data source allows access to the [canonical user ID](http://docs.aws.amazon.com/general/latest/gr/acct-identifiers.html)
for the effective account in which Terraform is working.  

## Example Usage

```hcl
data "aws_canonical_user_id" "current" {}

output "canonical_user_id" {
  value = "${data.aws_canonical_user_id.current.id}"
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The canonical user ID associated with the AWS account.

* `display_name` - The human-friendly name linked to the canonical user ID. The bucket owner's display name. **NOTE:** [This value](https://docs.aws.amazon.com/AmazonS3/latest/API/RESTServiceGET.html) is only included in the response in the US East (N. Virginia), US West (N. California), US West (Oregon), Asia Pacific (Singapore), Asia Pacific (Sydney), Asia Pacific (Tokyo), EU (Ireland), and South America (São Paulo) regions.
