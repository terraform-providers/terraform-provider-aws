---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_workteam"
description: |-
  Provides a Sagemaker Workteam resource.
---

# Resource: aws_sagemaker_workteam

Provides a Sagemaker Workteam resource.

## Example Usage

### Cognito Usage

```terraform
resource "aws_sagemaker_workteam" "example" {
  workteam_name  = "example"
  workforce_name = aws_sagemaker_workforce.example.id
  description    = "example"

  member_definition {
    cognito_member_definition {
      client_id  = aws_cognito_user_pool_client.example.id
      user_pool  = aws_cognito_user_pool_domain.example.user_pool_id
      user_group = aws_cognito_user_group.example.id
    }
  }
}
```

### Oidc Usage

```terraform
resource "aws_sagemaker_workteam" "example" {
  workteam_name  = "example"
  workforce_name = aws_sagemaker_workforce.example.id
  description    = "example"

  member_definition {
    oidc_member_definition {
      groups = ["example"]
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Required) A description of the work team.
* `workforce_name` - (Required) The name of the Workteam (must be unique).
* `workteam_name` - (Required) The name of the workforce.
* `member_definition` - (Required) A list of Member Definitions that contains objects that identify the workers that make up the work team. Workforces can be created using Amazon Cognito or your own OIDC Identity Provider (IdP). For private workforces created using Amazon Cognito use `cognito_member_definition`. For workforces created using your own OIDC identity provider (IdP) use `oidc_member_definition`. Do not provide input for both of these parameters in a single request. see [Member Definition](#member-definition) details below.
* `notification_configuration` - (Optional) Configures notification of workers regarding available or expiring work items. see [Notification Configuration](#notification-configuration) details below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Member Definition

* `cognito_member_definition` - (Optional) The Amazon Cognito user group that is part of the work team. See [Cognito Member Definition](#cognito-member-definition) details below.
* `oidc_member_definition` - (Optional) A list user groups that exist in your OIDC Identity Provider (IdP). One to ten groups can be used to create a single private work team. See [Cognito Member Definition](#oidc-member-definition) details below.

#### Cognito Member Definition

* `client_id` - (Required) An identifier for an application client. You must create the app client ID using Amazon Cognito.
* `user_pool` - (Required) An identifier for a user pool. The user pool must be in the same region as the service that you are calling.
* `user_group` - (Required) An identifier for a user group.

#### Oidc Member Definition

* `groups` - (Required) A list of comma separated strings that identifies user groups in your OIDC IdP. Each user group is made up of a group of private workers.

### Notification Configuration

* `notification_topic_arn` - (Required) The ARN for the SNS topic to which notifications should be published.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Workteam.
* `id` - The name of the Workteam.
* `subdomain` - The subdomain for your OIDC Identity Provider.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Sagemaker Workteams can be imported using the `workteam_name`, e.g.

```
$ terraform import aws_sagemaker_workteam.example example
```
