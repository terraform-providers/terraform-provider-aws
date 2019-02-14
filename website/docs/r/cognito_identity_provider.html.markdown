---
layout: "aws"
page_title: "AWS: aws_cognito_identity_provider"
side_bar_current: "docs-aws-resource-cognito-identity-provider"
description: |-
  Provides a Cognito User Identity Provider resource.
---

# aws_cognito_identity_provider

Provides a Cognito User Identity Provider resource.

## Example Usage

```hcl
resource "aws_cognito_user_pool" "example" {
  name                     = "example-pool"
  auto_verified_attributes = ["email"]
}

resource "aws_cognito_identity_provider" "example_provider" {
  user_pool_id  = "${aws_cognito_user_pool.example.id}"
  provider_name = "Google"
  provider_type = "Google"

  provider_details = {
    authorize_scopes = "email"
    client_id        = "your client_id"
    client_secret    = "your client_secret"
  }

  attribute_mapping = {
    email    = "email"
    username = "sub"
  }
}
```

## Argument Reference

The following arguments are supported:

* `user_pool_id` (Required) - The user pool id
* `provider_name` (Required) - The provider name
* `provider_type` (Required) - The provider type.  [See AWS API for valid values](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateIdentityProvider.html#CognitoUserPools-CreateIdentityProvider-request-ProviderType)
* `attribute_mapping` (Optional) - The map of attribute mapping of user pool attributes. [AttributeMapping in AWS API documentation](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateIdentityProvider.html#CognitoUserPools-CreateIdentityProvider-request-AttributeMapping)
* `idp_identifiers` (Optional) - The list of identity providers.
* `provider_details` (Optional) - The map of identity details, such as access token

## Import

`aws_cognito_identity_provider` resources can be imported using their User Pool ID and Provider Name, e.g.

```
$ terraform import aws_cognito_identity_provider.example xxx_yyyyy:example
```
