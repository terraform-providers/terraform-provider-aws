---
subcategory: "IoT"
layout: "aws"
page_title: "AWS: aws_iot_provisioning_template"
description: |-
  Creates an IoT fleet provisioning template.
---

# Resource: aws_iot_provisioning_template

Creates an IoT fleet provisioning template. For more info, see the AWS documentation on [fleet provisioning](https://docs.aws.amazon.com/iot/latest/developerguide/provision-wo-cert.html).

## Example Usage

```hcl
data "aws_iam_policy_document" "iot_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["iot.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "iot_fleet_provisioning" {
  name               = "IoTProvisioningServiceRole"
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.iot_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "iot_fleet_provisioning_registration" {
  role       = aws_iam_role.iot_fleet_provisioning.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSIoTThingsRegistration"
}

data "aws_iam_policy_document" "device_policy" {
  statement {
    actions   = ["iot:Subscribe"]
    resources = ["*"]
  }
}

resource "aws_iot_policy" "device_policy" {
  name   = "DevicePolicy"
  policy = data.aws_iam_policy_document.device_policy.json
}

resource "aws_iot_provisioning_template" "fleet" {
	template_name         = "FleetTemplate"
	description           = "My provisioning template"
	provisioning_role_arn = aws_iam_role.iot_fleet_provisioning.arn

  template_body = jsonencode({
    Parameters = {
      "AWS::IoT::Certificate::Id" = { Type = "String" }
      SerialNumber                = { Type = "String" }
		}

    Resources = {
      certificate = {
        Properties = {
          CertificateId = { Ref = "AWS::IoT::Certificate::Id" }
          Status        = "Active"
        }

        Type = "AWS::IoT::Certificate"
      }

      policy = {
        Properties = {
          PolicyName = aws_iot_policy.device_policy.name
        }

        Type = "AWS::IoT::Policy"
      }
    }
  })
}
```

## Argument Reference

The following arguments are supported:

* `template_name` - (Required) The name of the fleet provisioning template.
* `description` - (Optional) The description of the fleet provisioning template.
* `enabled` - (Optional) True to enable the fleet provisioning template, otherwise false.
* `provisioningRoleArn` - (Required) The role ARN for the role associated with the fleet provisioning template. This IoT role grants permission to provision a device.
* `template_body` - (Required) The JSON formatted contents of the fleet provisioning template.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `default_version_id` - The default version of the fleet provisioning template.
* `template_arn` - The ARN that identifies the provisioning template.

## Import

IoT fleet provisioning templates can be imported using the `template_name`, e.g.

```
$ terraform import aws_iot_provisioning_template.fleet FleetProvisioningTemplate
```
