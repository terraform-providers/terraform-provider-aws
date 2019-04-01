---
layout: "aws"
page_title: "AWS: aws_codepipeline_custom_action_type"
sidebar_current: "docs-aws-resource-codepipeline-webhook"
description: |-
  Provides a CodePipeline CustomActionType.
---

# aws_codepipeline_custom_action_type

Provides a CodeDeploy CustomActionType

## Example Usage

```hcl
resource "aws_codepipeline_custom_action_type" "example" {
  category = "Build"
  input_artifact_details {
    maximum_count = 1
    minimum_count = 0
  }
  output_artifact_details {
    maximum_count = 1
    minimum_count = 0
  }
  provider_name = "example"
  version = "1"
}
```

## Argument Reference

The following arguments are supported:

* `category` - (Required) The category of the custom action. Valid values: `Source`, `Build`, `Deploy`, `Test`, `Invoke`, `Approval`
* `configuration_properties` - (Optional) The configuration properties for the custom action. Max 10 items.

The `configuration_properties` object supports the following:

* `description` - (Optional) The description of the action configuration property.
* `key` - (Required) Whether the configuration property is a key.
* `name` - (Required) The name of the action configuration property.
* `queryable` - (Optional) Indicates that the property will be used in conjunction with PollForJobs.
* `required` - (Required) Whether the configuration property is a required value.
* `secret`- (Required) Whether the configuration property is secret.
* `type`- (Optional) The type of the configuration property. Valid values: `String`, `Number`, `Boolean`

* `input_artifact_details` - (Required) The details of the input artifact for the action.

The `input_artifact_details` object supports the following:

* `maximum_count` - (Required) The maximum number of artifacts allowed for the action type. Min: 0, Max: 5
* `minimum_count` - (Required) The minimum number of artifacts allowed for the action type. Min: 0, Max: 5

* `output_artifact_details` - (Required) The details of the output artifact of the action.

The `output_artifact_details` object supports the following:

* `maximum_count` - (Required) The maximum number of artifacts allowed for the action type. Min: 0, Max: 5
* `minimum_count` - (Required) The minimum number of artifacts allowed for the action type. Min: 0, Max: 5

* `provider_name` - (Required) The provider of the service used in the custom action
* `settings` - (Optional) The settings for an action type.

The `settings` object supports the following:

* `entity_url_template` - (Optional) The URL returned to the AWS CodePipeline console that provides a deep link to the resources of the external system.
* `execution_url_template` - (Optional) The URL returned to the AWS CodePipeline console that contains a link to the top-level landing page for the external system.
* `revision_url_template` - (Optional) The URL returned to the AWS CodePipeline console that contains a link to the page where customers can update or change the configuration of the external action.
* `third_party_configuration_url` - (Optional) The URL of a sign-up page where users can sign up for an external service and perform initial configuration of the action provided by that service.

* `version` - (Required) The version identifier of the custom action.

## Attribute Reference

The following arguments are exported:

* `id` - Composed of owner, category, provider and version. For example, `Custom:Build:terraform:1`
* `owner` - The creator of the action being called.

## Import

CodeDeploy CustomActionType can be imported using the `id`, e.g.

```
$ terraform import aws_codepipeline_custom_action_type.example Custom:Build:terraform:1
```
