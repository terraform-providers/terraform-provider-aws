---
layout: "aws"
page_title: "AWS: aws_cognito_user_pool"
sidebar_current: "docs-aws-resource-cognito-user-pool"
description: |-
  Provides a Cognito User Pool resource.
---

# aws_cognito_user_pool

Provides a Cognito User Pool resource.

## Example Usage

### Basic configuration

```hcl
resource "aws_cognito_user_pool" "pool" {
  name = "mypool"
}
```

## Argument Reference

The following arguments are supported:

* `admin_create_user_config` (Optional) - The configuration for [AdminCreateUser](#admin-create-user-config) requests.
* `alias_attributes` - (Optional) Attributes supported as an alias for this user pool. Possible values: phone_number, email, or preferred_username. Conflicts with `username_attributes`.
* `auto_verified_attributes` - (Optional) The attributes to be auto-verified. Possible values: email, phone_number.
* `device_configuration` (Optional) - The configuration for the [user pool's device tracking](#device-configuration).
* `email_configuration` (Optional) - The [Email Configuration](#email-configuration).
* `name` - (Required) The name of the user pool.
* `email_verification_subject` - (Optional) A string representing the email verification subject. **NOTE:** - If `email_verification_subject` and `verification_message_template.email_subject` are specified and the values are different, either one is prioritized and updated.
* `email_verification_message` - (Optional) A string representing the email verification message. Must contain the `{####}` placeholder. **NOTE:** - If `email_verification_message` and `verification_message_template.email_message` are specified and the values are different, either one is prioritized and updated.
* `lambda_config` (Optional) - A container for the AWS [Lambda triggers](#lambda-configuration) associated with the user pool.
* `mfa_configuration` - (Optional, Default: OFF) Set to enable multi-factor authentication. Must be one of the following values (ON, OFF, OPTIONAL)
* `password_policy` (Optional) - A container for information about the [user pool password policy](#password-policy).
* `schema` (Optional) - A container with the [schema attributes](#schema-attributes) of a user pool. Maximum of 50 attributes.
* `sms_authentication_message` - (Optional) A string representing the SMS authentication message.
* `sms_configuration` (Optional) - The [SMS Configuration](#sms-configuration).
* `sms_verification_message` - (Optional) A string representing the SMS verification message.
* `tags` - (Optional) A mapping of tags to assign to the User Pool.
* `username_attributes` - (Optional) Specifies whether email addresses or phone numbers can be specified as usernames when a user signs up. Conflicts with `alias_attributes`.
* `verification_message_template` (Optional) - The [verification message templates](#verification-message-template) configuration.

#### Admin Create User Config

  * `allow_admin_create_user_only` (Optional) - Set to True if only the administrator is allowed to create user profiles. Set to False if users can sign themselves up via an app.
  * `invite_message_template` (Optional) - The [invite message template structure](#invite-message-template).
  * `unused_account_validity_days` (Optional) - The user account expiration limit, in days, after which the account is no longer usable.

##### Invite Message template

  * `email_message` (Optional) - The message template for email messages. Must contain `{username}` and `{####}` placeholders, for username and temporary password, respectively.
  * `email_subject` (Optional) - The subject line for email messages.
  * `sms_message` (Optional) - The message template for SMS messages. Must contain `{username}` and `{####}` placeholders, for username and temporary password, respectively.

#### Device Configuration

  * `challenge_required_on_new_device` (Optional) - Indicates whether a challenge is required on a new device. Only applicable to a new device.
  * `device_only_remembered_on_user_prompt` (Optional) - If true, a device is only remembered on user prompt.

#### Email Configuration

  * `reply_to_email_address` (Optional) - The REPLY-TO email address.
  * `source_arn` (Optional) - The ARN of the email source.

#### Lambda Configuration

  * `create_auth_challenge` (Optional) - The ARN of the lambda creating an authentication challenge.
  * `custom_message` (Optional) - A custom Message AWS Lambda trigger.
  * `define_auth_challenge` (Optional) - Defines the authentication challenge.
  * `post_authentication` (Optional) - A post-authentication AWS Lambda trigger.
  * `post_confirmation` (Optional) - A post-confirmation AWS Lambda trigger.
  * `pre_authentication` (Optional) - A pre-authentication AWS Lambda trigger.
  * `pre_sign_up` (Optional) - A pre-registration AWS Lambda trigger.
  * `pre_token_generation` (Optional) - Allow to customize identity token claims before token generation.
  * `user_migration` (Optional) - The user migration Lambda config type.
  * `verify_auth_challenge_response` (Optional) - Verifies the authentication challenge response.

#### Password Policy

  * `minimum_length` (Optional) - The minimum length of the password policy that you have set.
  * `require_lowercase` (Optional) - Whether you have required users to use at least one lowercase letter in their password.
  * `require_numbers` (Optional) - Whether you have required users to use at least one number in their password.
  * `require_symbols` (Optional) - Whether you have required users to use at least one symbol in their password.
  * `require_uppercase` (Optional) - Whether you have required users to use at least one uppercase letter in their password.

#### Schema Attributes

  * `attribute_data_type` (Required) - The attribute data type. Must be one of `Boolean`, `Number`, `String`, `DateTime`.
  * `developer_only_attribute` (Optional) - Specifies whether the attribute type is developer only.
  * `mutable` (Optional) - Specifies whether the attribute can be changed once it has been created.
  * `name` (Required) - The name of the attribute.
  * `number_attribute_constraints` (Optional) - Specifies the [constraints for an attribute of the number type](#number-attribute-constraints).
  * `required` (Optional) - Specifies whether a user pool attribute is required. If the attribute is required and the user does not provide a value, registration or sign-in will fail.
  * `string_attribute_constraints` (Optional) -Specifies the [constraints for an attribute of the string type](#string-attribute-constraints).

##### Number Attribute Constraints

  * `max_value` (Optional) - The maximum value of an attribute that is of the number data type.
  * `min_value` (Optional) - The minimum value of an attribute that is of the number data type.

##### String Attribute Constraints

  * `max_length` (Optional) - The maximum length of an attribute value of the string type.
  * `min_length` (Optional) - The minimum length of an attribute value of the string type.

#### SMS Configuration

  * `external_id` (Required) - The external ID used in IAM role trust relationships. For more information about using external IDs, see [How to Use an External ID When Granting Access to Your AWS Resources to a Third Party](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html).
  * `sns_caller_arn` (Required) - The ARN of the Amazon SNS caller. This is usually the IAM role that you've given Cognito permission to assume.

#### Verification Message Template

  * `default_email_option` (Optional) - The default email option. Must be either `CONFIRM_WITH_CODE` or `CONFIRM_WITH_LINK`. Defaults to `CONFIRM_WITH_CODE`.
  * `email_message` (Optional) - The email message template. Must contain the `{####}` placeholder. **NOTE:** - If `email_verification_message` and `verification_message_template.email_message` are specified and the values are different, either one is prioritized and updated.
  * `email_message_by_link` (Optional) - The email message template for sending a confirmation link to the user, it must contain the `{##VERIFY EMAIL##}` placeholder.
  * `email_subject` (Optional) - The subject line for the email message template. **NOTE:** - If `email_verification_subject` and `verification_message_template.email_subject` are specified and the values are different, either one is prioritized and updated.
  * `email_subject_by_link` (Optional) - The subject line for the email message template for sending a confirmation link to the user.
  * `sms_message` (Optional) - The SMS message template. Must contain the `{####}` placeholder.

## Attribute Reference

The following additional attributes are exported:

* `id` - The id of the user pool.
* `arn` - The ARN of the user pool.
* `creation_date` - The date the user pool was created.
* `last_modified_date` - The date the user pool was last modified.

## Import

Cognito User Pools can be imported using the `id`, e.g.

```
$ terraform import aws_cognito_user_pool.pool <id>
```
