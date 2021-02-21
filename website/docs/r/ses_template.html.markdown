---
subcategory: "SES"
layout: "aws"
page_title: "AWS: aws_ses_template"
description: |-
  Provides a resource to create a SES template
---

# Resource: aws_ses_template

Provides a resource to create a SES template.

## Example Usage

```hcl
resource "aws_ses_template" "MyTemplate" {
  name    = "MyTemplate"
  subject = "Greetings, {{name}}!"
  html    = "<h1>Hello {{name}},</h1><p>Your favorite animal is {{favoriteanimal}}.</p>"
  text    = "Hello {{name}},\r\nYour favorite animal is {{favoriteanimal}}."
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the template. Cannot exceed 64 characters. You will refer to this name when you send email.
* `html` - (Optional) The HTML body of the email. Must be less than 500KB in size, including both the text and HTML parts.
* `subject` - (Optional) The subject line of the email.
* `text` - (Optional) The email body that will be visible to recipients whose email clients do not display HTML. Must be less than 500KB in size, including both the text and HTML parts.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the SES template
* `id` - The name of the SES template

## Import

SES templates can be imported using the template name, e.g.

```
$ terraform import aws_ses_template.MyTemplate MyTemplate
```
