---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_layer_version_permission"
description: |-
  Provides a Lambda Layer Version Permission resource. It allows you to share you own Lambda Layers to another account by account ID, to all accounts in AWS organization or even to all AWS accounts.
---

# Resource: aws_lambda_layer_version_permission

Provides a Lambda Layer Version Permission resource. It allows you to share you own Lambda Layers to another account by account ID, to all accounts in AWS organization or even to all AWS accounts.

For information about Lambda Layer Permissions and how to use them, see [Using Resource-based Policies for AWS Lambda][1]

## Example Usage

```hcl
resource "aws_lambda_layer_version_permission" "lambda_layer_permission" {
  layer_arn     = "arn:aws:lambda:us-west-2:123456654321:layer:test_layer1"
  layer_version = 1
  principal     = "111111111111"
  action        = "lambda:GetLayerVersion"
  statement_id  = "dev-account"
}
```

## Argument Reference

* `layer_arn` (Required) ARN of the Lambda Layer, which you want to grant access to.
* `layer_version` (Required) Version of Lambda Layer, which you want to grant access to. Note: permissions only apply to a single version of a layer.
* `principal` - (Required) AWS account ID which should be able to use your Lambda Layer. `*` can be used here, if you want to share your Lambda Layer widely.
* `organization_id` - (Optional) An identifier of AWS Organization, which should be able to use your Lambda Layer. `principal` should be equal to `*` if `organization_id` provided.
* `action` - (Required) Action, which will be allowed. `lambda:GetLayerVersion` value is suggested by AWS documantation.
* `statement_id` - (Required) The name of Lambda Layer Permission, for example `dev-account` - human readable note about what is this permission for.


## Attributes Reference

* `layer_arn` - The Amazon Resource Name (ARN) of the Lambda Layer without version.
* `layer_version` - The version of Lambda Layer.
* `principal` - The principal which was granted access to your Lambda Layer.
* `organization_id` - The AWS Organization which was granted access to your Lambda Layer.
* `action` - Action, which is allowed to principal.
* `statement_id` - Human readable name of Lambda Layer Permission.
* `revision_id` - Identifier of Lambda Layer Permission.
* `policy` - Full Lambda Layer Permission policy.

[1]: https://docs.aws.amazon.com/lambda/latest/dg/access-control-resource-based.html#permissions-resource-xaccountlayer

## Import

Lambda Layer Permissions can be imported using `arn`.

```
$ terraform import \
    aws_lambda_layer_version_permission.lambda_layer_permission \
    arn:aws:lambda:_REGION_:_ACCOUNT_ID_:layer:_LAYER_NAME_:_LAYER_VERSION_
```
