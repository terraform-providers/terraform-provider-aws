---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_provisioned_product"
description: |-
  Manages a Service Catalog Provisioned Product
---

# Resource: aws_servicecatalog_provisioned_product

This resource provisions and manages a Service Catalog provisioned product.

A provisioned product is a resourced instance of a product. For example, provisioning a product based on a CloudFormation template launches a CloudFormation stack and its underlying resources.

Like this resource, the `aws_servicecatalog_record` data source also provides information about a provisioned product. Although a Service Catalog record provides some overlapping information with this resource, a record is tied to a provisioned product event, such as provisioning, termination, and updating.

-> **Tip:** If you include conflicted keys as tags, AWS will report an error, "Parameter validation failed: Missing required parameter in Tags[N]:Value".

-> **Tip:** A "provisioning artifact" is also referred to as a "version." A "distributor" is also referred to as a "vendor."

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_provisioned_product" "example" {
  name  = "example"
  owner = [aws_security_group.example.id]
  type  = aws_subnet.main.id

  provisioning_artifact_parameters {
    template_url = "https://s3.amazonaws.com/cf-templates-ozkq9d3hgiq2-us-east-1/temp1.json"
  }

  tags = {
    foo = "bar"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) User-friendly name of the provisioned product.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.
* `ignore_errors` - (Optional) _Only applies to deleting._ If set to `true`, AWS Service Catalog stops managing the specified provisioned product even if it cannot delete the underlying resources. The default value is `false`.
* `notification_arns` - (Optional) Passed to CloudFormation. The SNS topic ARNs to which to publish stack-related events.
* `path_id` - (Optional) Path identifier of the product. This value is optional if the product has a default path, and required if the product has more than one path. To list the paths for a product, use `aws_servicecatalog_launch_paths`. When required, you must provide `path_id` or `path_name`, but not both.
* `path_name` - (Optional) Name of the path. You must provide `path_id` or `path_name`, but not both.
* `product_id` - (Optional) Product identifier. For example, `prod-abcdzk7xy33qa`. You must provide `product_id` or `product_name`, but not both.
* `product_name` - (Optional) Name of the product. You must provide `product_id` or `product_name`, but not both.
* `provisioning_artifact_id` - (Optional) Identifier of the provisioning artifact. For example, `pa-4abcdjnxjj6ne`. You must provide the `provisioning_artifact_id` or `provisioning_artifact_name`, but not both.
* `provisioning_artifact_name` - (Optional) Name of the provisioning artifact. You must provide the `provisioning_artifact_id` or `provisioning_artifact_name`, but not both.
* `provisioning_parameters` - (Optional) Configuration block with parameters specified by the administrator that are required for provisioning the product. See details below.
* `retain_physical_resources` - (Optional) _Only applies to deleting._ Whether to delete the Service Catalog provisioned product but leave the CloudFormation stack, stack set, or the underlying resources of the deleted provisioned product. The default value is `false`.
* `stack_set_provisioning_preferences` - (Optional) Configuration block with information about the provisioning preferences for a stack set. See details below.
* `tags` - (Optional) Tags to apply to the provisioned product. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### provisioning_parameters

The following arguments are supported:

* `key` - (Required) Parameter key.
* `use_previous_value` - (Optional) Whether to ignore `value` and keep the previous parameter value. Ignored when initially provisioning a product.
* `value` - (Optional) Parameter value.

### stack_set_provisioning_preferences

All of the `stack_set_provisioning_preferences` are only applicable to a `CFN_STACKSET` provisioned product type.

The following arguments are supported:

* `accounts` - (Optional) One or more AWS accounts that will have access to the provisioned product. The AWS accounts specified should be within the list of accounts in the STACKSET constraint. To get the list of accounts in the STACKSET constraint, use the `aws_servicecatalog_provisioning_parameters` data source. If no values are specified, the default value is all accounts from the STACKSET constraint.
* `failure_tolerance_count` - (Optional) Number of accounts, per region, for which this operation can fail before AWS Service Catalog stops the operation in that region. If the operation is stopped in a region, AWS Service Catalog doesn't attempt the operation in any subsequent regions. You must specify either `failure_tolerance_count` or `failure_tolerance_percentage`, but not both. The default value is 0 if no value is specified.
* `failure_tolerance_percentage` - (Optional) Percentage of accounts, per region, for which this stack operation can fail before AWS Service Catalog stops the operation in that region. If the operation is stopped in a region, AWS Service Catalog doesn't attempt the operation in any subsequent regions. When calculating the number of accounts based on the specified percentage, AWS Service Catalog rounds down to the next whole number. You must specify either `failure_tolerance_count` or `failure_tolerance_percentage`, but not both.
* `max_concurrency_count` - (Optional) Maximum number of accounts in which to perform this operation at one time. This is dependent on the value of `failure_tolerance_count`. `max_concurrency_count` is at most one more than the `failure_tolerance_count`. Note that this setting lets you specify the maximum for operations. For large deployments, under certain circumstances the actual number of accounts acted upon concurrently may be lower due to service throttling. You must specify either `max_concurrency_count` or `max_concurrency_percentage`, but not both.
* `max_concurrency_percentage` - (Optional) Maximum percentage of accounts in which to perform this operation at one time. When calculating the number of accounts based on the specified percentage, AWS Service Catalog rounds down to the next whole number. This is true except in cases where rounding down would result is zero. In this case, AWS Service Catalog sets the number as 1 instead. Note that this setting lets you specify the maximum for operations. For large deployments, under certain circumstances the actual number of accounts acted upon concurrently may be lower due to service throttling. You must specify either `max_concurrency_count` or `max_concurrency_percentage`, but not both.
* `regions` - (Optional) One or more AWS Regions where the provisioned product will be available. The specified regions should be within the list of regions from the STACKSET constraint. To get the list of regions in the STACKSET constraint, use the `aws_servicecatalog_provisioning_parameters` data source. If no values are specified, the default value is all regions from the STACKSET constraint.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the provisioned product.
* `cloudwatch_dashboard_names` - Set of CloudWatch dashboards that were created when provisioning the product.
* `created_time` - Time when the provisioned product was created.
* `id` - Provisioned Product ID.
* `last_provisioning_record_id` - Record identifier of the last request performed on this provisioned product of the following types: `ProvisionedProduct`, `UpdateProvisionedProduct`, `ExecuteProvisionedProductPlan`, `TerminateProvisionedProduct`.
* `last_record_id` - Record identifier of the last request performed on this provisioned product.
* `last_successful_provisioning_record_id` - Record identifier of the last successful request performed on this provisioned product of the following types: `ProvisionedProduct`, `UpdateProvisionedProduct`, `ExecuteProvisionedProductPlan`, `TerminateProvisionedProduct`.
* `launch_role_arn` - ARN of the launch role associated with the provisioned product.
* `status` - Current status of the provisioned product. See meanings below.
* `status_message` - Current status message of the provisioned product.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `type` - Type of provisioned product. Valid values are `CFN_STACK` and `CFN_STACKSET`.

### status Meanings

* `AVAILABLE` - Stable state, ready to perform any operation. The most recent operation succeeded and completed.
* `UNDER_CHANGE` - Transitive state. Operations performed might not have
valid results. Wait for an `AVAILABLE` status before performing operations.
* `TAINTED` - Stable state, ready to perform any operation. The stack has completed the requested operation but is not exactly what was requested. For example, a request to update to a new version failed and the stack rolled back to the current version.
* `ERROR` - An unexpected error occurred. The provisioned product exists but the stack is not running. For example, CloudFormation received a parameter value that was not valid and could not launch the stack.
* `PLAN_IN_PROGRESS` - Transitive state. The plan operations were performed to provision a new product, but resources have not yet been created. After reviewing the list of resources to be created, execute the plan. Wait for an `AVAILABLE` status before performing operations.

## Import

`aws_servicecatalog_provisioned_product` can be imported using the provisioned product ID, e.g.

```
$ terraform import aws_servicecatalog_provisioned_product.example pp-dnigbtea24ste
```
