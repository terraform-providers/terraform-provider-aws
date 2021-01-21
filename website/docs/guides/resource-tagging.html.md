---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Resource Tagging"
description: |-
  Managing resource tags with the Terraform AWS Provider.
---

# Resource Tagging

Many AWS services implement [resource tags](https://docs.aws.amazon.com/general/latest/gr/aws_tagging.html) as an essential part of managing components. These arbitrary key-value pairs can be utilized for billing, ownership, automation, [access control](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_tags.html), and many other use cases. Given that these tags are an important aspect of successfully managing an AWS environment, the Terraform AWS Provider implements additional functionality beyond the typical one-to-one resource lifecycle management for easier and more customized implementations.

-> Not all AWS resources support tagging, which can differ across AWS services and even across resources within the same service. Browse the individual Terraform AWS Provider resource documentation pages for the `tags` argument, to see which support resource tagging. If the AWS API implements tagging support for a resource and it is missing from the Terraform AWS Provider resource, a [feature request](https://github.com/hashicorp/terraform-provider-aws/issues/new?labels=enhancement&template=Feature_Request.md) can be submitted.

<!-- TOC depthFrom:2 -->

- [Getting Started with Resource Tags](#getting-started-with-resource-tags)
- [Ignoring Changes to Specific Tags](#ignoring-changes-to-specific-tags)
    - [Ignoring Changes in Individual Resources](#ignoring-changes-in-individual-resources)
    - [Ignoring Changes in All Resources](#ignoring-changes-in-all-resources)
- [Managing Individual Resource Tags](#managing-individual-resource-tags)

<!-- /TOC -->

## Getting Started with Resource Tags

Terraform AWS Provider resources that support resource tags implement a consistent argument named `tags` which accepts a key-value map, e.g.

```hcl
resource "aws_vpc" "example" {
  # ... other configuration ...

  tags = {
    Name = "MyVPC"
  }
}
```

The tags for the resource are wholly managed by Terraform except tag keys beginning with `aws:` as these are managed by AWS services and cannot typically be edited or deleted. Any non-AWS tags added to the VPC outside of Terraform will be proposed for removal on the next Terraform execution. Missing tags or those with incorrect values from the Terraform configuration will be proposed for addition or update on the next Terraform execution. Advanced patterns that can adjust these behaviors for special use cases, such as Terraform AWS Provider configurations that affect all resources and the ability to manage resource tags for resources not managed by Terraform, can be found later in this guide.

For most environments and use cases, this is the typical implementation pattern, whether it be in a standalone Terraform configuration or within a [Terraform Module](https://www.terraform.io/docs/modules/). The Terraform configuration language also enables less repetitive configurations via [variables](https://www.terraform.io/docs/configuration/variables.html), [locals](https://www.terraform.io/docs/configuration/locals.html), or potentially a combination of these, e.g.

```hcl
# Terraform 0.12 and later syntax
variable "additional_tags" {
  default     = {}
  description = "Additional resource tags"
  type        = map(string)
}

resource "aws_vpc" "example" {
  # ... other configuration ...

  # This configuration combines some "default" tags with optionally provided additional tags
  tags = merge(
    var.additional_tags,
    {
      Name = "MyVPC"
    },
  )
}
```

## Ignoring Changes to Specific Tags

Systems outside of Terraform may automatically interact with the tagging associated with AWS resources. These external systems may be for administrative purposes, such as a Configuration Management Database, or the tagging may be required functionality for those systems, such as Kubernetes. This section shows methods to prevent Terraform from showing differences for specific tags.

### Ignoring Changes in Individual Resources

All Terraform resources support the [`lifecycle` configuration block `ignore_changes` argument](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes), which can be used to explicitly ignore all tags changes on a resource beyond an initial configuration or individual tag values.

In this example, the `Name` tag will be added to the VPC on resource creation, however any external changes to the `Name` tag value or the addition/removal of any tag (including the `Name` tag) will be ignored:

```hcl
# Terraform 0.12 and later syntax
resource "aws_vpc" "example" {
  # ... other configuration ...

  tags = {
    Name = "MyVPC"
  }

  lifecycle {
    ignore_changes = [tags]
  }
}
```

In this example, the `Name` and `Owner` tags will be added to the VPC on resource creation, however any external changes to the value of the `Name` tag will be ignored while any changes to other tags (including the `Owner` tag and any additions) will still be proposed:

```hcl
# Terraform 0.12 and later syntax
resource "aws_vpc" "example" {
  # ... other configuration ...

  tags = {
    Name  = "MyVPC"
    Owner = "Operations"
  }

  lifecycle {
    ignore_changes = [tags.Name]
  }
}
```

### Ignoring Changes in All Resources

As of version 2.60.0 of the Terraform AWS Provider, there is support for ignoring tag changes across all resources under a provider. This simplifies situations where certain tags may be externally applied more globally and enhances functionality beyond `ignore_changes` to support cases such as tag key prefixes.

In this example, all resources will ignore any addition of the `LastScanned` tag:

```hcl
provider "aws" {
  # ... potentially other configuration ...

  ignore_tags {
    keys = ["LastScanned"]
  }
}
```

In this example, all resources will ignore any addition of tags with the `kubernetes.io/` prefix, such as `kubernetes.io/cluster/name` or `kubernetes.io/role/elb`:

```hcl
provider "aws" {
  # ... potentially other configuration ...

  ignore_tags {
    key_prefixes = ["kubernetes.io/"]
  }
}
```

Any of the `ignore_tags` configurations can be combined as needed.

The provider ignore tags configuration applies to all Terraform AWS Provider resources under that particular instance (the `default` provider instance in the above cases). If multiple, different Terraform AWS Provider configurations are being used (e.g. [multiple provider instances](https://www.terraform.io/docs/configuration/providers.html#alias-multiple-provider-instances)), the ignore tags configuration must be added to all applicable provider configurations.

## Managing Individual Resource Tags

Certain Terraform AWS Provider services support a special resource for managing an individual tag on a resource without managing the resource itself. One example is the [`aws_ec2_tag` resource](/docs/providers/aws/r/ec2_tag.html). These resources enable tagging where resources are created outside Terraform such as EC2 Images (AMIs), shared across accounts via Resource Access Manager (RAM), or implicitly created by other means such as EC2 VPN Connections implicitly creating a taggable EC2 Transit Gateway VPN Attachment.

~> **NOTE:** This is an advanced use case and can cause conflicting management issues when improperly implemented. These individual tag resources should not be combined with the Terraform resource for managing the parent resource. For example, using `aws_vpc` and `aws_ec2_tag` to manage tags of the same VPC will cause a perpetual difference where the `aws_vpc` resource will try to remove the tag being added by the `aws_ec2_tag` resource.

-> Not all services supported by the Terraform AWS Provider implement these resources. Browse the Terraform AWS Provider resource documentation pages for a resource with a type ending in `_tag`. If there is a use case where this type of resource is missing, a [feature request](https://github.com/hashicorp/terraform-provider-aws/issues/new?labels=enhancement&template=Feature_Request.md) can be submitted.

```hcl
# Terraform 0.12 and later syntax
# ... other configuration ...

resource "aws_ec2_tag" "example" {
  resource_id = aws_vpn_connection.example.transit_gateway_attachment_id
  key         = "Owner"
  value       = "Operations"
}
```

To manage multiple tags for a resource in this scenario, [`for_each`](https://www.terraform.io/docs/configuration/meta-arguments/for_each.html) can be used:

```hcl
# Terraform 0.12 and later syntax
# ... other configuration ...

resource "aws_ec2_tag" "example" {
  for_each = { "Name" : "MyAttachment", "Owner" : "Operations" }

  resource_id = aws_vpn_connection.example.transit_gateway_attachment_id
  key         = each.key
  value       = each.value
}
```

The inline map provided to `for_each` in the example above is used for brevity, but other Terraform configuration language features similar to those noted at the beginning of this guide can be used to make the example more extensible.
