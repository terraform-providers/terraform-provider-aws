---
subcategory: "S3"
layout: "aws"
page_title: "AWS: awslogsdelivery_canonical_user_id"
description: |-
  Provides the canonical user ID for the AWS `awslogsdelivery` account.
---

# Data Source: awslogsdelivery_canonical_user_id

The Awslogsdelivery Canonical User ID data source allows access to the [canonical user ID](http://docs.aws.amazon.com/general/latest/gr/acct-identifiers.html)
for the AWS `awslogsdelivery` account.  

## Example Usage

```hcl
data "aws_awslogsdelivery_canonical_user_id" "example" {}

resource "aws_s3_bucket" "example" {
  bucket = "example"

  grant {
    id          = data.aws_awslogsdelivery_canonical_user_id.example.id
    type        = "CanonicalUser"
    permissions = ["FULL_CONTROL"]
  }
}
```

## Argument Reference

There are no arguments available for this data source.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The canonical user ID for the AWS `awslogsdelivery` account.

