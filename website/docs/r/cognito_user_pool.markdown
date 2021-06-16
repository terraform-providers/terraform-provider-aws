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

```terraform
resource "aws_cognito_user_pool" "pool" {
  name = "mypool"
}
```

### Enabling SMS and Software Token Multi-Factor Authentication

```terraform
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

```terraform
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

The following argument is required:

* `name` - (Required) Name of the user pool.

The following arguments are optional:

* `account_recovery_setting` - (Optional) Configuration block to define which verified available method a user can use to recover their forgotten password. [Detailed below](#account_recovery_setting).
* `admin_create_user_config` - (Optional) Configuration block for creating a new user profile. [Detailed below](#admin_create_user_config).
* `alias_attributes` - (Optional) Attributes supported as an alias for this user pool. Valid values: `phone_number`, `email`, or `preferred_username`. Conflicts with `username_attributes`.
* `auto_verified_attributes` - (Optional) Attributes to be auto-verified. Valid values: `email`, `phone_number`.
* `device_configuration` - (Optional) Configuration block for the user pool's device tracking. [Detailed below](#device_configuration).
* `email_configuration` - (Optional) Configuration block for configuring email. [Detailed below](#email_configuration).
* `email_verification_message` - (Optional) String representing the email verification message. Conflicts with `verification_message_template` configuration block `email_message` argument.
* `email_verification_subject` - (Optional) String representing the email verification subject. Conflicts with `verification_message_template` configuration block `email_subject` argument.
* `lambda_config` - (Optional) Configuration block for the AWS Lambda triggers associated with the user pool. [Detailed below](#lambda_configuration).
* `mfa_configuration` - (Optional) Multi-Factor Authentication (MFA) configuration for the User Pool. Defaults of `OFF`. Valid values are `OFF` (MFA Tokens are not required), `ON` (MFA is required for all users to sign in; requires at least one of `sms_configuration` or `software_token_mfa_configuration` to be configured), or `OPTIONAL` (MFA Will be required only for individual users who have MFA Enabled; requires at least one of `sms_configuration` or `software_token_mfa_configuration` to be configured).
* `password_policy` - (Optional) Configuration blocked for information about the user pool password policy. [Detailed below](#password_policy).
* `schema` - (Optional) Configuration block for the schema attributes of a user pool. [Detailed below](#schema). Schema attributes from the [standard attribute set](https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-settings-attributes.html#cognito-user-pools-standard-attributes) only need to be specified if they are different from the default configuration. Attributes can be added, but not modified or removed. Maximum of 50 attributes.
* `sms_authentication_message` - (Optional) String representing the SMS authentication message. The Message must contain the `{####}` placeholder, which will be replaced with the code.
* `sms_configuration` - (Optional) Configuration block for Short Message Service (SMS) settings. [Detailed below](#sms_configuration). These settings apply to SMS user verification and SMS Multi-Factor Authentication (MFA). Due to Cognito API restrictions, the SMS configuration cannot be removed without recreating the Cognito User Pool. For user data safety, this resource will ignore the removal of this configuration by disabling drift detection. To force resource recreation after this configuration has been applied, see the [`taint` command](https://www.terraform.io/docs/commands/taint.html).
* `sms_verification_message` - (Optional) String representing the SMS verification message. Conflicts with `verification_message_template` configuration block `sms_message` argument.
* `software_token_mfa_configuration` - (Optional) Configuration block for software token Mult-Factor Authentication (MFA) settings. [Detailed below](#software_token_mfa_configuration).
* `tags` - (Optional) Map of tags to assign to the User Pool. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user_pool_add_ons` - (Optional) Configuration block for user pool add-ons to enable user pool advanced security mode features. [Detailed below](#user_pool_add_ons).
* `username_attributes` - (Optional) Whether email addresses or phone numbers can be specified as usernames when a user signs up. Conflicts with `alias_attributes`.
* `username_configuration` - (Optional) Configuration block for username configuration. [Detailed below](#username_configuration).
* `verification_message_template` - (Optional) Configuration block for verification message templates. [Detailed below](#verification_message_template).

### account_recovery_setting

* `recovery_mechanism` - (Required) List of Account Recovery Options of the following structure:
    * `name` - (Required) Recovery method for a user. Can be of the following: `verified_email`, `verified_phone_number`, and `admin_only`.
    * `priority` - (Required) Positive integer specifying priority of a method with 1 being the highest priority.

### admin_create_user_config

* `allow_admin_create_user_only` - (Optional) Set to True if only the administrator is allowed to create user profiles. Set to False if users can sign themselves up via an app.
* `invite_message_template` - (Optional) Invite message template structure. [Detailed below](#invite_message_template).

#### invite_message_template

* `email_message` - (Optional) Message template for email messages. Must contain `{username}` and `{####}` placeholders, for username and temporary password, respectively.
* `email_subject` - (Optional) Subject line for email messages.
* `sms_message` - (Optional) Message template for SMS messages. Must contain `{username}` and `{####}` placeholders, for username and temporary password, respectively.

### device_configuration

* `challenge_required_on_new_device` - (Optional) Whether a challenge is required on a new device. Only applicable to a new device.
* `device_only_remembered_on_user_prompt` - (Optional) Whether a device is only remembered on user prompt. `false` equates to "Always" remember, `true` is "User Opt In," and not using a `device_configuration` block is "No."

### email_configuration

* `configuration_set` - (Optional) Email configuration set name from SES.
* `email_sending_account` - (Optional) Email delivery method to use. `COGNITO_DEFAULT` for the default email functionality built into Cognito or `DEVELOPER` to use your Amazon SES configuration.
* `from_email_address` - (Optional) Sender’s email address or sender’s display name with their email address (e.g. `john@example.com`, `John Smith <john@example.com>` or `\"John Smith Ph.D.\" <john@example.com>`). Escaped double quotes are required around display names that contain certain characters as specified in [RFC 5322](https://tools.ietf.org/html/rfc5322).
* `reply_to_email_address` - (Optional) REPLY-TO email address.
* `source_arn` - (Optional) ARN of the SES verified email identity to to use. Required if `email_sending_account` is set to `DEVELOPER`.

### lambda_config

* `create_auth_challenge` - (Optional) ARN of the lambda creating an authentication challenge.
* `custom_message` - (Optional) Custom Message AWS Lambda trigger.
* `define_auth_challenge` - (Optional) Defines the authentication challenge.
* `post_authentication` - (Optional) Post-authentication AWS Lambda trigger.
* `post_confirmation` - (Optional) Post-confirmation AWS Lambda trigger.
* `pre_authentication` - (Optional) Pre-authentication AWS Lambda trigger.
* `pre_sign_up` - (Optional) Pre-registration AWS Lambda trigger.
* `pre_token_generation` - (Optional) Allow to customize identity token claims before token generation.
* `user_migration` - (Optional) User migration Lambda config type.
* `verify_auth_challenge_response` - (Optional) Verifies the authentication challenge response.
* `kms_key_id` - (Optional) The Amazon Resource Name of Key Management Service Customer master keys. Amazon Cognito uses the key to encrypt codes and temporary passwords sent to CustomEmailSender and CustomSMSSender.
* `custom_email_sender` - (Optional) A custom email sender AWS Lambda trigger. See [custom_email_sender](#custom_email_sender) Below.
* `custom_sms_sender` - (Optional) A custom SMS sender AWS Lambda trigger. See [custom_sms_sender](#custom_sms_sender) Below.

#### custom_email_sender

* `lambda_arn` - (Required) The Lambda Amazon Resource Name of the Lambda function that Amazon Cognito triggers to send email notifications to users.
* `lambda_version` - (Required) The Lambda version represents the signature of the "request" attribute in the "event" information Amazon Cognito passes to your custom email Lambda function. The only supported value is `V1_0`.

#### custom_sms_sender

* `lambda_arn` - (Required) he Lambda Amazon Resource Name of the Lambda function that Amazon Cognito triggers to send SMS notifications to users.
* `lambda_version` - (Required) The Lambda version represents the signature of the "request" attribute in the "event" information Amazon Cognito passes to your custom SMS Lambda function. The only supported value is `V1_0`.

### password_policy

* `minimum_length` - (Optional) Minimum length of the password policy that you have set.
* `require_lowercase` - (Optional) Whether you have required users to use at least one lowercase letter in their password.
* `require_numbers` - (Optional) Whether you have required users to use at least one number in their password.
* `require_symbols` - (Optional) Whether you have required users to use at least one symbol in their password.
* `require_uppercase` - (Optional) Whether you have required users to use at least one uppercase letter in their password.
* `temporary_password_validity_days` - (Optional) In the password policy you have set, refers to the number of days a temporary password is valid. If the user does not sign-in during this time, their password will need to be reset by an administrator.

### schema

~> **NOTE:** When defining an `attribute_data_type` of `String` or `Number`, the respective attribute constraints configuration block (e.g `string_attribute_constraints` or `number_attribute_contraints`) is **required** to prevent recreation of the Terraform resource. This requirement is true for both standard (e.g. name, email) and custom schema attributes.

* `attribute_data_type` - (Required) Attribute data type. Must be one of `Boolean`, `Number`, `String`, `DateTime`.
* `developer_only_attribute` - (Optional) Whether the attribute type is developer only.
* `mutable` - (Optional) Whether the attribute can be changed once it has been created.
* `name` - (Required) Name of the attribute.
* `number_attribute_constraints` - (Required when `attribute_data_type` is `Number`) Configuration block for the constraints for an attribute of the number type. [Detailed below](#number_attribute_constraints).
* `required` - (Optional) Whether a user pool attribute is required. If the attribute is required and the user does not provide a value, registration or sign-in will fail.
* `string_attribute_constraints` - (Required when `attribute_data_type` is `String`) Constraints for an attribute of the string type. [Detailed below](#string_attribute_constraints).

#### schema: Defaults for Standard Attributes

The [standard attributes](https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-settings-attributes.html#cognito-user-pools-standard-attributes) have the following defaults. Note that attributes which match the default values are not stored in Terraform state when importing.

```terraform
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

#### number_attribute_constraints

* `max_value` - (Optional) Maximum value of an attribute that is of the number data type.
* `min_value` - (Optional) Minimum value of an attribute that is of the number data type.

#### string_attribute_constraints

* `max_length` - (Optional) Maximum length of an attribute value of the string type.
* `min_length` - (Optional) Minimum length of an attribute value of the string type.

### sms_configuration

* `external_id` - (Required) External ID used in IAM role trust relationships. For more information about using external IDs, see [How to Use an External ID When Granting Access to Your AWS Resources to a Third Party](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html).
* `sns_caller_arn` - (Required) ARN of the Amazon SNS caller. This is usually the IAM role that you've given Cognito permission to assume.

### software_token_mfa_configuration

The following arguments are required in the `software_token_mfa_configuration` configuration block:

* `enabled` - (Required) Boolean whether to enable software token Multi-Factor (MFA) tokens, such as Time-based One-Time Password (TOTP). To disable software token MFA When `sms_configuration` is not present, the `mfa_configuration` argument must be set to `OFF` and the `software_token_mfa_configuration` configuration block must be fully removed.

### user_pool_add_ons

* `advanced_security_mode` - (Required) Mode for advanced security, must be one of `OFF`, `AUDIT` or `ENFORCED`.

### username_configuration

* `case_sensitive` - (Required) Whether username case sensitivity will be applied for all users in the user pool through Cognito APIs.

### verification_message_template

* `default_email_option` - (Optional) Default email option. Must be either `CONFIRM_WITH_CODE` or `CONFIRM_WITH_LINK`. Defaults to `CONFIRM_WITH_CODE`.
* `email_message` - (Optional) Email message template. Must contain the `{####}` placeholder. Conflicts with `email_verification_message` argument.
* `email_message_by_link` - (Optional) Email message template for sending a confirmation link to the user, it must contain the `{##Click Here##}` placeholder.
* `email_subject` - (Optional) Subject line for the email message template. Conflicts with `email_verification_subject` argument.
* `email_subject_by_link` - (Optional) Subject line for the email message template for sending a confirmation link to the user.
* `sms_message` - (Optional) SMS message template. Must contain the `{####}` placeholder. Conflicts with `sms_verification_message` argument.
  

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the user pool.
* `creation_date` - Date the user pool was created.
* `custom_domain` - A custom domain name that you provide to Amazon Cognito. This parameter applies only if you use a custom domain to host the sign-up and sign-in pages for your application. For example: `auth.example.com`.
* `domain` - Holds the domain prefix if the user pool has a domain associated with it.
* `endpoint` - Endpoint name of the user pool. Example format: `cognito-idp.REGION.amazonaws.com/xxxx_yyyyy`
* `estimated_number_of_users` - A number estimating the size of the user pool.
* `id` - ID of the user pool.
* `last_modified_date` - Date the user pool was last modified.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Cognito User Pools can be imported using the `id`, e.g.

```
$ terraform import aws_cognito_user_pool.pool <id>
```
