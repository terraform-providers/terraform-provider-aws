---
subcategory: "Cognito"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool"
description: |-
  Provides a Cognito User Pool resource.
---

# Resource: aws_cognito_user_pool

Provides a Cognito User Pool resource.

## Example Usage

### Basic configuration

```hcl
resource "aws_cognito_user_pool" "pool" {
  name = "mypool"
}
```

### Enabling SMS and Software Token Multi-Factor Authentication

```hcl
resource "aws_cognito_user_pool" "example" {
  # ... other configuration ...

  mfa_configuration          = "ON"
  sms_authentication_message = "Your code is {####}"

  sms_configuration {
    external_id    = "example"
    sns_caller_arn = aws_iam_role.example.arn
  }

  software_token_mfa_configuration {
    enabled = true
  }
}
```

### Using Account Recovery Setting

```hcl
resource "aws_cognito_user_pool" "test" {
  name = "mypool"

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }

    recovery_mechanism {
      name     = "verified_phone_number"
      priority = 2
    }
  }
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
* `email_verification_subject` - (Optional) A string representing the email verification subject. Conflicts with `verification_message_template` configuration block `email_subject` argument.
* `email_verification_message` - (Optional) A string representing the email verification message. Conflicts with `verification_message_template` configuration block `email_message` argument.
* `lambda_config` (Optional) - A container for the AWS [Lambda triggers](#lambda-configuration) associated with the user pool.
* `mfa_configuration` - (Optional) Multi-Factor Authentication (MFA) configuration for the User Pool. Defaults of `OFF`. Valid values:
    * `OFF` - MFA tokens are not required.
    * `ON` - MFA is required for all users to sign in. Requires at least one of `sms_configuration` or `software_token_mfa_configuration` to be configured.
    * `OPTIONAL` - MFA will be required only for individual users who have MFA enabled. Requires at least one of `sms_configuration` or `software_token_mfa_configuration` to be configured.
* `password_policy` (Optional) - A container for information about the [user pool password policy](#password-policy).
* `schema` (Optional) - A container with the [schema attributes](#schema-attributes) of a user pool. Schema attributes from the [standard attribute set](https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-settings-attributes.html#cognito-user-pools-standard-attributes) only need to be specified if they are different from the default configuration. Maximum of 50 attributes.
* `sms_authentication_message` - (Optional) A string representing the SMS authentication message. The message must contain the `{####}` placeholder, which will be replaced with the code.
* `sms_configuration` (Optional) - Configuration block for Short Message Service (SMS) settings. Detailed below. These settings apply to SMS user verification and SMS Multi-Factor Authentication (MFA). Due to Cognito API restrictions, the SMS configuration cannot be removed without recreating the Cognito User Pool. For user data safety, this resource will ignore the removal of this configuration by disabling drift detection. To force resource recreation after this configuration has been applied, see the [`taint` command](https://www.terraform.io/docs/commands/taint.html).
* `sms_verification_message` - (Optional) A string representing the SMS verification message. Conflicts with `verification_message_template` configuration block `sms_message` argument.
* `software_token_mfa_configuration` - (Optional) Configuration block for software token Mult-Factor Authentication (MFA) settings. Detailed below.
* `tags` - (Optional) A map of tags to assign to the User Pool.
* `username_attributes` - (Optional) Specifies whether email addresses or phone numbers can be specified as usernames when a user signs up. Conflicts with `alias_attributes`.
* `username_configuration` - (Optional) The [Username Configuration](#username-configuration).
* `user_pool_add_ons` - (Optional) Configuration block for [user pool add-ons](#user-pool-add-ons) to enable user pool advanced security mode features.
* `verification_message_template` (Optional) - The [verification message templates](#verification-message-template) configuration.
* `account_recovery_setting` (Optional) - The [account_recovery_setting](#account-recovery-setting) configuration.

#### Admin Create User Config

* `allow_admin_create_user_only` (Optional) - Set to True if only the administrator is allowed to create user profiles. Set to False if users can sign themselves up via an app.
* `invite_message_template` (Optional) - The [invite message template structure](#invite-message-template).

##### Invite Message template

* `email_message` (Optional) - The message template for email messages. Must contain `{username}` and `{####}` placeholders, for username and temporary password, respectively.
* `email_subject` (Optional) - The subject line for email messages.
* `sms_message` (Optional) - The message template for SMS messages. Must contain `{username}` and `{####}` placeholders, for username and temporary password, respectively.

#### Device Configuration

* `challenge_required_on_new_device` (Optional) - Indicates whether a challenge is required on a new device. Only applicable to a new device.
* `device_only_remembered_on_user_prompt` (Optional) - If true, a device is only remembered on user prompt.

#### Email Configuration

* `reply_to_email_address` (Optional) - The REPLY-TO email address.
* `source_arn` (Optional) - The ARN of the SES verified email identity to to use. Required if `email_sending_account` is set to `DEVELOPER`.
* `from_email_address` (Optional) - Sender’s email address or sender’s display name with their email address (e.g. `john@example.com`, `John Smith <john@example.com>` or `\"John Smith Ph.D.\" <john@example.com>`). Escaped double quotes are required around display names that contain certain characters as specified in [RFC 5322](https://tools.ietf.org/html/rfc5322).
* `email_sending_account` (Optional) - The email delivery method to use. `COGNITO_DEFAULT` for the default email functionality built into Cognito or `DEVELOPER` to use your Amazon SES configuration.

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
* `temporary_password_validity_days` (Optional) - In the password policy you have set, refers to the number of days a temporary password is valid. If the user does not sign-in during this time, their password will need to be reset by an administrator.

#### Schema Attributes

~> **NOTE:** When defining an `attribute_data_type` of `String` or `Number`, the respective attribute constraints configuration block (e.g `string_attribute_constraints` or `number_attribute_contraints`) is required to prevent recreation of the Terraform resource. This requirement is true for both standard (e.g. name, email) and custom schema attributes.

* `attribute_data_type` (Required) - The attribute data type. Must be one of `Boolean`, `Number`, `String`, `DateTime`.
* `developer_only_attribute` (Optional) - Specifies whether the attribute type is developer only.
* `mutable` (Optional) - Specifies whether the attribute can be changed once it has been created.
* `name` (Required) - The name of the attribute.
* `number_attribute_constraints` (Optional) - Specifies the [constraints for an attribute of the number type](#number-attribute-constraints).
* `required` (Optional) - Specifies whether a user pool attribute is required. If the attribute is required and the user does not provide a value, registration or sign-in will fail.
* `string_attribute_constraints` (Optional) -Specifies the [constraints for an attribute of the string type](#string-attribute-constraints).

##### Defaults for Standard Attributes

The [standard attributes](https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-settings-attributes.html#cognito-user-pools-standard-attributes) have the following defaults. Note that attributes which match the default values are not stored in Terraform state when importing.

```hcl
resource "aws_cognito_user_pool" "example" {
  # ... other configuration ...

  schema {
    name                     = "<name>"
    attribute_data_type      = "<appropriate type>"
    developer_only_attribute = false
    mutable                  = true  # false for "sub"
    required                 = false # true for "sub"
    string_attribute_constraints {   # if it is a string
      min_length = 0                 # 10 for "birthdate"
      max_length = 2048              # 10 for "birthdate"
    }
  }
}
```

##### Number Attribute Constraints

* `max_value` (Optional) - The maximum value of an attribute that is of the number data type.
* `min_value` (Optional) - The minimum value of an attribute that is of the number data type.

##### String Attribute Constraints

* `max_length` (Optional) - The maximum length of an attribute value of the string type.
* `min_length` (Optional) - The minimum length of an attribute value of the string type.

#### SMS Configuration

* `external_id` (Required) - The external ID used in IAM role trust relationships. For more information about using external IDs, see [How to Use an External ID When Granting Access to Your AWS Resources to a Third Party](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html).
* `sns_caller_arn` (Required) - The ARN of the Amazon SNS caller. This is usually the IAM role that you've given Cognito permission to assume.

### Software Token MFA Configuration

The following arguments are required in the `software_token_mfa_configuration` configuration block:

* `enabled` - (Required) Boolean whether to enable software token Multi-Factor (MFA) tokens, such as Time-based One-Time Password (TOTP). To disable software token MFA when `sms_configuration` is not present, the `mfa_configuration` argument must be set to `OFF` and the `software_token_mfa_configuration` configuration block must be fully removed.

#### Username Configuration

* `case_sensitive` (Required) - Specifies whether username case sensitivity will be applied for all users in the user pool through Cognito APIs.

#### User Pool Add-ons

* `advanced_security_mode` (Required) - The mode for advanced security, must be one of `OFF`, `AUDIT` or `ENFORCED`.

#### Verification Message Template

* `default_email_option` (Optional) - The default email option. Must be either `CONFIRM_WITH_CODE` or `CONFIRM_WITH_LINK`. Defaults to `CONFIRM_WITH_CODE`.
* `email_message` (Optional) - The email message template. Must contain the `{####}` placeholder. Conflicts with `email_verification_message` argument.
* `email_message_by_link` (Optional) - The email message template for sending a confirmation link to the user, it must contain the `{##Click Here##}` placeholder.
* `email_subject` (Optional) - The subject line for the email message template. Conflicts with `email_verification_subject` argument.
* `email_subject_by_link` (Optional) - The subject line for the email message template for sending a confirmation link to the user.
* `sms_message` (Optional) - The SMS message template. Must contain the `{####}` placeholder. Conflicts with `sms_verification_message` argument.
  
### Account Recovery Setting

* `recovery_mechanism` (Required) - The list of Account Recovery Options of the following structure:
    * `name` (Required) - Specifies the recovery method for a user. Can be of the following: `verified_email`, `verified_phone_number`, and `admin_only`.
    * `priority` (Required) - A positive integer specifying priority of a method with 1 being the highest priority.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The id of the user pool.
* `arn` - The ARN of the user pool.
* `endpoint` - The endpoint name of the user pool. Example format: cognito-idp.REGION.amazonaws.com/xxxx_yyyyy
* `creation_date` - The date the user pool was created.
* `last_modified_date` - The date the user pool was last modified.

## Import

Cognito User Pools can be imported using the `id`, e.g.

```
$ terraform import aws_cognito_user_pool.pool <id>
```
