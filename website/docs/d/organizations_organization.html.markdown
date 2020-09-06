---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_organization"
description: |-
  Get information about the organization that the user's account belongs to
---

# Data Source: aws_organizations_organization

Get information about the organization that the user's account belongs to

## Example Usage

### List all account IDs for the organization

```hcl
data "aws_organizations_organization" "example" {}

output "account_ids" {
  value = data.aws_organizations_organization.example.accounts[*].id
}
```

### SNS topic that can be interacted by the organization only

```hcl
data "aws_organizations_organization" "example" {}

resource "aws_sns_topic" "sns_topic" {
  name = "my-sns-topic"
}

resource "aws_sns_topic_policy" "sns_topic_policy" {
  arn = aws_sns_topic.sns_topic.arn

  policy = data.aws_iam_policy_document.sns_topic_policy.json
}

data "aws_iam_policy_document" "sns_topic_policy" {
  statement {
    effect = "Allow"

    actions = [
      "SNS:Subscribe",
      "SNS:Publish",
    ]

    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalOrgID"

      values = [
        data.aws_organizations_organization.example.id,
      ]
    }

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    resources = [
      aws_sns_topic.sns_topic.arn,
    ]
  }
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the organization.
* `feature_set` - The FeatureSet of the organization.
* `id` - The ID of the organization.
* `master_account_arn` - The Amazon Resource Name (ARN) of the account that is designated as the master account for the organization.
* `master_account_email` - The email address that is associated with the AWS account that is designated as the master account for the organization.
* `master_account_id` - The unique identifier (ID) of the master account of an organization.

### Master Account Attributes Reference

If the account is the master account for the organization, the following attributes are also exported:

* `accounts` - List of organization accounts including the master account. For a list excluding the master account, see the `non_master_accounts` attribute. All elements have these attributes:
    * `arn` - ARN of the account
    * `email` - Email of the account
    * `id` - Identifier of the account
    * `name` - Name of the account
* `aws_service_access_principals` - A list of AWS service principal names that have integration enabled with your organization. Organization must have `feature_set` set to `ALL`. For additional information, see the [AWS Organizations User Guide](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_integrate_services.html).
* `enabled_policy_types` - A list of Organizations policy types that are enabled in the Organization Root. Organization must have `feature_set` set to `ALL`. For additional information about valid policy types (e.g. `SERVICE_CONTROL_POLICY`), see the [AWS Organizations API Reference](https://docs.aws.amazon.com/organizations/latest/APIReference/API_EnablePolicyType.html).
* `non_master_accounts` - List of organization accounts excluding the master account. For a list including the master account, see the `accounts` attribute. All elements have these attributes:
    * `arn` - ARN of the account
    * `email` - Email of the account
    * `id` - Identifier of the account
    * `name` - Name of the account
* `roots` - List of organization roots. All elements have these attributes:
    * `arn` - ARN of the root
    * `id` - Identifier of the root
    * `name` - Name of the root
    * `policy_types` - List of policy types enabled for this root. All elements have these attributes:
        * `name` - The name of the policy type
        * `status` - The status of the policy type as it relates to the associated root
