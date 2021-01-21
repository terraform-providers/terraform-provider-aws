---
layout: "aws"
page_title: "Provider: AWS"
description: |-
  The Amazon Web Services (AWS) provider is used to interact with the many resources supported by AWS. The provider needs to be configured with the proper credentials before it can be used.
---

# AWS Provider

The Amazon Web Services (AWS) provider is used to interact with the
many resources supported by AWS. The provider needs to be configured
with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

Terraform 0.13 and later:

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

# Configure the AWS Provider
provider "aws" {
  region = "us-east-1"
}

# Create a VPC
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}
```

Terraform 0.12 and earlier:

```hcl
# Configure the AWS Provider
provider "aws" {
  version = "~> 3.0"
  region  = "us-east-1"
}

# Create a VPC
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}
```

## Authentication

The AWS provider offers a flexible means of providing credentials for
authentication. The following methods are supported, in this order, and
explained below:

- Static credentials
- Environment variables
- Shared credentials/configuration file
- CodeBuild, ECS, and EKS Roles
- EC2 Instance Metadata Service (IMDS and IMDSv2)

### Static Credentials

!> **Warning:** Hard-coded credentials are not recommended in any Terraform
configuration and risks secret leakage should this file ever be committed to a
public version control system.

Static credentials can be provided by adding an `access_key` and `secret_key`
in-line in the AWS provider block:

Usage:

```hcl
provider "aws" {
  region     = "us-west-2"
  access_key = "my-access-key"
  secret_key = "my-secret-key"
}
```

### Environment Variables

You can provide your credentials via the `AWS_ACCESS_KEY_ID` and
`AWS_SECRET_ACCESS_KEY`, environment variables, representing your AWS
Access Key and AWS Secret Key, respectively.  Note that setting your
AWS credentials using either these (or legacy) environment variables
will override the use of `AWS_SHARED_CREDENTIALS_FILE` and `AWS_PROFILE`.
The `AWS_DEFAULT_REGION` and `AWS_SESSION_TOKEN` environment variables
are also used, if applicable:

```hcl
provider "aws" {}
```

Usage:

```sh
$ export AWS_ACCESS_KEY_ID="anaccesskey"
$ export AWS_SECRET_ACCESS_KEY="asecretkey"
$ export AWS_DEFAULT_REGION="us-west-2"
$ terraform plan
```

### Shared Credentials File

You can use an [AWS credentials or configuration file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) to specify your credentials. The default location is `$HOME/.aws/credentials` on Linux and macOS, or `"%USERPROFILE%\.aws\credentials"` on Windows. You can optionally specify a different location in the Terraform configuration by providing the `shared_credentials_file` argument or using the `AWS_SHARED_CREDENTIALS_FILE` environment variable. This method also supports a `profile` configuration and matching `AWS_PROFILE` environment variable:

Usage:

```hcl
provider "aws" {
  region                  = "us-west-2"
  shared_credentials_file = "/Users/tf_user/.aws/creds"
  profile                 = "customprofile"
}
```

Please note that the [AWS Go SDK](https://aws.amazon.com/sdk-for-go/), the underlying authentication handler used by the Terraform AWS Provider, does not support all AWS CLI features, such as Single Sign On (SSO) configuration or credentials.

### CodeBuild, ECS, and EKS Roles

If you're running Terraform on CodeBuild or ECS and have configured an [IAM Task Role](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html), Terraform will use the container's Task Role. This support is based on the underlying `AWS_CONTAINER_CREDENTIALS_RELATIVE_URI` and `AWS_CONTAINER_CREDENTIALS_FULL_URI` environment variables being automatically set by those services or manually for advanced usage.

If you're running Terraform on EKS and have configured [IAM Roles for Service Accounts (IRSA)](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html), Terraform will use the pod's role. This support is based on the underlying `AWS_ROLE_ARN` and `AWS_WEB_IDENTITY_TOKEN_FILE` environment variables being automatically set by Kubernetes or manually for advanced usage.

### Custom User-Agent Information

By default, the underlying AWS client used by the Terraform AWS Provider creates requests with User-Agent headers including information about Terraform and AWS Go SDK versions. To provide additional information in the User-Agent headers, the `TF_APPEND_USER_AGENT` environment variable can be set and its value will be directly added to HTTP requests. e.g.

```sh
$ export TF_APPEND_USER_AGENT="JenkinsAgent/i-12345678 BuildID/1234 (Optional Extra Information)"
```

### EC2 Instance Metadata Service

If you're running Terraform from an EC2 instance with IAM Instance Profile
using IAM Role, Terraform will just ask
[the metadata API](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#instance-metadata-security-credentials)
endpoint for credentials.

This is a preferred approach over any other when running in EC2 as you can avoid
hard coding credentials. Instead these are leased on-the-fly by Terraform
which reduces the chance of leakage.

You can provide the custom metadata API endpoint via the `AWS_METADATA_URL` variable
which expects the endpoint URL, including the version, and defaults to `http://169.254.169.254:80/latest`.

### Assume Role

If provided with a role ARN, Terraform will attempt to assume this role
using the supplied credentials.

Usage:

```hcl
provider "aws" {
  assume_role {
    role_arn     = "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME"
    session_name = "SESSION_NAME"
    external_id  = "EXTERNAL_ID"
  }
}
```

## Argument Reference

In addition to [generic `provider` arguments](https://www.terraform.io/docs/configuration/providers.html)
(e.g. `alias` and `version`), the following arguments are supported in the AWS
 `provider` block:

* `access_key` - (Optional) This is the AWS access key. It must be provided, but
  it can also be sourced from the `AWS_ACCESS_KEY_ID` environment variable, or via
  a shared credentials file if `profile` is specified.

* `secret_key` - (Optional) This is the AWS secret key. It must be provided, but
  it can also be sourced from the `AWS_SECRET_ACCESS_KEY` environment variable, or
  via a shared credentials file if `profile` is specified.

* `region` - (Optional) This is the AWS region. It must be provided, but
  it can also be sourced from the `AWS_DEFAULT_REGION` environment variables, or
  via a shared credentials file if `profile` is specified.

* `profile` - (Optional) This is the AWS profile name as set in the shared credentials
  file.

* `assume_role` - (Optional) An `assume_role` block (documented below). Only one
  `assume_role` block may be in the configuration.

* `endpoints` - (Optional) Configuration block for customizing service endpoints. See the
[Custom Service Endpoints Guide](/docs/providers/aws/guides/custom-service-endpoints.html)
for more information about connecting to alternate AWS endpoints or AWS compatible solutions.

* `shared_credentials_file` = (Optional) This is the path to the shared credentials file.
  If this is not set and a profile is specified, `~/.aws/credentials` will be used.

* `token` - (Optional) Session token for validating temporary credentials. Typically provided after successful identity federation or Multi-Factor Authentication (MFA) login. With MFA login, this is the session token provided afterward, not the 6 digit MFA code used to get temporary credentials.  It can also be sourced from the `AWS_SESSION_TOKEN` environment variable.

* `max_retries` - (Optional) This is the maximum number of times an API
  call is retried, in the case where requests are being throttled or
  experiencing transient failures. The delay between the subsequent API
  calls increases exponentially. If omitted, the default value is `25`.

* `allowed_account_ids` - (Optional) List of allowed AWS
  account IDs to prevent you from mistakenly using an incorrect one (and
  potentially end up destroying a live environment). Conflicts with
  `forbidden_account_ids`.

* `forbidden_account_ids` - (Optional) List of forbidden
  AWS account IDs to prevent you from mistakenly using the wrong one (and
  potentially end up destroying a live environment). Conflicts with
  `allowed_account_ids`.

* `ignore_tags` - (Optional) Configuration block with resource tag settings to ignore across all resources handled by this provider (except any individual service tag resources such as `aws_ec2_tag`) for situations where external systems are managing certain resource tags. Arguments to the configuration block are described below in the `ignore_tags` Configuration Block section. See the [Terraform multiple provider instances documentation](https://www.terraform.io/docs/configuration/providers.html#alias-multiple-provider-configurations) for more information about additional provider configurations.

* `insecure` - (Optional) Explicitly allow the provider to
  perform "insecure" SSL requests. If omitted, the default value is `false`.

* `skip_credentials_validation` - (Optional) Skip the credentials
  validation via the STS API. Useful for AWS API implementations that do
  not have STS available or implemented.

* `skip_get_ec2_platforms` - (Optional) Skip getting the supported EC2
  platforms. Used by users that don't have ec2:DescribeAccountAttributes
  permissions.

* `skip_region_validation` - (Optional) Skip validation of provided region name.
  Useful for AWS-like implementations that use their own region names
  or to bypass the validation for regions that aren't publicly available yet.

* `skip_requesting_account_id` - (Optional) Skip requesting the account
  ID.  Useful for AWS API implementations that do not have the IAM, STS
  API, or metadata API.  When set to `true` and not determined previously,
  returns an empty account ID when manually constructing ARN attributes with
  the following:
    - [`aws_api_gateway_deployment` resource](/docs/providers/aws/r/api_gateway_deployment.html)
    - [`aws_api_gateway_rest_api` resource](/docs/providers/aws/r/api_gateway_rest_api.html)
    - [`aws_api_gateway_stage` resource](/docs/providers/aws/r/api_gateway_stage.html)
    - [`aws_apigatewayv2_api` resource](/docs/providers/aws/r/apigatewayv2_api.html)
    - [`aws_apigatewayv2_stage` resource](/docs/providers/aws/r/apigatewayv2_stage.html)
    - [`aws_athena_workgroup` resource](/docs/providers/aws/r/athena_workgroup.html)
    - [`aws_budgets_budget` resource](/docs/providers/aws/r/budgets_budget.html)
    - [`aws_cognito_identity_pool` resource](/docs/providers/aws/r/cognito_identity_pool.html)
    - [`aws_cognito_user_pools` data source](/docs/providers/aws/d/cognito_user_pools.html)
    - [`aws_default_network_acl` resource](/docs/providers/aws/r/default_network_acl.html)
    - [`aws_default_vpc_dhcp_options`](/docs/providers/aws/r/default_vpc_dhcp_options.html)
    - [`aws_dms_event_subscription` resource](/docs/providers/aws/r/dms_event_subscription.html)
    - [`aws_dms_replication_subnet_group` resource](/docs/providers/aws/r/dms_replication_subnet_group.html)
    - [`aws_dx_connection` resource](/docs/providers/aws/r/dx_connection.html)
    - [`aws_dx_hosted_private_virtual_interface_accepter` resource](/docs/providers/aws/r/dx_hosted_private_virtual_interface_accepter.html)
    - [`aws_dx_hosted_private_virtual_interface` resource](/docs/providers/aws/r/dx_hosted_private_virtual_interface.html)
    - [`aws_dx_hosted_public_virtual_interface_accepter` resource](/docs/providers/aws/r/dx_hosted_public_virtual_interface_accepter.html)
    - [`aws_dx_hosted_public_virtual_interface` resource](/docs/providers/aws/r/dx_hosted_public_virtual_interface.html)
    - [`aws_dx_hosted_transit_virtual_interface_accepter` resource](/docs/providers/aws/r/dx_hosted_transit_virtual_interface_accepter.html)
    - [`aws_dx_hosted_transit_virtual_interface` resource](/docs/providers/aws/r/dx_hosted_transit_virtual_interface.html)
    - [`aws_dx_lag` resource](/docs/providers/aws/r/dx_lag.html)
    - [`aws_dx_private_virtual_interface` resource](/docs/providers/aws/r/dx_private_virtual_interface.html)
    - [`aws_dx_public_virtual_interface` resource](/docs/providers/aws/r/dx_public_virtual_interface.html)
    - [`aws_dx_transit_virtual_interface` resource](/docs/providers/aws/r/dx_transit_virtual_interface.html)
    - [`aws_ebs_volume` data source](/docs/providers/aws/d/ebs_volume.html)
    - [`aws_ec2_capacity_reservation` resource](/docs/providers/aws/r/ec2_capacity_reservation.html)
    - [`aws_ec2_client_vpn_endpoint` resource](/docs/providers/aws/r/ec2_client_vpn_endpoint.html)
    - [`aws_ec2_traffic_mirror_session` resource](/docs/providers/aws/r/ec2_traffic_mirror_session.html)
    - [`aws_ec2_traffic_mirror_target` resource](/docs/providers/aws/r/ec2_traffic_mirror_target.html)
    - [`aws_ec2_transit_gateway_route_table` data source](/docs/providers/aws/d/ec2_transit_gateway_route_table.html)  
    - [`aws_ec2_transit_gateway_route_table` resource](/docs/providers/aws/r/ec2_transit_gateway_route_table.html)  
    - [`aws_ecs_capacity_provider` resource (import)](/docs/providers/aws/r/ecs_capacity_provider.html)
    - [`aws_ecs_cluster` resource (import)](/docs/providers/aws/r/ecs_cluster.html)
    - [`aws_ecs_service` resource (import)](/docs/providers/aws/r/ecs_service.html)
    - [`aws_customer_gateway` data source](/docs/providers/aws/d/customer_gateway.html)
    - [`aws_customer_gateway` resource](/docs/providers/aws/r/customer_gateway.html)  
    - [`aws_efs_access_point` data source](/docs/providers/aws/d/efs_access_point.html)
    - [`aws_efs_access_point` resource](/docs/providers/aws/r/efs_access_point.html)
    - [`aws_efs_file_system` data source](/docs/providers/aws/d/efs_file_system.html)
    - [`aws_efs_file_system` resource](/docs/providers/aws/r/efs_file_system.html)
    - [`aws_efs_mount_target` data source](/docs/providers/aws/d/efs_mount_target.html)
    - [`aws_efs_mount_target` resource](/docs/providers/aws/r/efs_mount_target.html)
    - [`aws_elasticache_cluster` data source](/docs/providers/aws/d/elasticache_cluster.html)
    - [`aws_elasticache_cluster` resource](/docs/providers/aws/r/elasticache_cluster.html)
    - [`aws_elb` data source](/docs/providers/aws/d/elb.html)
    - [`aws_elb` resource](/docs/providers/aws/r/elb.html)
    - [`aws_flow_log` resource](/docs/providers/aws/r/flow_log.html)
    - [`aws_glue_catalog_database` resource](/docs/providers/aws/r/glue_catalog_database.html)
    - [`aws_glue_catalog_table` resource](/docs/providers/aws/r/glue_catalog_table.html)
    - [`aws_glue_connection` resource](/docs/providers/aws/r/glue_connection.html)
    - [`aws_glue_crawler` resource](/docs/providers/aws/r/glue_crawler.html)
    - [`aws_glue_job` resource](/docs/providers/aws/r/glue_job.html)
    - [`aws_glue_ml_transform` resource](/docs/providers/aws/r/glue_ml_transform.html)
    - [`aws_glue_trigger` resource](/docs/providers/aws/r/glue_trigger.html)
    - [`aws_glue_user_defined_function` resource](/docs/providers/aws/r/glue_user_defined_function.html)
    - [`aws_glue_workflow` resource](/docs/providers/aws/r/glue_workflow.html)
    - [`aws_guardduty_detector` resource](/docs/providers/aws/r/guardduty_detector.html)
    - [`aws_guardduty_ipset` resource](/docs/providers/aws/r/guardduty_ipset.html)
    - [`aws_guardduty_threatintelset` resource](/docs/providers/aws/r/guardduty_threatintelset.html)
    - [`aws_instance` data source](/docs/providers/aws/d/instance.html)
    - [`aws_instance` resource](/docs/providers/aws/r/instance.html)
    - [`aws_internet_gateway` data source](/docs/providers/aws/d/internet_gateway.html)
    - [`aws_internet_gateway` resource](/docs/providers/aws/r/internet_gateway.html)
    - [`aws_key_pair` resource](/docs/providers/aws/r/key_pair.html)
    - [`aws_launch_template` data source](/docs/providers/aws/d/launch_template.html)
    - [`aws_launch_template` resource](/docs/providers/aws/r/launch_template.html)
    - [`aws_network_acl` resource](/docs/providers/aws/r/network_acl.html)
    - [`aws_placement_group` resource](/docs/providers/aws/r/placement_group.html)
    - [`aws_redshift_cluster` resource](/docs/providers/aws/r/redshift_cluster.html)
    - [`aws_redshift_event_subscription` resource](/docs/providers/aws/r/redshift_event_subscription.html)
    - [`aws_redshift_parameter_group` resource](/docs/providers/aws/r/redshift_parameter_group.html)
    - [`aws_redshift_snapshot_copy_grant` resource](/docs/providers/aws/r/redshift_snapshot_copy_grant.html)
    - [`aws_redshift_snapshot_schedule` resource](/docs/providers/aws/r/redshift_snapshot_schedule.html)
    - [`aws_redshift_subnet_group` resource](/docs/providers/aws/r/redshift_subnet_group.html)
    - [`aws_s3_account_public_access_block` resource](/docs/providers/aws/r/s3_account_public_access_block.html)
    - [`aws_ses_domain_identity` resource](/docs/providers/aws/r/ses_domain_identity.html)
    - [`aws_ses_domain_identity_verification` resource](/docs/providers/aws/r/ses_domain_identity_verification.html)
    - [`aws_ses_email_identity` resource](/docs/providers/aws/r/ses_email_identity.html)
    - [`aws_ses_receipt_filter` resource](/docs/providers/aws/r/ses_receipt_filter.html)
    - [`aws_ssm_document` data source](/docs/providers/aws/d/ssm_document.html)
    - [`aws_ssm_document` resource](/docs/providers/aws/r/ssm_document.html)
    - [`aws_ssm_parameter` data source](/docs/providers/aws/d/ssm_parameter.html)
    - [`aws_ssm_parameter` resource](/docs/providers/aws/r/ssm_parameter.html)
    - [`aws_vpc` data source](/docs/providers/aws/d/vpc.html)
    - [`aws_vpc` resource](/docs/providers/aws/r/vpc.html)
    - [`aws_vpc_dhcp_options` data source](/docs/providers/aws/d/vpc_dhcp_options.html)
    - [`aws_vpc_dhcp_options` resource](/docs/providers/aws/r/vpc_dhcp_options.html)
    - [`aws_vpc_endpoint` data source](/docs/providers/aws/d/vpc_endpoint.html)
    - [`aws_vpc_endpoint` resource](/docs/providers/aws/r/vpc_endpoint.html)
    - [`aws_vpc_endpoint_service` data source](/docs/providers/aws/d/vpc_endpoint_service.html)
    - [`aws_vpc_endpoint_service` resource](/docs/providers/aws/r/vpc_endpoint_service.html)
    - [`aws_vpn_connection` resource](/docs/providers/aws/r/vpn_connection.html)
    - [`aws_vpn_gateway` data source](/docs/providers/aws/d/vpn_gateway.html)
    - [`aws_vpn_gateway` resource](/docs/providers/aws/r/vpn_gateway.html)
    - [`aws_waf_geo_match_set` resource](/docs/providers/aws/r/waf_geo_match_set.html)
    - [`aws_waf_ipset` resource](/docs/providers/aws/r/waf_ipset.html)
    - [`aws_waf_rate_based_rule` resource](/docs/providers/aws/r/waf_rate_based_rule.html)
    - [`aws_waf_regex_match_set` resource](/docs/providers/aws/r/waf_regex_match_set.html)
    - [`aws_waf_regex_pattern_set` resource](/docs/providers/aws/r/waf_regex_pattern_set.html)
    - [`aws_wafregional_ipset` resource](/docs/providers/aws/r/wafregional_ipset.html)
    - [`aws_wafregional_rate_based_rule` resource](/docs/providers/aws/r/wafregional_rate_based_rule.html)
    - [`aws_wafregional_rule` resource](/docs/providers/aws/r/wafregional_rule.html)
    - [`aws_wafregional_rule_group` resource](/docs/providers/aws/r/wafregional_rule_group.html)
    - [`aws_wafregional_web_acl` resource](/docs/providers/aws/r/wafregional_web_acl.html)
    - [`aws_waf_rule` resource](/docs/providers/aws/r/waf_rule.html)
    - [`aws_waf_rule_group` resource](/docs/providers/aws/r/waf_rule_group.html)
    - [`aws_waf_size_constraint_set` resource](/docs/providers/aws/r/waf_size_constraint_set.html)
    - [`aws_waf_web_acl` resource](/docs/providers/aws/r/waf_web_acl.html)
    - [`aws_waf_xss_match_set` resource](/docs/providers/aws/r/waf_xss_match_set.html)

* `skip_metadata_api_check` - (Optional) Skip the AWS Metadata API
  check.  Useful for AWS API implementations that do not have a metadata
  API endpoint.  Setting to `true` prevents Terraform from authenticating
  via the Metadata API. You may need to use other authentication methods
  like static credentials, configuration variables, or environment
  variables.

* `s3_force_path_style` - (Optional) Set this to `true` to force the
  request to use path-style addressing, i.e.,
  `http://s3.amazonaws.com/BUCKET/KEY`. By default, the S3 client will use
  virtual hosted bucket addressing, `http://BUCKET.s3.amazonaws.com/KEY`,
  when possible. Specific to the Amazon S3 service.

### assume_role Configuration Block

The `assume_role` configuration block supports the following optional arguments:

* `duration_seconds` - (Optional) Number of seconds to restrict the assume role session duration.
* `external_id` - (Optional) External identifier to use when assuming the role.
* `policy` - (Optional) IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.
* `policy_arns` - (Optional) Set of Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.
* `role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM Role to assume.
* `session_name` - (Optional) Session name to use when assuming the role.
* `tags` - (Optional) Map of assume role session tags.
* `transitive_tag_keys` - (Optional) Set of assume role session tag keys to pass to any subsequent sessions.

### ignore_tags Configuration Block

Example:

```hcl
provider "aws" {
  ignore_tags {
    keys = ["TagKey1"]
  }
}
```

The `ignore_tags` configuration block supports the following arguments:

* `keys` - (Optional) List of exact resource tag keys to ignore across all resources handled by this provider. This configuration prevents Terraform from returning the tag in any `tags` attributes and displaying any configuration difference for the tag value. If any resource configuration still has this tag key configured in the `tags` argument, it will display a perpetual difference until the tag is removed from the argument or [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) is also used.
* `key_prefixes` - (Optional) List of resource tag key prefixes to ignore across all resources handled by this provider. This configuration prevents Terraform from returning any tag key matching the prefixes in any `tags` attributes and displaying any configuration difference for those tag values. If any resource configuration still has a tag matching one of the prefixes configured in the `tags` argument, it will display a perpetual difference until the tag is removed from the argument or [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) is also used.

## Getting the Account ID

If you use either `allowed_account_ids` or `forbidden_account_ids`,
Terraform uses several approaches to get the actual account ID
in order to compare it with allowed or forbidden IDs.

Approaches differ per authentication providers:

* EC2 instance w/ IAM Instance Profile - [Metadata API](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html)
    is always used. Introduced in Terraform `0.6.16`.
* All other providers (environment variable, shared credentials file, ...)
    will try three approaches in the following order
    * `iam:GetUser` - Typically useful for IAM Users. It also means
      that each user needs to be privileged to call `iam:GetUser` for themselves.
    * `sts:GetCallerIdentity` - _Should_ work for both IAM Users and federated IAM Roles,
      introduced in Terraform `0.6.16`.
    * `iam:ListRoles` - This is specifically useful for IdP-federated profiles
      which cannot use `iam:GetUser`. It also means that each federated user
      need to be _assuming_ an IAM role which allows `iam:ListRoles`.
      Used in Terraform `0.6.16+`.
      There used to be no better way to get account ID out of the API
      when using the federated account until `sts:GetCallerIdentity` was introduced.
