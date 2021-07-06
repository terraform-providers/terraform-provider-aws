---
subcategory: "Amplify Console"
layout: "aws"
page_title: "AWS: aws_amplify_domain_association"
description: |-
  Provides an Amplify Domain Association resource.
---

# Resource: aws_amplify_domain_association

Provides an Amplify Domain Association resource.

## Example Usage

```terraform
resource "aws_amplify_app" "example" {
  name = "app"

  # Setup redirect from https://example.com to https://www.example.com
  custom_rule {
    source = "https://example.com"
    status = "302"
    target = "https://www.example.com"
  }
}

resource "aws_amplify_branch" "master" {
  app_id      = aws_amplify_app.example.id
  branch_name = "master"
}

resource "aws_amplify_domain_association" "example" {
  app_id      = aws_amplify_app.example.id
  domain_name = "example.com"

  # https://example.com
  sub_domain {
    branch_name = aws_amplify_branch.master.branch_name
    prefix      = ""
  }

  # https://www.example.com
  sub_domain {
    branch_name = aws_amplify_branch.master.branch_name
    prefix      = "www"
  }
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) The unique ID for an Amplify app.
* `domain_name` - (Required) The domain name for the domain association.
* `sub_domain` - (Required) The setting for the subdomain. Documented below.
* `wait_for_verification` - (Optional) If enabled, the resource will wait for the domain association status to change to `PENDING_DEPLOYMENT` or `AVAILABLE`. Setting this to `false` will skip the process. Default: `true`.

The `sub_domain` configuration block supports the following arguments:

* `branch_name` - (Required) The branch name setting for the subdomain.
* `prefix` - (Required) The prefix setting for the subdomain.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) for the domain association.
* `certificate_verification_dns_record` - The DNS record for certificate verification.

The `sub_domain` configuration block exports the following attributes:

* `dns_record` - The DNS record for the subdomain.
* `verified` - The verified status of the subdomain.

## Import

Amplify domain association can be imported using `app_id` and `domain_name`, e.g.

```
$ terraform import aws_amplify_domain_association.app d2ypk4k47z8u6/example.com
```
