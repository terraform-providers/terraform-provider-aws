---
layout: "aws"
page_title: "AWS: aws_licensemanager_association"
sidebar_current: "docs-aws-resource-licensemanager-association"
description: |-
  Provides a License Manager association resource.
---

# aws_licensemanager_association

Provides a License Manager association.

~> **Note:** License configurations can also be associated with launch templates by specifying the `license_specifications` block for an `aws_launch_template`.

## Example Usage

```hcl
data "aws_ami" "example" {
  most_recent      = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-vpc-nat*"]
  }
}

resource "aws_instance" "example" {
  ami           = "${data.aws_ami.example.id}"
  instance_type = "t2.micro"
}

resource "aws_licensemanager_license_configuration" "example" {
  name                  = "Example"
  license_counting_type = "Instance"
}

resource "aws_licensemanager_association" "example" {
  license_configuration_arn = "${aws_licensemanager_license_configuration.example.arn}"
  resource_arn              = "${aws_instance.example.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `license_configuration_arn` - (Required) ARN of the license configuration.
* `resource_arn` - (Required) ARN of the resource associated with the license configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The license configuration ARN.

## Import

License configurations can be imported in the form `resource_arn,license_configuration_arn`, e.g.

```
$ terraform import aws_licensemanager_association.example arn:aws:ec2:eu-west-1:123456789012:image/ami-123456789abcdef01,arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef
```
