---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Version 3 Upgrade Guide"
description: |-
  Terraform AWS Provider Version 3 Upgrade Guide
---

# Terraform AWS Provider Version 3 Upgrade Guide

Version 3.0.0 of the AWS provider for Terraform is a major release and includes some changes that you will need to consider when upgrading. This guide is intended to help with that process and focuses only on changes from version 2.X to version 3.0.0. See the [Version 2 Upgrade Guide](/docs/providers/aws/guides/version-2-upgrade.html) for information about upgrading from 1.X to version 2.0.0.

Most of the changes outlined in this guide have been previously marked as deprecated in the Terraform plan/apply output throughout previous provider releases. These changes, such as deprecation notices, can always be found in the [Terraform AWS Provider CHANGELOG](https://github.com/hashicorp/terraform-provider-aws/blob/master/CHANGELOG.md).

~> **NOTE:** Version 3.0.0 and later of the AWS Provider can only be automatically installed on Terraform 0.12 and later.

Upgrade topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Version Configuration](#provider-version-configuration)
- [Provider Authentication Updates](#provider-authentication-updates)
- [Provider Custom Service Endpoint Updates](#provider-custom-service-endpoint-updates)
- [Data Source: aws_availability_zones](#data-source-aws_availability_zones)
- [Data Source: aws_lambda_invocation](#data-source-aws_lambda_invocation)
- [Data Source: aws_launch_template](#data-source-aws_launch_template)
- [Data Source: aws_route53_resolver_rule](#data-source-aws_route53_resolver_rule)
- [Data Source: aws_route53_zone](#data-source-aws_route53_zone)
- [Resource: aws_acm_certificate](#resource-aws_acm_certificate)
- [Resource: aws_api_gateway_method_settings](#resource-aws_api_gateway_method_settings)
- [Resource: aws_autoscaling_group](#resource-aws_autoscaling_group)
- [Resource: aws_cloudfront_distribution](#resource-aws_cloudfront_distribution)
- [Resource: aws_cloudwatch_log_group](#resource-aws_cloudwatch_log_group)
- [Resource: aws_codepipeline](#resource-aws_codepipeline)
- [Resource: aws_cognito_user_pool](#resource-aws_cognito_user_pool)
- [Resource: aws_dx_gateway](#resource-aws_dx_gateway)
- [Resource: aws_dx_gateway_association](#resource-aws_dx_gateway_association)
- [Resource: aws_dx_gateway_association_proposal](#resource-aws_dx_gateway_association_proposal)
- [Resource: aws_ebs_volume](#resource-aws_ebs_volume)
- [Resource: aws_elastic_transcoder_preset](#resource-aws_elastic_transcoder_preset)
- [Resource: aws_emr_cluster](#resource-aws_emr_cluster)
- [Resource: aws_glue_job](#resource-aws_glue_job)
- [Resource: aws_iam_access_key](#resource-aws_iam_access_key)
- [Resource: aws_iam_instance_profile](#resource-aws_iam_instance_profile)
- [Resource: aws_iam_server_certificate](#resource-aws_iam_server_certificate)
- [Resource: aws_instance](#resource-aws_instance)
- [Resource: aws_lambda_alias](#resource-aws_lambda_alias)
- [Resource: aws_launch_template](#resource-aws_launch_template)
- [Resource: aws_lb_listener_rule](#resource-aws_lb_listener_rule)
- [Resource: aws_msk_cluster](#resource-aws_msk_cluster)
- [Resource: aws_rds_cluster](#resource-aws_rds_cluster)
- [Resource: aws_route53_resolver_rule](#resource-aws_route53_resolver_rule)
- [Resource: aws_route53_zone](#resource-aws_route53_zone)
- [Resource: aws_s3_bucket](#resource-aws_s3_bucket)
- [Resource: aws_s3_bucket_metric](#resource-aws_s3_bucket_metric)
- [Resource: aws_security_group](#resource-aws_security_group)
- [Resource: aws_sns_platform_application](#resource-aws_sns_platform_application)
- [Resource: aws_spot_fleet_request](#resource-aws_spot_fleet_request)

<!-- /TOC -->

## Provider Version Configuration

-> Before upgrading to version 3.0.0, it is recommended to upgrade to the most recent 2.X version of the provider and ensure that your environment successfully runs [`terraform plan`](https://www.terraform.io/docs/commands/plan.html) without unexpected changes or deprecation notices.

It is recommended to use [version constraints when configuring Terraform providers](https://www.terraform.io/docs/configuration/providers.html#provider-versions). If you are following that recommendation, update the version constraints in your Terraform configuration and run [`terraform init`](https://www.terraform.io/docs/commands/init.html) to download the new version.

For example, given this previous configuration:

```hcl
provider "aws" {
  # ... other configuration ...

  version = "~> 2.70"
}
```

Update to latest 3.X version:

```hcl
provider "aws" {
  # ... other configuration ...

  version = "~> 3.0"
}
```

## Provider Authentication Updates

### Authentication Ordering

Previously, the provider preferred credentials in the following order:

- Static credentials (those defined in the Terraform configuration)
- Environment variables (e.g. `AWS_ACCESS_KEY_ID` or `AWS_PROFILE`)
- Shared credentials file (e.g. `~/.aws/credentials`)
- EC2 Instance Metadata Service
- Default AWS Go SDK handling (shared configuration, CodeBuild/ECS/EKS)

The provider now prefers the following credential ordering:

- Static credentials (those defined in the Terraform configuration)
- Environment variables (e.g. `AWS_ACCESS_KEY_ID` or `AWS_PROFILE`)
- Shared credentials and/or configuration file (e.g. `~/.aws/credentials` and `~/.aws/config`)
- Default AWS Go SDK handling (shared configuration, CodeBuild/ECS/EKS, EC2 Instance Metadata Service)

This means workarounds of disabling the EC2 Instance Metadata Service handling to enable CodeBuild/ECS/EKS credentials or to enable other credential methods such as `credential_process` in the AWS shared configuration are no longer necessary.

### Shared Configuration File Automatically Enabled

The `AWS_SDK_LOAD_CONFIG` environment variable is no longer necessary for the provider to automatically load the AWS shared configuration file (e.g. `~/.aws/config`).

### Removal of AWS_METADATA_TIMEOUT Environment Variable Usage

The provider now relies on the default AWS Go SDK timeouts for interacting with the EC2 Instance Metadata Service.

## Provider Custom Service Endpoint Updates

### Removal of kinesis_analytics and r53 Arguments

The [custom service endpoints](custom-service-endpoints.html) for Kinesis Analytics and Route 53 now use the `kinesisanalytics` and `route53` argument names in the provider configuration.

For example, given this previous configuration:

```hcl
provider "aws" {
  # ... potentially other configuration ...

  endpoints {
    # ... potentially other configuration ...

    kinesis_analytics = "https://example.com"
    r53               = "https://example.com"
  }
}
```

An updated configuration:

```hcl
provider "aws" {
  # ... potentially other configuration ...

  endpoints {
    # ... potentially other configuration ...

    kinesisanalytics = "https://example.com"
    route53          = "https://example.com"
  }
}
```

## Data Source: aws_availability_zones

### blacklisted_names Attribute Removal

Switch your Terraform configuration to the `exclude_names` attribute instead.

For example, given this previous configuration:

```hcl
data "aws_availability_zones" "example" {
  blacklisted_names = ["us-west-2d"]
}
```

An updated configuration:

```hcl
data "aws_availability_zones" "example" {
  exclude_names = ["us-west-2d"]
}
```

### blacklisted_zone_ids Attribute Removal

Switch your Terraform configuration to the `exclude_zone_ids` attribute instead.

For example, given this previous configuration:

```hcl
data "aws_availability_zones" "example" {
  blacklisted_zone_ids = ["usw2-az4"]
}
```

An updated configuration:

```hcl
data "aws_availability_zones" "example" {
  exclude_zone_ids = ["usw2-az4"]
}
```

## Data Source: aws_lambda_invocation

### result_map Attribute Removal

Switch your Terraform configuration to the `result` attribute with the [`jsondecode()` function](https://www.terraform.io/docs/configuration/functions/jsondecode.html) instead.

For example, given this previous configuration:

```hcl
# In Terraform 0.11 and earlier, the result_map attribute can be used
# to convert a result JSON string to a map of string keys to string values.
output "lambda_result" {
  value = "${data.aws_lambda_invocation.example.result_map["key1"]}"
}
```

An updated configuration:

```hcl
# In Terraform 0.12 and later, the jsondecode() function can be used
# to convert a result JSON string to native Terraform types.
output "lambda_result" {
  value = jsondecode(data.aws_lambda_invocation.example.result)["key1"]
}
```

## Data Source: aws_launch_template

### Error raised if no matching launch template is found

Previously, when a launch template matching the criteria was not found the data source would have been `null`.
Now this could produce errors similar to the below:

```
data.aws_launch_template.current: Refreshing state...

Error: error reading launch template: empty output
```

Configuration that depend on the previous behavior will need to be updated.

## Data Source: aws_route53_resolver_rule

### Removal of trailing period in domain_name argument

Previously the data-source returned the Resolver Rule Domain Name directly from the API, which included a `.` suffix. This proves difficult when many other AWS services do not accept this trailing period (e.g. ACM Certificate). This period is now automatically removed. For example, when the attribute would previously return a Resolver Rule Domain Name such as `example.com.`, the attribute now will be returned as `example.com`.
While the returned value will omit the trailing period, use of configurations with trailing periods will not be interrupted.

## Data Source: aws_route53_zone

### Removal of trailing period in name argument

Previously the data-source returned the Hosted Zone Domain Name directly from the API, which included a `.` suffix. This proves difficult when many other AWS services do not accept this trailing period (e.g. ACM Certificate). This period is now automatically removed. For example, when the attribute would previously return a Hosted Zone Domain Name such as `example.com.`, the attribute now will be returned as `example.com`.
While the returned value will omit the trailing period, use of configurations with trailing periods will not be interrupted.

## Resource: aws_acm_certificate

### domain_validation_options Changed from List to Set

Previously, the `domain_validation_options` attribute was a list type and completely unknown until after an initial `terraform apply`. This generally required complicated configuration workarounds to properly create DNS validation records since referencing this attribute directly could produce errors similar to the below:

```
Error: Invalid for_each argument

  on main.tf line 16, in resource "aws_route53_record" "existing":
  16:   for_each = aws_acm_certificate.existing.domain_validation_options

The "for_each" value depends on resource attributes that cannot be determined
until apply, so Terraform cannot predict how many instances will be created.
To work around this, use the -target argument to first apply only the
resources that the for_each depends on.
```

The `domain_validation_options` attribute is now a set type and the resource will attempt to populate the information necessary during the planning phase to handle the above situation in most environments without workarounds. This change also prevents Terraform from showing unexpected differences if the API returns the results in varying order.

Configuration references to this attribute will likely require updates since sets cannot be indexed (e.g. `domain_validation_options[0]` or the older `domain_validation_options.0.` syntax will return errors). If the `domain_validation_options` list previously contained only a single element like the two examples just shown, it may be possible to wrap these references using the [`tolist()` function](https://www.terraform.io/docs/configuration/functions/tolist.html) (e.g. `tolist(aws_acm_certificate.example.domain_validation_options)[0]`) as a quick configuration update, however given the complexity and workarounds required with the previous `domain_validation_options` attribute implementation, different environments will require different configuration updates and migration steps. Below is a more advanced example. Further questions on potential update steps can be submitted to the [community forums](https://discuss.hashicorp.com/c/terraform-providers/tf-aws/33).

For example, given this previous configuration using a `count` based resource approach that may have been used in certain environments:

```hcl
data "aws_route53_zone" "public_root_domain" {
  name = var.public_root_domain
}

resource "aws_acm_certificate" "existing" {
  domain_name = "existing.${var.public_root_domain}"
  subject_alternative_names = [
    "existing1.${var.public_root_domain}",
    "existing2.${var.public_root_domain}",
    "existing3.${var.public_root_domain}",
  ]
  validation_method = "DNS"
}

resource "aws_route53_record" "existing" {
  count = length(aws_acm_certificate.existing.subject_alternative_names) + 1

  allow_overwrite = true
  name            = aws_acm_certificate.existing.domain_validation_options[count.index].resource_record_name
  records         = [aws_acm_certificate.existing.domain_validation_options[count.index].resource_record_value]
  ttl             = 60
  type            = aws_acm_certificate.existing.domain_validation_options[count.index].resource_record_type
  zone_id         = data.aws_route53_zone.public_root_domain.zone_id
}

resource "aws_acm_certificate_validation" "existing" {
  certificate_arn         = aws_acm_certificate.existing.arn
  validation_record_fqdns = aws_route53_record.existing[*].fqdn
}

```

It will receive errors like the below after upgrading:

```
Error: Invalid index

  on main.tf line 14, in resource "aws_route53_record" "existing":
  14:   name    = aws_acm_certificate.existing.domain_validation_options[count.index].resource_record_name
    |----------------
    | aws_acm_certificate.existing.domain_validation_options is set of object with 4 elements
    | count.index is 1

This value does not have any indices.
```

Since the `domain_validation_options` attribute changed from a list to a set and sets cannot be indexed in Terraform, the recommendation is to update the configuration to use the more stable [resource `for_each` support](https://www.terraform.io/docs/configuration/meta-arguments/for_each.html) instead of [`count`](https://www.terraform.io/docs/configuration/meta-arguments/count.html). Note the slight change in the `validation_record_fqdns` syntax as well.

```hcl
resource "aws_route53_record" "existing" {
  for_each = {
    for dvo in aws_acm_certificate.existing.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.public_root_domain.zone_id
}

resource "aws_acm_certificate_validation" "existing" {
  certificate_arn         = aws_acm_certificate.existing.arn
  validation_record_fqdns = [for record in aws_route53_record.existing : record.fqdn]
}
```

After the configuration has been updated, a plan should no longer error and may look like the following:

```
------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create
  - destroy
-/+ destroy and then create replacement

Terraform will perform the following actions:

  # aws_acm_certificate_validation.existing must be replaced
-/+ resource "aws_acm_certificate_validation" "existing" {
        certificate_arn         = "arn:aws:acm:us-east-2:123456789012:certificate/ccbc58e8-061d-4443-9035-d3af0512e863"
      ~ id                      = "2020-07-16 00:01:19 +0000 UTC" -> (known after apply)
      ~ validation_record_fqdns = [
          - "_40b71647a8d88eb82d53fe988e8a3cc1.existing2.example.com",
          - "_812ddf11b781af1eec1643ec58f102d2.existing.example.com",
          - "_8dc56b6e35f699b8754afcdd79e9748d.existing3.example.com",
          - "_d7112da809a40e848207c04399babcec.existing1.example.com",
        ] -> (known after apply) # forces replacement
    }

  # aws_route53_record.existing will be destroyed
  - resource "aws_route53_record" "existing" {
      - fqdn    = "_812ddf11b781af1eec1643ec58f102d2.existing.example.com" -> null
      - id      = "Z123456789012__812ddf11b781af1eec1643ec58f102d2.existing.example.com._CNAME" -> null
      - name    = "_812ddf11b781af1eec1643ec58f102d2.existing.example.com" -> null
      - records = [
          - "_bdeba72164eec216c55a32374bcceafd.jfrzftwwjs.acm-validations.aws.",
        ] -> null
      - ttl     = 60 -> null
      - type    = "CNAME" -> null
      - zone_id = "Z123456789012" -> null
    }

  # aws_route53_record.existing[1] will be destroyed
  - resource "aws_route53_record" "existing" {
      - fqdn    = "_40b71647a8d88eb82d53fe988e8a3cc1.existing2.example.com" -> null
      - id      = "Z123456789012__40b71647a8d88eb82d53fe988e8a3cc1.existing2.example.com._CNAME" -> null
      - name    = "_40b71647a8d88eb82d53fe988e8a3cc1.existing2.example.com" -> null
      - records = [
          - "_638532db1fa6a1b71aaf063c8ea29d52.jfrzftwwjs.acm-validations.aws.",
        ] -> null
      - ttl     = 60 -> null
      - type    = "CNAME" -> null
      - zone_id = "Z123456789012" -> null
    }

  # aws_route53_record.existing[2] will be destroyed
  - resource "aws_route53_record" "existing" {
      - fqdn    = "_d7112da809a40e848207c04399babcec.existing1.example.com" -> null
      - id      = "Z123456789012__d7112da809a40e848207c04399babcec.existing1.example.com._CNAME" -> null
      - name    = "_d7112da809a40e848207c04399babcec.existing1.example.com" -> null
      - records = [
          - "_6e1da5574ab46a6c782ed73438274181.jfrzftwwjs.acm-validations.aws.",
        ] -> null
      - ttl     = 60 -> null
      - type    = "CNAME" -> null
      - zone_id = "Z123456789012" -> null
    }

  # aws_route53_record.existing[3] will be destroyed
  - resource "aws_route53_record" "existing" {
      - fqdn    = "_8dc56b6e35f699b8754afcdd79e9748d.existing3.example.com" -> null
      - id      = "Z123456789012__8dc56b6e35f699b8754afcdd79e9748d.existing3.example.com._CNAME" -> null
      - name    = "_8dc56b6e35f699b8754afcdd79e9748d.existing3.example.com" -> null
      - records = [
          - "_a419f8410d2e0720528a96c3506f3841.jfrzftwwjs.acm-validations.aws.",
        ] -> null
      - ttl     = 60 -> null
      - type    = "CNAME" -> null
      - zone_id = "Z123456789012" -> null
    }

  # aws_route53_record.existing["existing.example.com"] will be created
  + resource "aws_route53_record" "existing" {
      + allow_overwrite = true
      + fqdn            = (known after apply)
      + id              = (known after apply)
      + name            = "_812ddf11b781af1eec1643ec58f102d2.existing.example.com"
      + records         = [
          + "_bdeba72164eec216c55a32374bcceafd.jfrzftwwjs.acm-validations.aws.",
        ]
      + ttl             = 60
      + type            = "CNAME"
      + zone_id         = "Z123456789012"
    }

  # aws_route53_record.existing["existing1.example.com"] will be created
  + resource "aws_route53_record" "existing" {
      + allow_overwrite = true
      + fqdn            = (known after apply)
      + id              = (known after apply)
      + name            = "_d7112da809a40e848207c04399babcec.existing1.example.com"
      + records         = [
          + "_6e1da5574ab46a6c782ed73438274181.jfrzftwwjs.acm-validations.aws.",
        ]
      + ttl             = 60
      + type            = "CNAME"
      + zone_id         = "Z123456789012"
    }

  # aws_route53_record.existing["existing2.example.com"] will be created
  + resource "aws_route53_record" "existing" {
      + allow_overwrite = true
      + fqdn            = (known after apply)
      + id              = (known after apply)
      + name            = "_40b71647a8d88eb82d53fe988e8a3cc1.existing2.example.com"
      + records         = [
          + "_638532db1fa6a1b71aaf063c8ea29d52.jfrzftwwjs.acm-validations.aws.",
        ]
      + ttl             = 60
      + type            = "CNAME"
      + zone_id         = "Z123456789012"
    }

  # aws_route53_record.existing["existing3.example.com"] will be created
  + resource "aws_route53_record" "existing" {
      + allow_overwrite = true
      + fqdn            = (known after apply)
      + id              = (known after apply)
      + name            = "_8dc56b6e35f699b8754afcdd79e9748d.existing3.example.com"
      + records         = [
          + "_a419f8410d2e0720528a96c3506f3841.jfrzftwwjs.acm-validations.aws.",
        ]
      + ttl             = 60
      + type            = "CNAME"
      + zone_id         = "Z123456789012"
    }

Plan: 5 to add, 0 to change, 5 to destroy.
```

Due to the type of configuration change, Terraform does not know that the previous `aws_route53_record` resources (indexed by number in the existing state) and the new resources (indexed by domain names in the updated configuration) are equivalent. Typically in this situation, the [`terraform state mv` command](https://www.terraform.io/docs/commands/state/mv.html) can be used to reduce the plan to show no changes. This is done by associating the count index (e.g. `[1]`) with the equivalent domain name index (e.g. `["existing2.example.com"]`), making one of the four commands to fix the above example: `terraform state mv 'aws_route53_record.existing[1]' 'aws_route53_record.existing["existing2.example.com"]'`. It is recommended to use this `terraform state mv` update process where possible to reduce chances of unexpected behaviors or changes in an environment.

If using `terraform state mv` to reduce the plan to show no changes, no additional steps are required.

In larger or more complex environments though, this process can be tedius to match the old resource address to the new resource address and run all the necessary `terraform state mv` commands. Instead, since the `aws_route53_record` resource implements the `allow_overwrite = true` argument, it is possible to just remove the old `aws_route53_record` resources from the Terraform state using the [`terraform state rm` command](https://www.terraform.io/docs/commands/state/rm.html). In this case, Terraform will leave the existing records in Route 53 and plan to just overwrite the existing validation records with the same exact (previous) values.

-> This guide is showing the simpler `terraform state rm` option below as a potential shortcut in this specific situation, however in most other cases `terraform state mv` is required to change from `count` based resources to `for_each` based resources and properly match the existing Terraform state to the updated Terraform configuration.

```console
$ terraform state rm aws_route53_record.existing
Removed aws_route53_record.existing[0]
Removed aws_route53_record.existing[1]
Removed aws_route53_record.existing[2]
Removed aws_route53_record.existing[3]
Successfully removed 4 resource instance(s).
```

Now the Terraform plan will show only the additions of new Route 53 records (which are exactly the same as before the upgrade) and the proposed recreation of the `aws_acm_certificate_validation` resource. The `aws_acm_certificate_validation` resource recreation will have no effect as the certificate is already validated and issued.

```
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create
-/+ destroy and then create replacement

Terraform will perform the following actions:

  # aws_acm_certificate_validation.existing must be replaced
-/+ resource "aws_acm_certificate_validation" "existing" {
        certificate_arn         = "arn:aws:acm:us-east-2:123456789012:certificate/ccbc58e8-061d-4443-9035-d3af0512e863"
      ~ id                      = "2020-07-16 00:01:19 +0000 UTC" -> (known after apply)
      ~ validation_record_fqdns = [
          - "_40b71647a8d88eb82d53fe988e8a3cc1.existing2.example.com",
          - "_812ddf11b781af1eec1643ec58f102d2.existing.example.com",
          - "_8dc56b6e35f699b8754afcdd79e9748d.existing3.example.com",
          - "_d7112da809a40e848207c04399babcec.existing1.example.com",
        ] -> (known after apply) # forces replacement
    }

  # aws_route53_record.existing["existing.example.com"] will be created
  + resource "aws_route53_record" "existing" {
      + allow_overwrite = true
      + fqdn            = (known after apply)
      + id              = (known after apply)
      + name            = "_812ddf11b781af1eec1643ec58f102d2.existing.example.com"
      + records         = [
          + "_bdeba72164eec216c55a32374bcceafd.jfrzftwwjs.acm-validations.aws.",
        ]
      + ttl             = 60
      + type            = "CNAME"
      + zone_id         = "Z123456789012"
    }

  # aws_route53_record.existing["existing1.example.com"] will be created
  + resource "aws_route53_record" "existing" {
      + allow_overwrite = true
      + fqdn            = (known after apply)
      + id              = (known after apply)
      + name            = "_d7112da809a40e848207c04399babcec.existing1.example.com"
      + records         = [
          + "_6e1da5574ab46a6c782ed73438274181.jfrzftwwjs.acm-validations.aws.",
        ]
      + ttl             = 60
      + type            = "CNAME"
      + zone_id         = "Z123456789012"
    }

  # aws_route53_record.existing["existing2.example.com"] will be created
  + resource "aws_route53_record" "existing" {
      + allow_overwrite = true
      + fqdn            = (known after apply)
      + id              = (known after apply)
      + name            = "_40b71647a8d88eb82d53fe988e8a3cc1.existing2.example.com"
      + records         = [
          + "_638532db1fa6a1b71aaf063c8ea29d52.jfrzftwwjs.acm-validations.aws.",
        ]
      + ttl             = 60
      + type            = "CNAME"
      + zone_id         = "Z123456789012"
    }

  # aws_route53_record.existing["existing3.example.com"] will be created
  + resource "aws_route53_record" "existing" {
      + allow_overwrite = true
      + fqdn            = (known after apply)
      + id              = (known after apply)
      + name            = "_8dc56b6e35f699b8754afcdd79e9748d.existing3.example.com"
      + records         = [
          + "_a419f8410d2e0720528a96c3506f3841.jfrzftwwjs.acm-validations.aws.",
        ]
      + ttl             = 60
      + type            = "CNAME"
      + zone_id         = "Z123456789012"
    }

Plan: 5 to add, 0 to change, 1 to destroy.
```

Once applied, no differences should be shown and no additional steps should be necessary.

### subject_alternative_names Changed from List to Set

Previously the `subject_alternative_names` argument was stored in the Terraform state as an ordered list while the API returned information in an unordered manner. The attribute is now configured as a set instead of a list. Certain Terraform configuration language features distinguish between these two attribute types such as not being able to index a set (e.g. `aws_acm_certificate.example.subject_alternative_names[0]` is no longer a valid reference). Depending on the implementation details of a particular configuration using `subject_alternative_names` as a reference, possible solutions include changing references to using `for`/`for_each` or using the `tolist()` function as a temporary workaround to keep the previous behavior until an appropriate configuration (properly using the unordered set) can be determined. Usage questions can be submitted to the [community forums](https://discuss.hashicorp.com/c/terraform-providers/tf-aws/33).

### certificate_body, certificate_chain, and private_key Arguments No Longer Stored as Hash

Previously when the `certificate_body`, `certificate_chain`, and `private_key` arguments were stored in state, they were stored as a hash of the actual value. This prevented Terraform from properly updating the resource when necessary and the hashing has been removed. The Terraform AWS Provider will show an update to these arguments on the first apply after upgrading to version 3.0.0, which is fixing the Terraform state to remove the hash. Since the `private_key` attribute is marked as sensitive, the values in the update will not be visible in the Terraform output. If the non-hashed values have not changed, then no update is occurring other than the Terraform state update. If these arguments are the only updates and they all match the hash removal, the apply will occur without submitting API calls.

## Resource: aws_api_gateway_method_settings

### throttling_burst_limit and throttling_rate_limit Arguments Now Default to -1

Previously when the `throttling_burst_limit` or `throttling_rate_limit` argument was not configured, the resource would enable throttling and set the limit value to the AWS API Gateway default. In addition, as these arguments were marked as `Computed`, Terraform ignored any subsequent changes made to these arguments in the resource. These behaviors have been removed and, by default, the `throttling_burst_limit` and `throttling_rate_limit` arguments will be disabled in the resource with a value of `-1`.

## Resource: aws_autoscaling_group

### availability_zones and vpc_zone_identifier Arguments Now Report Plan-Time Conflict

Specifying both the `availability_zones` and `vpc_zone_identifier` arguments previously led to confusing behavior and errors. Now this issue is reported at plan-time. Use the `null` value instead of `[]` (empty list) in conditionals to ensure this validation does not unexpectedly trigger.

### Drift detection enabled for `load_balancers` and `target_group_arns` arguments

If you previously set one of these arguments to an empty list to enable drift detection (e.g. when migrating an ASG from ELB to ALB), this can be updated as follows.

For example, given this previous configuration:

```hcl
resource "aws_autoscaling_group" "example" {
  # ... other configuration ...
  load_balancers    = []
  target_group_arns = [aws_lb_target_group.example.arn]
}
```

An updated configuration:

```hcl
resource "aws_autoscaling_group" "example" {
  # ... other configuration ...
  target_group_arns = [aws_lb_target_group.example.arn]
}
```

If `aws_autoscaling_attachment` resources reference your ASG configurations, you will need to add the [`lifecycle` configuration block](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html) with an `ignore_changes` argument to prevent Terraform non-empty plans (i.e. forcing resource update) during the next state refresh.

For example, given this previous configuration:

```hcl
resource "aws_autoscaling_attachment" "example" {
  autoscaling_group_name = aws_autoscaling_group.example.id
  elb                    = aws_elb.example.id
}

resource "aws_autoscaling_group" "example" {
  # ... other configuration ...
}
```

An updated configuration:

```hcl
resource "aws_autoscaling_attachment" "example" {
  autoscaling_group_name = aws_autoscaling_group.example.id
  elb                    = aws_elb.example.id
}

resource "aws_autoscaling_group" "example" {
  # ... other configuration ...

  lifecycle {
    ignore_changes = [load_balancers, target_group_arns]
  }
}
```

## Resource: aws_cloudfront_distribution

### active_trusted_signers Attribute Name and Type Change

Previously, the `active_trusted_signers` computed attribute was implemented with a Map that did not support accessing its computed `items` attribute in Terraform 0.12 correctly.
To address this, the `active_trusted_signers` attribute has been renamed to `trusted_signers` and is now implemented as a List with a computed `items` List attribute and computed `enabled` boolean attribute.
The nested `items` attribute includes computed `aws_account_number` and `key_pair_ids` sub-fields, with the latter implemented as a List.
Thus, user configurations referencing the `active_trusted_signers` attribute and its sub-fields will need to be changed as follows.

Given these previous references:

```
aws_cloudfront_distribution.example.active_trusted_signers.enabled
aws_cloudfront_distribution.example.active_trusted_signers.items
```

Updated references:

```
aws_cloudfront_distribution.example.trusted_signers[0].enabled
aws_cloudfront_distribution.example.trusted_signers[0].items
```

## Resource: aws_cloudwatch_log_group

### Removal of arn Wildcard Suffix

Previously, the resource returned the Amazon Resource Name (ARN) directly from the API, which included a `:*` suffix to denote all CloudWatch Log Streams under the CloudWatch Log Group. Most other AWS resources that return ARNs and many other AWS services do not use the `:*` suffix. The suffix is now automatically removed. For example, the resource previously returned an ARN such as `arn:aws:logs:us-east-1:123456789012:log-group:/example:*` but will now return `arn:aws:logs:us-east-1:123456789012:log-group:/example`.

Workarounds, such as using `replace()` as shown below, should be removed:

```hcl
resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}
resource "aws_datasync_task" "example" {
  # ... other configuration ...
  cloudwatch_log_group_arn = replace(aws_cloudwatch_log_group.example.arn, ":*", "")
}
```

Removing the `:*` suffix is a breaking change for some configurations. Fix these configurations using string interpolations as demonstrated below. For example, this configuration is now broken:

```hcl
data "aws_iam_policy_document" "ad-log-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    principals {
      identifiers = ["ds.amazonaws.com"]
      type        = "Service"
    }
    resources = [aws_cloudwatch_log_group.example.arn]
    effect = "Allow"
  }
}
```

An updated configuration:

```hcl
data "aws_iam_policy_document" "ad-log-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    principals {
      identifiers = ["ds.amazonaws.com"]
      type        = "Service"
    }
    resources = ["${aws_cloudwatch_log_group.example.arn}:*"]
    effect = "Allow"
  }
}
```

## Resource: aws_codepipeline

### GITHUB_TOKEN environment variable removal

Switch your Terraform configuration to the `OAuthToken` element in the `action` `configuration` map instead.

For example, given this previous configuration:

```bash
$ GITHUB_TOKEN=<token> terraform apply
```

```hcl
resource "aws_codepipeline" "example" {
  # ... other configuration ...

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["example"]

      configuration = {
        Owner  = "lifesum-terraform"
        Repo   = "example"
        Branch = "main"
      }
    }
  }
}
```

The configuration could be updated as follows:

```bash
$ TF_VAR_github_token=<token> terraform apply
```

```hcl
variable "github_token" {}

resource "aws_codepipeline" "example" {
  # ... other configuration ...

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["example"]

      configuration = {
        Owner      = "lifesum-terraform"
        Repo       = "example"
        Branch     = "main"
        OAuthToken = var.github_token
      }
    }
  }
}
```

## Resource: aws_cognito_user_pool

### Removal of admin_create_user_config.unused_account_validity_days Argument

The Cognito API previously deprecated the `admin_create_user_config` configuration block `unused_account_validity_days` argument in preference of the `password_policy` configuration block `temporary_password_validity_days` argument. Configurations will need to be updated to use the API supported configuration.

For example, given this previous configuration:

```hcl
resource "aws_cognito_user_pool" "example" {
  # ... other configuration ...

  admin_create_user_config {
    # ... potentially other configuration ...

    unused_account_validity_days = 7
  }
}
```

An updated configuration:

```hcl
resource "aws_cognito_user_pool" "example" {
  # ... other configuration ...

  password_policy {
    # ... potentially other configuration ...

    temporary_password_validity_days = 7
  }
}
```

## Resource: aws_dx_gateway

### Removal of Automatic aws_dx_gateway_association Import

Previously when importing the `aws_dx_gateway` resource with the [`terraform import` command](https://www.terraform.io/docs/commands/import.html), the Terraform AWS Provider would automatically attempt to import an associated `aws_dx_gateway_association` resource(s) as well. This automatic resource import has been removed. Use the [`aws_dx_gateway_association` resource import](/docs/providers/aws/r/dx_gateway_association.html#import) to import those resources separately.

## Resource: aws_dx_gateway_association

### vpn_gateway_id Argument Removal

Switch your Terraform configuration to the `associated_gateway_id` argument instead.

For example, given this previous configuration:

```hcl
resource "aws_dx_gateway_association" "example" {
  # ... other configuration ...
  vpn_gateway_id = aws_vpn_gateway.example.id
}
```

An updated configuration:

```hcl
resource "aws_dx_gateway_association" "example" {
  # ... other configuration ...
  associated_gateway_id = aws_vpn_gateway.example.id
}
```

## Resource: aws_dx_gateway_association_proposal

### vpn_gateway_id Argument Removal

Switch your Terraform configuration to the `associated_gateway_id` argument instead.

For example, given this previous configuration:

```hcl
resource "aws_dx_gateway_association_proposal" "example" {
  # ... other configuration ...
  vpn_gateway_id = aws_vpn_gateway.example.id
}
```

An updated configuration:

```hcl
resource "aws_dx_gateway_association_proposal" "example" {
  # ... other configuration ...
  associated_gateway_id = aws_vpn_gateway.example.id
}
```

## Resource: aws_ebs_volume

### iops Argument Apply-Time Validation

Previously when the `iops` argument was configured with a `type` other than `io1` (either explicitly or omitted, indicating the default type `gp2`), the Terraform AWS Provider would automatically disregard the value provided to `iops` as it is only configurable for the `io1` volume type per the AWS EC2 API. This behavior has changed such that the Terraform AWS Provider will instead return an error at apply time indicating an `iops` value is invalid for types other than `io1`.
Exceptions to this are in cases where `iops` is set to `null` or `0` such that the Terraform AWS Provider will continue to accept the value regardless of `type`.

## Resource: aws_elastic_transcoder_preset

### video Configuration Block max_frame_rate Argument No Longer Uses 30 Default

Previously when the `max_frame_rate` argument was not configured, the resource would default to 30. This behavior has been removed and allows for auto frame rate presets to automatically set the appropriate value.

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
    instance_role = "MASTER"
    instance_type = "m4.large"
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
  cluster_id     = aws_emr_cluster.example.id
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

## Resource: aws_glue_job

### allocated_capacity Argument Removal

The Glue API has deprecated the `allocated_capacity` argument. Switch your Terraform configuration to the `max_capacity` argument instead.

For example, given this previous configuration:

```hcl
resource "aws_glue_job" "example" {
  # ... other configuration ...

  allocated_capacity = 2
}
```

An updated configuration:

```hcl
resource "aws_glue_job" "example" {
  # ... other configuration ...

  max_capacity = 2
}
```

## Resource: aws_iam_access_key

### ses_smtp_password Attribute Removal

In many regions today and in all regions after October 1, 2020, the [SES API will only accept version 4 signatures](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/using-ses-api-authentication.html). If referencing the `ses_smtp_password` attribute, switch your Terraform configuration to the `ses_smtp_password_v4` attribute instead. Please note that this signature is based on the region of the Terraform AWS Provider. If you need the SES v4 password in multiple regions, it may require using [multiple provider instances](https://www.terraform.io/docs/configuration/providers.html#alias-multiple-provider-configurations).

## Resource: aws_iam_instance_profile

### roles Argument Removal

Switch your Terraform configuration to the `role` argument instead.

For example, given this previous configuration:

```hcl
resource "aws_iam_instance_profile" "example" {
  # ... other configuration ...

  roles = [aws_iam_role.example.id]
}
```

An updated configuration:

```hcl
resource "aws_iam_instance_profile" "example" {
  # ... other configuration ...

  role = aws_iam_role.example.id
}
```

## Resource: aws_iam_server_certificate

### certificate_body, certificate_chain, and private_key Arguments No Longer Stored as Hash

Previously when the `certificate_body`, `certificate_chain`, and `private_key` arguments were stored in state, they were stored as a hash of the actual value. This hashing has been removed for new or recreated resources to prevent lifecycle issues.

## Resource: aws_instance

### ebs_block_device.iops and root_block_device.iops Argument Apply-Time Validations

Previously when the `iops` argument was configured in either the `ebs_block_device` or `root_block_device` configuration block, the Terraform AWS Provider would automatically disregard the value provided to `iops` if the `type` argument was also configured with a value other than `io1` (either explicitly or omitted, indicating the default type `gp2`) as `iops` are only configurable for the `io1` volume type per the AWS EC2 API. This behavior has changed such that the Terraform AWS Provider will instead return an error at apply time indicating an `iops` value is invalid for volume types other than `io1`.
Exceptions to this are in cases where `iops` is set to `null` or `0` such that the Terraform AWS Provider will continue to accept the value regardless of `type`.

## Resource: aws_lambda_alias

### Import No Longer Converts Function Name to ARN

Previously the resource import would always convert the `function_name` portion of the import identifier into the ARN format. Configurations using the Lambda Function name would show this as an unexpected difference after import. Now this will passthrough the given value on import whether its a Lambda Function name or ARN.

## Resource: aws_launch_template

### network_interfaces.delete_on_termination Argument type change

The `network_interfaces.delete_on_termination` argument is now of type `string`, allowing an unspecified value for the argument since the previous `bool` type only allowed for `true/false` and defaulted to `false` when no value was set. Now to enforce `delete_on_termination` to `false`, the string `"false"` or bare `false` value must be used.

For example, given this previous configuration:

```hcl
resource "aws_launch_template" "example" {
  # ... other configuration ...

  network_interfaces {
    # ... other configuration ...

    delete_on_termination = null
  }
}
```

An updated configuration:

```hcl
resource "aws_launch_template" "example" {
  # ... other configuration ...

  network_interfaces {
    # ... other configuration ...

    delete_on_termination = false
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

## Resource: aws_msk_cluster

### encryption_info.encryption_in_transit.client_broker Default Updated to Match API

A few weeks after general availability launch and initial release of the `aws_msk_cluster` resource, the MSK API default for client broker encryption switched from `TLS_PLAINTEXT` to `TLS`. The attribute default has now been updated to match the more secure API default, however existing Terraform configurations may show a difference if this setting is not configured.

To continue using the old default when it was previously not configured, add or modify this configuration:

```hcl
resource "aws_msk_cluster" "example" {
  # ... other configuration ...

  encryption_info {
    # ... potentially other configuration ...

    encryption_in_transit {
      # ... potentially other configuration ...

      client_broker = "TLS_PLAINTEXT"
    }
  }
}
```

## Resource: aws_rds_cluster

### scaling_configuration.min_capacity Now Defaults to 1

Previously when the `min_capacity` argument in a `scaling_configuration` block was not configured, the resource would default to 2. This behavior has been updated to align with the AWS RDS Cluster API default of 1.

## Resource: aws_route53_resolver_rule

### Removal of trailing period in domain_name argument

Previously the resource returned the Resolver Rule Domain Name directly from the API, which included a `.` suffix. This proves difficult when many other AWS services do not accept this trailing period (e.g. ACM Certificate). This period is now automatically removed. For example, when the attribute would previously return a Resolver Rule Domain Name such as `example.com.`, the attribute now will be returned as `example.com`.
While the returned value will omit the trailing period, use of configurations with trailing periods will not be interrupted.

## Resource: aws_route53_zone

### Removal of trailing period in name argument

Previously the resource returned the Hosted Zone Domain Name directly from the API, which included a `.` suffix. This proves difficult when many other AWS services do not accept this trailing period (e.g. ACM Certificate). This period is now automatically removed. For example, when the attribute would previously return a Hosted Zone Domain Name such as `example.com.`, the attribute now will be returned as `example.com`.
While the returned value will omit the trailing period, use of configurations with trailing periods will not be interrupted.

## Resource: aws_s3_bucket

### Removal of Automatic aws_s3_bucket_policy Import

Previously when importing the `aws_s3_bucket` resource with the [`terraform import` command](https://www.terraform.io/docs/commands/import.html), the Terraform AWS Provider would automatically attempt to import an associated `aws_s3_bucket_policy` resource as well. This automatic resource import has been removed. Use the [`aws_s3_bucket_policy` resource import](/docs/providers/aws/r/s3_bucket_policy.html#import) to import that resource separately.

### region Attribute Is Now Read-Only

The `region` attribute is no longer configurable, but it remains as a read-only attribute. The region of the `aws_s3_bucket` resource is determined by the region of the Terraform AWS Provider, similar to all other resources.

For example, given this previous configuration:

```hcl
resource "aws_s3_bucket" "example" {
  # ... other configuration ...

  region = "us-west-2"
}
```

An updated configuration:

```hcl
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
}
```

## Resource: aws_s3_bucket_metric

### filter configuration block Plan-Time Validation Change

The `filter` configuration block no longer supports the empty block `{}` and requires at least one of the `prefix` or `tags` attributes to be specified.

For example, given this previous configuration:

```hcl
resource "aws_s3_bucket_metric" "example" {
  # ... other configuration ...

  filter {}
}
```

An updated configuration:

```hcl
resource "aws_s3_bucket_metric" "example" {
  # ... other configuration ...
}
```

## Resource: aws_security_group

### Removal of Automatic aws_security_group_rule Import

Previously when importing the `aws_security_group` resource with the [`terraform import` command](https://www.terraform.io/docs/commands/import.html), the Terraform AWS Provider would automatically attempt to import an associated `aws_security_group_rule` resource(s) as well. This automatic resource import has been removed. Use the [`aws_security_group_rule` resource import](/docs/providers/aws/r/security_group_rule.html#import) to import those resources separately.

## Resource: aws_sns_platform_application

### platform_credential and platform_principal Arguments No Longer Stored as SHA256 Hash

Previously when the `platform_credential` and `platform_principal` arguments were stored in state, they were stored as a SHA256 hash of the actual value. This prevented Terraform from properly updating the resource when necessary and the hashing has been removed. The Terraform AWS Provider will show an update to these arguments on the first apply after upgrading to version 3.0.0, which is fixing the Terraform state to remove the hash. Since the attributes are marked as sensitive, the values in the update will not be visible in the Terraform output. If the non-hashed values have not changed, then no update is occurring other than the Terraform state update. If these arguments are the only two updates and they both match the SHA256 removal, the apply will occur without submitting an actual `SetPlatformApplicationAttributes` API call.

## Resource: aws_spot_fleet_request

### valid_until Argument No Longer Uses 24 Hour Default

Previously when the `valid_until` argument was not configured, the resource would default to a 24 hour request. This behavior has been removed and allows for non-expiring requests. To recreate the old behavior, the [`time_offset` resource](https://registry.terraform.io/providers/hashicorp/time/latest/docs/resources/offset) can potentially be used.

## Resource: aws_ssm_maintenance_window_task

### logging_info Configuration Block Removal

Switch your Terraform configuration to the `task_invocation_parameters` configuration block `run_command_parameters` configuration block `output_s3_bucket` and `output_s3_key_prefix` arguments instead.

For example, given this previous configuration:

```hcl
resource "aws_ssm_maintenance_window_task" "example" {
  # ... other configuration ...

  logging_info {
    s3_bucket_name       = aws_s3_bucket.example.id
    s3_bucket_key_prefix = "example"
  }
}
```

An updated configuration:

```hcl
resource "aws_ssm_maintenance_window_task" "example" {
  # ... other configuration ...

  task_invocation_parameters {
    # ... potentially other configuration ...

    run_command_parameters {
      # ... potentially other configuration ...

      output_s3_bucket     = aws_s3_bucket.example.id
      output_s3_key_prefix = "example"
    }
  }
}
```

### task_parameters Configuration Block Removal

Switch your Terraform configuration to the `task_invocation_parameters` configuration block `run_command_parameters` configuration block `parameter` configuration blocks instead.

For example, given this previous configuration:

```hcl
resource "aws_ssm_maintenance_window_task" "example" {
  # ... other configuration ...

  task_parameters {
    name   = "commands"
    values = ["date"]
  }
}
```

An updated configuration:

```hcl
resource "aws_ssm_maintenance_window_task" "example" {
  # ... other configuration ...

  task_invocation_parameters {
    # ... potentially other configuration ...

    run_command_parameters {
      # ... potentially other configuration ...

      parameter {
        name   = "commands"
        values = ["date"]
      }
    }
  }
}
```
