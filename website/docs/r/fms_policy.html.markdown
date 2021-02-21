---
subcategory: "Firewall Manager (FMS)"
layout: "aws"
page_title: "AWS: aws_fms_policy"
description: |-
  Provides a resource to create an AWS Firewall Manager policy
---

# Resource: aws_fms_policy

Provides a resource to create an AWS Firewall Manager policy. You need to be using AWS organizations and have enabled the Firewall Manager administrator account.

## Example Usage

```hcl
resource "aws_fms_policy" "example" {
  name                  = "FMS-Policy-Example"
  exclude_resource_tags = false
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  security_service_policy_data {
    type = "WAF"

    managed_service_data = jsonencode({
      type = "WAF",
      ruleGroups = [{
        id = aws_wafregional_rule_group.example.id
        overrideAction = {
          type = "COUNT"
        }
      }]
      defaultAction = {
        type = "BLOCK"
      }
      overrideCustomerWebACLAssociation = false
    })
  }
}

resource "aws_wafregional_rule_group" "example" {
  metric_name = "WAFRuleGroupExample"
  name        = "WAF-Rule-Group-Example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, Forces new resource) The friendly name of the AWS Firewall Manager Policy.
* `delete_all_policy_resources` - (Optional) If true, the request will also perform a clean-up process. Defaults to `true`. More information can be found here [AWS Firewall Manager delete policy](https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_DeletePolicy.html)
* `exclude_map` - (Optional) A map of lists, with a single key named 'account' with a list of AWS Account IDs to exclude from this policy.
* `exclude_resource_tags` - (Required, Forces new resource) A boolean value, if true the tags that are specified in the `resource_tags` are not protected by this policy. If set to false and resource_tags are populated, resources that contain tags will be protected by this policy.
* `include_map` - (Optional) A map of lists, with a single key named 'account' with a list of AWS Account IDs to include for this policy.
* `remediation_enabled` - (Required) A boolean value, indicates if the policy should automatically applied to resources that already exist in the account.
* `resource_tags` - (Optional) A map of resource tags, that if present will filter protections on resources based on the exclude_resource_tags.
* `resource_type` - (Optional) A resource type to protect, valid values are: `AWS::ElasticLoadBalancingV2::LoadBalancer`, `AWS::ApiGateway::Stage`, `AWS::CloudFront::Distribution`, `AWS::EC2::Instance`, `AWS::EC2::NetworkInterface`, `AWS::EC2::SecurityGroup`. Conflicts with `resource_type_list`.
* `resource_type_list` - (Optional) A list of resource types to protect, valid values are: `AWS::ElasticLoadBalancingV2::LoadBalancer`, `AWS::ApiGateway::Stage`, `AWS::CloudFront::Distribution`, `AWS::EC2::Instance`, `AWS::EC2::NetworkInterface`, `AWS::EC2::SecurityGroup`, and `AWS::EC2::VPC`. Conflicts with `resource_type`.
* `security_service_policy_data` - (Required) The objects to include in Security Service Policy Data. Documented below.

## `exclude_map` Configuration Block

* `account` - (Required) A list of AWS Organization member Accounts that you want to exclude from this AWS FMS Policy.

## `include_map` Configuration Block

* `account` - (Required) A list of AWS Organization member Accounts that you want to include for this AWS FMS Policy.

## `security_service_policy_data` Configuration Block

* `managed_service_data` (Optional) Details about the service that are specific to the service type, in JSON format. For service type `SHIELD_ADVANCED`, this is an empty string. Examples depending on `type` can be found in the [AWS Firewall Manager SecurityServicePolicyData API Reference](https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_SecurityServicePolicyData.html).
* `type` - (Required, Forces new resource) The service that the policy is using to protect the resources. Valid values are `WAFV2`, `WAF`, `SHIELD_ADVANCED`, `SECURITY_GROUPS_COMMON`, `SECURITY_GROUPS_CONTENT_AUDIT`, and `SECURITY_GROUPS_USAGE_AUDIT`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The AWS account ID of the AWS Firewall Manager administrator account.
* `policy_update_token` - A unique identifier for each update to the policy.

## Import

Firewall Manager policies can be imported using the policy ID, e.g.

```
$ terraform import aws_fms_policy.example 5be49585-a7e3-4c49-dde1-a179fe4a619a
```
