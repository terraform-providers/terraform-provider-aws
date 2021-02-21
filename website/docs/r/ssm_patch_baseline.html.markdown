---
subcategory: "SSM"
layout: "aws"
page_title: "AWS: aws_ssm_patch_baseline"
description: |-
  Provides an SSM Patch Baseline resource
---

# Resource: aws_ssm_patch_baseline

Provides an SSM Patch Baseline resource

~> **NOTE on Patch Baselines:** The `approved_patches` and `approval_rule` are
both marked as optional fields, but the Patch Baseline requires that at least one
of them is specified.

## Example Usage

Basic usage using `approved_patches` only

```hcl
resource "aws_ssm_patch_baseline" "production" {
  name             = "patch-baseline"
  approved_patches = ["KB123456"]
}
```

Advanced usage, specifying patch filters

```hcl
resource "aws_ssm_patch_baseline" "production" {
  name             = "patch-baseline"
  description      = "Patch Baseline Description"
  approved_patches = ["KB123456", "KB456789"]
  rejected_patches = ["KB987654"]

  global_filter {
    key    = "PRODUCT"
    values = ["WindowsServer2008"]
  }

  global_filter {
    key    = "CLASSIFICATION"
    values = ["ServicePacks"]
  }

  global_filter {
    key    = "MSRC_SEVERITY"
    values = ["Low"]
  }

  approval_rule {
    approve_after_days = 7
    compliance_level   = "HIGH"

    patch_filter {
      key    = "PRODUCT"
      values = ["WindowsServer2016"]
    }

    patch_filter {
      key    = "CLASSIFICATION"
      values = ["CriticalUpdates", "SecurityUpdates", "Updates"]
    }

    patch_filter {
      key    = "MSRC_SEVERITY"
      values = ["Critical", "Important", "Moderate"]
    }
  }

  approval_rule {
    approve_after_days = 7

    patch_filter {
      key    = "PRODUCT"
      values = ["WindowsServer2012"]
    }
  }
}
```

Advanced usage, specifying Microsoft application and Windows patch rules

```hcl
resource "aws_ssm_patch_baseline" "windows_os_apps" {
  name             = "WindowsOSAndMicrosoftApps"
  description      = "Patch both Windows and Microsoft apps"
  operating_system = "WINDOWS"

  approval_rule {
    approve_after_days = 7

    patch_filter {
      key    = "CLASSIFICATION"
      values = ["CriticalUpdates", "SecurityUpdates"]
    }

    patch_filter {
      key    = "MSRC_SEVERITY"
      values = ["Critical", "Important"]
    }
  }

  approval_rule {
    approve_after_days = 7

    patch_filter {
      key    = "PATCH_SET"
      values = ["APPLICATION"]
    }

    # Filter on Microsoft product if necessary
    patch_filter {
      key    = "PRODUCT"
      values = ["Office 2013", "Office 2016"]
    }
  }
}
```

Advanced usage, specifying alternate patch source repository

```hcl
resource "aws_ssm_patch_baseline" "al_2017_09" {
  name             = "Amazon-Linux-2017.09"
  description      = "My patch repository for Amazon Linux 2017.09"
  operating_system = "AMAZON_LINUX"

  approval_rule {
    # ...
  }

  source {
    name          = "My-AL2017.09"
    products      = ["AmazonLinux2017.09"]
    configuration = <<EOF
[amzn-main]
name=amzn-main-Base
mirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list
mirrorlist_expire=300
metadata_expire=300
priority=10
failovermethod=priority
fastestmirror_enabled=0
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga
enabled=1
retries=3
timeout=5
report_instanceid=yes
EOF
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the patch baseline.
* `description` - (Optional) The description of the patch baseline.
* `operating_system` - (Optional) Defines the operating system the patch baseline applies to. Supported operating systems include `WINDOWS`, `AMAZON_LINUX`, `AMAZON_LINUX_2`, `SUSE`, `UBUNTU`, `CENTOS`, and `REDHAT_ENTERPRISE_LINUX`. The Default value is `WINDOWS`.
* `approved_patches_compliance_level` - (Optional) Defines the compliance level for approved patches. This means that if an approved patch is reported as missing, this is the severity of the compliance violation. Valid compliance levels include the following: `CRITICAL`, `HIGH`, `MEDIUM`, `LOW`, `INFORMATIONAL`, `UNSPECIFIED`. The default value is `UNSPECIFIED`.
* `approved_patches` - (Optional) A list of explicitly approved patches for the baseline.
* `rejected_patches` - (Optional) A list of rejected patches.
* `global_filter` - (Optional) A set of global filters used to exclude patches from the baseline. Up to 4 global filters can be specified using Key/Value pairs. Valid Keys are `PRODUCT | CLASSIFICATION | MSRC_SEVERITY | PATCH_ID`.
* `approval_rule` - (Optional) A set of rules used to include patches in the baseline. up to 10 approval rules can be specified. Each approval_rule block requires the fields documented below.
* `source` - (Optional) Configuration block(s) with alternate sources for patches. Applies to Linux instances only. Documented below.
* `rejected_patches_action` - (Optional) The action for Patch Manager to take on patches included in the `rejected_patches` list. Allow values are `ALLOW_AS_DEPENDENCY` and `BLOCK`.
* `approved_patches_enable_non_security` - (Optional) Indicates whether the list of approved patches includes non-security updates that should be applied to the instances. Applies to Linux instances only.

The `approval_rule` block supports:

* `approve_after_days` - (Optional) The number of days after the release date of each patch matched by the rule the patch is marked as approved in the patch baseline. Valid Range: 0 to 100. Conflicts with `approve_until_date`
* `approve_until_date` - (Optional) The cutoff date for auto approval of released patches. Any patches released on or before this date are installed automatically. Date is formatted as `YYYY-MM-DD`. Conflicts with `approve_after_days`
* `patch_filter` - (Required) The patch filter group that defines the criteria for the rule. Up to 5 patch filters can be specified per approval rule using Key/Value pairs. Valid Keys are `PATCH_SET | PRODUCT | CLASSIFICATION | MSRC_SEVERITY | PATCH_ID`. Valid combinations of these Keys and the `operating_system` value can be found in the [SSM DescribePatchProperties API Reference](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_DescribePatchProperties.html). Valid Values are exact values for the patch property given as the key, or a wildcard `*`, which matches all values.
    * `PATCH_SET` defaults to `OS` if unspecified
* `compliance_level` - (Optional) Defines the compliance level for patches approved by this rule. Valid compliance levels include the following: `CRITICAL`, `HIGH`, `MEDIUM`, `LOW`, `INFORMATIONAL`, `UNSPECIFIED`. The default value is `UNSPECIFIED`.
* `enable_non_security` - (Optional) Boolean enabling the application of non-security updates. The default value is 'false'. Valid for Linux instances only.
* `tags` - (Optional) A map of tags to assign to the resource.

The `source` block supports:

* `name` - (Required) The name specified to identify the patch source.
* `configuration` - (Required) The value of the yum repo configuration. For information about other options available for your yum repository configuration, see the [`dnf.conf` documentation](https://man7.org/linux/man-pages/man5/dnf.conf.5.html)
* `products` - (Required) The specific operating system versions a patch repository applies to, such as `"Ubuntu16.04"`, `"AmazonLinux2016.09"`, `"RedhatEnterpriseLinux7.2"` or `"Suse12.7"`. For lists of supported product values, see [PatchFilter](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_PatchFilter.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the patch baseline.
* `arn` - The ARN of the patch baseline.

## Import

SSM Patch Baselines can be imported by their baseline ID, e.g.

```
$ terraform import aws_ssm_patch_baseline.example pb-12345678
```
