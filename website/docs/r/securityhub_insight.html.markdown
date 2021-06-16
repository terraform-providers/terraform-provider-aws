---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_insight"
description: |-
  Provides a Security Hub custom insight resource.
---

# Resource: aws_securityhub_insight

Provides a Security Hub custom insight resource. See the [Managing custom insights section](https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-custom-insights.html) of the AWS User Guide for more information.

## Example Usage

### Filter by AWS account ID

```terraform
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_insight" "example" {
  filters {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }

    aws_account_id {
      comparison = "EQUALS"
      value      = "09876543210"
    }
  }

  group_by_attribute = "AwsAccountId"

  name = "example-insight"

  depends_on = [aws_securityhub_account.example]
}
```

### Filter by date range

```terraform
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_insight" "example" {
  filters {
    created_at {
      date_range {
        unit  = "DAYS"
        value = 5
      }
    }
  }

  group_by_attribute = "CreatedAt"

  name = "example-insight"

  depends_on = [aws_securityhub_account.example]
}
```

### Filter by destination IPv4 address

```terraform
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_insight" "example" {
  filters {
    network_destination_ipv4 {
      cidr = "10.0.0.0/16"
    }
  }

  group_by_attribute = "NetworkDestinationIpV4"

  name = "example-insight"

  depends_on = [aws_securityhub_account.example]
}
```

### Filter by finding's confidence

```terraform
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_insight" "example" {
  filters {
    confidence {
      gte = "80"
    }
  }

  group_by_attribute = "Confidence"

  name = "example-insight"

  depends_on = [aws_securityhub_account.example]
}
```

### Filter by resource tags

```terraform
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_insight" "example" {
  filters {
    resource_tags {
      comparison = "EQUALS"
      key        = "Environment"
      value      = "Production"
    }
  }

  group_by_attribute = "ResourceTags"

  name = "example-insight"

  depends_on = [aws_securityhub_account.example]
}
```

## Argument Reference

The following arguments are required:

* `filters` - (Required) A configuration block including one or more (up to 10 distinct) attributes used to filter the findings included in the insight. The insight only includes findings that match criteria defined in the filters. See [filters](#filters) below for more details.
* `group_by_attribute` - (Required) The attribute used to group the findings for the insight e.g. if an insight is grouped by `ResourceId`, then the insight produces a list of resource identifiers.
* `name` - (Required) The name of the custom insight.

### filters

The `filters` configuration block supports the following arguments:

~> **NOTE:** For each argument below, up to 20 can be provided.

* `aws_account_id` - (Optional) AWS account ID that a finding is generated in. See [String_Filter](#string-filter-argument-reference) below for more details.
* `company_name` - (Optional) The name of the findings provider (company) that owns the solution (product) that generates findings. See [String_Filter](#string-filter-argument-reference) below for more details.
* `compliance_status` - (Optional) Exclusive to findings that are generated as the result of a check run against a specific rule in a supported standard, such as CIS AWS Foundations. Contains security standard-related finding details. See [String Filter](#string-filter-argument-reference) below for more details.
* `confidence` - (Optional) A finding's confidence. Confidence is defined as the likelihood that a finding accurately identifies the behavior or issue that it was intended to identify. Confidence is scored on a 0-100 basis using a ratio scale, where 0 means zero percent confidence and 100 means 100 percent confidence. See [Number Filter](#number-filter-argument-reference) below for more details.
* `created_at` - (Optional) An ISO8601-formatted timestamp that indicates when the security-findings provider captured the potential security issue that a finding captured. See [Date Filter](#date-filter-argument-reference) below for more details.
* `criticality` - (Optional) The level of importance assigned to the resources associated with the finding. A score of 0 means that the underlying resources have no criticality, and a score of 100 is reserved for the most critical resources. See [Number Filter](#number-filter-argument-reference) below for more details.
* `description` - (Optional) A finding's description. See [String Filter](#string-filter-argument-reference) below for more details.
* `finding_provider_fields_confidence` - (Optional) The finding provider value for the finding confidence. Confidence is defined as the likelihood that a finding accurately identifies the behavior or issue that it was intended to identify. Confidence is scored on a 0-100 basis using a ratio scale, where 0 means zero percent confidence and 100 means 100 percent confidence. See [Number Filter](#number-filter-argument-reference) below for more details.
* `finding_provider_fields_criticality` - (Optional) The finding provider value for the level of importance assigned to the resources associated with the findings. A score of 0 means that the underlying resources have no criticality, and a score of 100 is reserved for the most critical resources. See [Number Filter](#number-filter-argument-reference) below for more details.
* `finding_provider_fields_related_findings_id` - (Optional) The finding identifier of a related finding that is identified by the finding provider. See [String Filter](#string-filter-argument-reference) below for more details.
* `finding_provider_fields_related_findings_product_arn` - (Optional) The ARN of the solution that generated a related finding that is identified by the finding provider. See [String Filter](#string-filter-argument-reference) below for more details.
* `finding_provider_fields_severity_label` - (Optional) The finding provider value for the severity label. See [String Filter](#string-filter-argument-reference) below for more details.
* `finding_provider_fields_severity_original` - (Optional) The finding provider's original value for the severity. See [String Filter](#string-filter-argument-reference) below for more details.
* `finding_provider_fields_types` - (Optional) One or more finding types that the finding provider assigned to the finding. Uses the format of `namespace/category/classifier` that classify a finding. Valid namespace values include: `Software and Configuration Checks`, `TTPs`, `Effects`, `Unusual Behaviors`, and `Sensitive Data Identifications`. See [String Filter](#string-filter-argument-reference) below for more details.
* `first_observed_at` - (Optional) An ISO8601-formatted timestamp that indicates when the security-findings provider first observed the potential security issue that a finding captured. See [Date Filter](#date-filter-argument-reference) below for more details.
* `generator_id` - (Optional) The identifier for the solution-specific component (a discrete unit of logic) that generated a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `id` - (Optional) The security findings provider-specific identifier for a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `keyword` - (Optional) A keyword for a finding. See [Keyword Filter](#keyword-filter-argument-reference) below for more details.
* `last_observed_at` - (Optional) An ISO8601-formatted timestamp that indicates when the security-findings provider most recently observed the potential security issue that a finding captured. See [Date Filter](#date-filter-argument-reference) below for more details.
* `malware_name` - (Optional) The name of the malware that was observed. See [String Filter](#string-filter-argument-reference) below for more details.
* `malware_path` - (Optional) The filesystem path of the malware that was observed. See [String Filter](#string-filter-argument-reference) below for more details.
* `malware_state` - (Optional) The state of the malware that was observed. See [String Filter](#string-filter-argument-reference) below for more details.
* `malware_type` - (Optional) The type of the malware that was observed. See [String Filter](#string-filter-argument-reference) below for more details.
* `network_destination_domain` - (Optional) The destination domain of network-related information about a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `network_destination_ipv4` - (Optional) The destination IPv4 address of network-related information about a finding. See [Ip Filter](#ip-filter-argument-reference) below for more details.
* `network_destination_ipv6` - (Optional) The destination IPv6 address of network-related information about a finding. See [Ip Filter](#ip-filter-argument-reference) below for more details.
* `network_destination_port` - (Optional) The destination port of network-related information about a finding. See [Number Filter](#number-filter-argument-reference) below for more details.
* `network_direction` - (Optional) Indicates the direction of network traffic associated with a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `network_protocol` - (Optional) The protocol of network-related information about a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `network_source_domain` - (Optional) The source domain of network-related information about a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `network_source_ipv4` - (Optional) The source IPv4 address of network-related information about a finding. See [Ip Filter](#ip-filter-argument-reference) below for more details.
* `network_source_ipv6` - (Optional) The source IPv6 address of network-related information about a finding. See [Ip Filter](#ip-filter-argument-reference) below for more details.
* `network_source_mac` - (Optional) The source media access control (MAC) address of network-related information about a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `network_source_port` - (Optional) The source port of network-related information about a finding. See [Number Filter](#number-filter-argument-reference) below for more details.
* `note_text` - (Optional) The text of a note. See [String Filter](#string-filter-argument-reference) below for more details.
* `note_updated_at` - (Optional) The timestamp of when the note was updated. See [Date Filter](#date-filter-argument-reference) below for more details.
* `note_updated_by` - (Optional) The principal that created a note. See [String Filter](#string-filter-argument-reference) below for more details.
* `process_launched_at` - (Optional) The date/time that the process was launched. See [Date Filter](#date-filter-argument-reference) below for more details.
* `process_name` - (Optional) The name of the process. See [String Filter](#string-filter-argument-reference) below for more details.
* `process_parent_pid` - (Optional) The parent process ID. See [Number Filter](#number-filter-argument-reference) below for more details.
* `process_path` - (Optional) The path to the process executable. See [String Filter](#string-filter-argument-reference) below for more details.
* `process_pid` - (Optional) The process ID. See [Number Filter](#number-filter-argument-reference) below for more details.
* `process_terminated_at` - (Optional) The date/time that the process was terminated. See [Date Filter](#date-filter-argument-reference) below for more details.
* `product_arn` - (Optional) The ARN generated by Security Hub that uniquely identifies a third-party company (security findings provider) after this provider's product (solution that generates findings) is registered with Security Hub. See [String Filter](#string-filter-argument-reference) below for more details.
* `product_fields` - (Optional) A data type where security-findings providers can include additional solution-specific details that aren't part of the defined `AwsSecurityFinding` format. See [Map Filter](#map-filter-argument-reference) below for more details.
* `product_name` - (Optional) The name of the solution (product) that generates findings. See [String Filter](#string-filter-argument-reference) below for more details.
* `recommendation_text` - (Optional) The recommendation of what to do about the issue described in a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `record_state` - (Optional) The updated record state for the finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `related_findings_id` - (Optional) The solution-generated identifier for a related finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `related_findings_product_arn` - (Optional) The ARN of the solution that generated a related finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_aws_ec2_instance_iam_instance_profile_arn` - (Optional) The IAM profile ARN of the instance. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_aws_ec2_instance_image_id` - (Optional) The Amazon Machine Image (AMI) ID of the instance. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_aws_ec2_instance_ipv4_addresses` - (Optional) The IPv4 addresses associated with the instance. See [Ip Filter](#ip-filter-argument-reference) below for more details.
* `resource_aws_ec2_instance_ipv6_addresses` - (Optional) The IPv6 addresses associated with the instance. See [Ip Filter](#ip-filter-argument-reference) below for more details.
* `resource_aws_ec2_instance_key_name` - (Optional) The key name associated with the instance. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_aws_ec2_instance_launched_at` - (Optional) The date and time the instance was launched. See [Date Filter](#date-filter-argument-reference) below for more details.
* `resource_aws_ec2_instance_subnet_id` - (Optional) The identifier of the subnet that the instance was launched in. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_aws_ec2_instance_type` - (Optional) The instance type of the instance. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_aws_ec2_instance_vpc_id` - (Optional) The identifier of the VPC that the instance was launched in. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_aws_iam_access_key_created_at` - (Optional) The creation date/time of the IAM access key related to a finding. See [Date Filter](#date-filter-argument-reference) below for more details.
* `resource_aws_iam_access_key_status` - (Optional) The status of the IAM access key related to a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_aws_iam_access_key_user_name` - (Optional) The user associated with the IAM access key related to a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_aws_s3_bucket_owner_id` - (Optional) The canonical user ID of the owner of the S3 bucket. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_aws_s3_bucket_owner_name` - (Optional) The display name of the owner of the S3 bucket. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_container_image_id` - (Optional) The identifier of the image related to a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_container_image_name` - (Optional) The name of the image related to a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_container_launched_at` - (Optional) The date/time that the container was started. See [Date Filter](#date-filter-argument-reference) below for more details.
* `resource_container_name` - (Optional) The name of the container related to a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_details_other` - (Optional) The details of a resource that doesn't have a specific subfield for the resource type defined. See [Map Filter](#map-filter-argument-reference) below for more details.
* `resource_id` - (Optional) The canonical identifier for the given resource type. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_partition` - (Optional) The canonical AWS partition name that the Region is assigned to. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_region` - (Optional) The canonical AWS external Region name where this resource is located. See [String Filter](#string-filter-argument-reference) below for more details.
* `resource_tags` - (Optional) A list of AWS tags associated with a resource at the time the finding was processed. See [Map Filter](#map-filter-argument-reference) below for more details.
* `resource_type` - (Optional) Specifies the type of the resource that details are provided for. See [String Filter](#string-filter-argument-reference) below for more details.
* `severity_label` - (Optional) The label of a finding's severity. See [String Filter](#string-filter-argument-reference) below for more details.
* `source_url` - (Optional) A URL that links to a page about the current finding in the security-findings provider's solution. See [String Filter](#string-filter-argument-reference) below for more details.
* `threat_intel_indicator_category` - (Optional) The category of a threat intelligence indicator. See [String Filter](#string-filter-argument-reference) below for more details.
* `threat_intel_indicator_last_observed_at` - (Optional) The date/time of the last observation of a threat intelligence indicator. See [Date Filter](#date-filter-argument-reference) below for more details.
* `threat_intel_indicator_source` - (Optional) The source of the threat intelligence. See [String Filter](#string-filter-argument-reference) below for more details.
* `threat_intel_indicator_source_url` - (Optional) The URL for more details from the source of the threat intelligence. See [String Filter](#string-filter-argument-reference) below for more details.
* `threat_intel_indicator_type` - (Optional) The type of a threat intelligence indicator. See [String Filter](#string-filter-argument-reference) below for more details.
* `threat_intel_indicator_value` - (Optional) The value of a threat intelligence indicator. See [String Filter](#string-filter-argument-reference) below for more details.
* `title` - (Optional) A finding's title. See [String Filter](#string-filter-argument-reference) below for more details.
* `type` - (Optional) A finding type in the format of `namespace/category/classifier` that classifies a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `updated_at` - (Optional) An ISO8601-formatted timestamp that indicates when the security-findings provider last updated the finding record. See [Date Filter](#date-filter-argument-reference) below for more details.
* `user_defined_values` - (Optional) A list of name/value string pairs associated with the finding. These are custom, user-defined fields added to a finding. See [Map Filter](#map-filter-argument-reference) below for more details.
* `verification_state` - (Optional) The veracity of a finding. See [String Filter](#string-filter-argument-reference) below for more details.
* `workflow_status` - (Optional) The status of the investigation into a finding. See [Workflow Status Filter](#workflow-status-filter-argument-reference) below for more details.

### Date Filter Argument reference

The date filter configuration block supports the following arguments:

* `date_range` - (Optional) A configuration block of the date range for the date filter. See [date_range](#date_range-argument-reference) below for more details.
* `end` - (Optional) An end date for the date filter. Required with `start` if `date_range` is not specified.
* `start` - (Optional) A start date for the date filter. Required with `end` if `date_range` is not specified.

### date_range Argument reference

The `date_range` configuration block supports the following arguments:

* `unit` - (Required) A date range unit for the date filter. Valid values: `DAYS`.
* `value` - (Required) A date range value for the date filter, provided as an Integer.

### Ip Filter Argument Reference

The Ip filter configuration block supports the following arguments:

* `cidr` - (Required) A finding's CIDR value.

### Keyword Filter Argument Reference

The keyword filter configuration block supports the following arguments:

* `value` - (Required) A value for the keyword.

### Map Filter Argument reference

The map filter configuration block supports the following arguments:

* `comparison` - (Required) The condition to apply to a string value when querying for findings. Valid values include: `EQUALS` and `NOT_EQUALS`.
* `key` - (Required) The key of the map filter. For example, for `ResourceTags`, `Key` identifies the name of the tag. For `UserDefinedFields`, `Key` is the name of the field.
* `value` - (Required) The value for the key in the map filter. Filter values are case sensitive. For example, one of the values for a tag called `Department` might be `Security`. If you provide `security` as the filter value, then there is no match.

### Number Filter Argument reference

The number filter configuration block supports the following arguments:

~> **NOTE:** Only one of `eg`, `gte`, or `lte` must be specified.

* `eq` - (Optional) The equal-to condition to be applied to a single field when querying for findings, provided as a String.
* `gte` - (Optional) The greater-than-equal condition to be applied to a single field when querying for findings, provided as a String.
* `lte` - (Optional) The less-than-equal condition to be applied to a single field when querying for findings, provided as a String.

### String Filter Argument reference

The string filter configuration block supports the following arguments:

* `comparison` - (Required) The condition to apply to a string value when querying for findings. Valid values include: `EQUALS`, `PREFIX`, `NOT_EQUALS`, `PREFIX_NOT_EQUALS`.
* `value` - (Required) The string filter value. Filter values are case sensitive.

### Workflow Status Filter Argument reference

The workflow status filter configuration block supports the following arguments:

* `comparison` - (Required) The condition to apply to a string value when querying for findings. Valid values include: `EQUALS`, `PREFIX`, `NOT_EQUALS`, `PREFIX_NOT_EQUALS`.
* `value` - (Required) The string filter value. Valid values include: `NEW`, `NOTIFIED`, `SUPPRESSED`, and `RESOLVED`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ARN of the insight.
* `arn` - ARN of the insight.

## Import

Security Hub insights can be imported using the ARN, e.g.

```
$ terraform import aws_securityhub_insight.example arn:aws:securityhub:us-west-2:1234567890:insight/1234567890/custom/91299ed7-abd0-4e44-a858-d0b15e37141a
```
