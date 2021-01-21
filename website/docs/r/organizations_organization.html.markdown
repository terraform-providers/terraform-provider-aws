---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_organization"
description: |-
  Provides a resource to create an organization.
---

# Resource: aws_organizations_organization

Provides a resource to create an organization.

!> **WARNING:** When migrating from a `feature_set` of `CONSOLIDATED_BILLING` to `ALL`, the Organization account owner will received an email stating the following: "You started the process to enable all features for your AWS organization. As part of that process, all member accounts that joined your organization by invitation must approve the change. You don’t need approval from member accounts that you directly created from within your AWS organization." After all member accounts have accepted the invitation, the Organization account owner must then finalize the changes via the [AWS Console](https://console.aws.amazon.com/organizations/home#/organization/settings/migration-progress). Until these steps are performed, Terraform will perpetually show a difference, and the `DescribeOrganization` API will continue to show the `FeatureSet` as `CONSOLIDATED_BILLING`. See the [AWS Organizations documentation](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_org_support-all-features.html) for more information.

## Example Usage

```hcl
resource "aws_organizations_organization" "org" {
  aws_service_access_principals = [
    "cloudtrail.amazonaws.com",
    "config.amazonaws.com",
  ]

  feature_set = "ALL"
}
```

## Argument Reference

The following arguments are supported:

* `aws_service_access_principals` - (Optional) List of AWS service principal names for which you want to enable integration with your organization. This is typically in the form of a URL, such as service-abbreviation.amazonaws.com. Organization must have `feature_set` set to `ALL`. For additional information, see the [AWS Organizations User Guide](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_integrate_services.html).
* `enabled_policy_types` - (Optional) List of Organizations policy types to enable in the Organization Root. Organization must have `feature_set` set to `ALL`. For additional information about valid policy types (e.g. `BACKUP_POLICY`, `SERVICE_CONTROL_POLICY`, and `TAG_POLICY`), see the [AWS Organizations API Reference](https://docs.aws.amazon.com/organizations/latest/APIReference/API_EnablePolicyType.html).
* `feature_set` - (Optional) Specify "ALL" (default) or "CONSOLIDATED_BILLING".

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `accounts` - List of organization accounts including the master account. For a list excluding the master account, see the `non_master_accounts` attribute. All elements have these attributes:
    * `arn` - ARN of the account
    * `email` - Email of the account
    * `id` - Identifier of the account
    * `name` - Name of the account
    * `status` - Current status of the account
* `arn` - ARN of the organization
* `id` - Identifier of the organization
* `master_account_arn` - ARN of the master account
* `master_account_email` - Email address of the master account
* `master_account_id` - Identifier of the master account
* `non_master_accounts` - List of organization accounts excluding the master account. For a list including the master account, see the `accounts` attribute. All elements have these attributes:
    * `arn` - ARN of the account
    * `email` - Email of the account
    * `id` - Identifier of the account
    * `name` - Name of the account
    * `status` - Current status of the account
* `roots` - List of organization roots. All elements have these attributes:
    * `arn` - ARN of the root
    * `id` - Identifier of the root
    * `name` - Name of the root
    * `policy_types` - List of policy types enabled for this root. All elements have these attributes:
        * `name` - The name of the policy type
        * `status` - The status of the policy type as it relates to the associated root

## Import

The AWS organization can be imported by using the `id`, e.g.

```
$ terraform import aws_organizations_organization.my_org o-1234567
```
