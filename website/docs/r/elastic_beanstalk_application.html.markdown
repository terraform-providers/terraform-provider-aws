---
subcategory: "Elastic Beanstalk"
layout: "aws"
page_title: "AWS: aws_elastic_beanstalk_application"
description: |-
  Provides an Elastic Beanstalk Application Resource
---

# Resource: aws_elastic_beanstalk_application

Provides an Elastic Beanstalk Application Resource. Elastic Beanstalk allows
you to deploy and manage applications in the AWS cloud without worrying about
the infrastructure that runs those applications.

This resource creates an application that has one configuration template named
`default`, and no application versions

## Example Usage

```hcl
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "tf-test-name"
  description = "tf-test-desc"

  appversion_lifecycle {
    service_role          = aws_iam_role.beanstalk_service.arn
    max_count             = 128
    delete_source_from_s3 = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the application, must be unique within your account
* `description` - (Optional) Short description of the application
* `tags` - (Optional) Key-value map of tags for the Elastic Beanstalk Application.

Application version lifecycle (`appversion_lifecycle`) supports the following settings.  Only one of either `max_count` or `max_age_in_days` can be provided:

* `service_role` - (Required) The ARN of an IAM service role under which the application version is deleted.  Elastic Beanstalk must have permission to assume this role.
* `max_count` - (Optional) The maximum number of application versions to retain ('max_age_in_days' and 'max_count' cannot be enabled simultaneously.).
* `max_age_in_days` - (Optional) The number of days to retain an application version ('max_age_in_days' and 'max_count' cannot be enabled simultaneously.).
* `delete_source_from_s3` - (Optional) Set to `true` to delete a version's source bundle from S3 when the application version is deleted.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN assigned by AWS for this Elastic Beanstalk Application.


## Import

Elastic Beanstalk Applications can be imported using the `name`, e.g.

```
$ terraform import aws_elastic_beanstalk_application.tf_test tf-test-name
```
