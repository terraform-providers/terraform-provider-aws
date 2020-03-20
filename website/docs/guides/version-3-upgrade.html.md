---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Version 3 Upgrade Guide"
description: |-
  Terraform AWS Provider Version 3 Upgrade Guide
---

# Terraform AWS Provider Version 3 Upgrade Guide

~> **NOTE:** This upgrade guide is a work in progress and will not be completed until the release of version 3.0.0 of the provider in the coming months. Many of the topics discussed, except for the actual provider upgrade, can be performed using the most recent 2.X version of the provider.

Version 3.0.0 of the AWS provider for Terraform is a major release and includes some changes that you will need to consider when upgrading. This guide is intended to help with that process and focuses only on changes from version 1.X to version 3.0.0.

Most of the changes outlined in this guide have been previously marked as deprecated in the Terraform plan/apply output throughout previous provider releases. These changes, such as deprecation notices, can always be found in the [Terraform AWS Provider CHANGELOG](https://github.com/terraform-providers/terraform-provider-aws/blob/master/CHANGELOG.md).

Upgrade topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Version Configuration](#provider-version-configuration)
- [Resource: aws_emr_cluster](#resource-aws_emr_cluster)

<!-- /TOC -->

## Provider Version Configuration

!> **WARNING:** This topic is placeholder documentation until version 3.0.0 is released in the coming months.

-> Before upgrading to version 3.0.0, it is recommended to upgrade to the most recent 2.X version of the provider and ensure that your environment successfully runs [`terraform plan`](https://www.terraform.io/docs/commands/plan.html) without unexpected changes or deprecation notices.

It is recommended to use [version constraints when configuring Terraform providers](https://www.terraform.io/docs/configuration/providers.html#provider-versions). If you are following that recommendation, update the version constraints in your Terraform configuration and run [`terraform init`](https://www.terraform.io/docs/commands/init.html) to download the new version.

For example, given this previous configuration:

```hcl
provider "aws" {
  # ... other configuration ...

  version = "~> 2.8"
}
```

Update to latest 3.X version:

```hcl
provider "aws" {
  # ... other configuration ...

  version = "~> 3.0"
}
```

## Resource: aws_emr_cluster

### core_instance_count Argument Removal

Switch your Terraform configuration to the `core_instance_group` configuration block instead.

For example, given this previous configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  core_instance_count = 2
}
```

An updated configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  core_instance_group {
    # ... other configuration ...

    instance_count = 2
  }
}
```

### core_instance_type Argument Removal

Switch your Terraform configuration to the `core_instance_group` configuration block instead.

For example, given this previous configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  core_instance_type = "m4.large"
}
```

An updated configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  core_instance_group {
    instance_type = "m4.large"
  }
}
```

### instance_group Configuration Block Removal

Switch your Terraform configuration to the `master_instance_group` and `core_instance_group` configuration blocks instead. For any task instance groups, use the `aws_emr_instance_group` resource.

For example, given this previous configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  instance_group {
    instance_role  = "MASTER"
    instance_type  = "m4.large"
  }

  instance_group {
    instance_count = 1
    instance_role  = "CORE"
    instance_type  = "c4.large"
  }

  instance_group {
    instance_count = 2
    instance_role  = "TASK"
    instance_type  = "c4.xlarge"
  }
}
```

An updated configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }
}

resource "aws_emr_instance_group" "example" {
  cluster_id     = "${aws_emr_cluster.example.id}"
  instance_count = 2
  instance_type  = "c4.xlarge"
}
```

### master_instance_type Argument Removal

Switch your Terraform configuration to the `master_instance_group` configuration block instead.

For example, given this previous configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  master_instance_type = "m4.large"
}
```

An updated configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  master_instance_group {
    instance_type = "m4.large"
  }
}
```

## Resource: aws_lb_listener_rule

### condition.field and condition.values Arguments Removal

Switch your Terraform configuration to use the `host_header` or `path_pattern` configuration block instead.

For example, given this previous configuration:

```hcl
resource "aws_lb_listener_rule" "example" {
  # ... other configuration ...

  condition {
    field  = "path-pattern"
    values = ["/static/*"]
  }
}
```

An updated configuration:

```hcl
resource "aws_lb_listener_rule" "example" {
  # ... other configuration ...

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}
```
