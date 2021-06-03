---
subcategory: "AppFlow"
layout: "aws"
page_title: "AWS: aws_appflow_connector_profile"
description: |-
  Provides an AppFlow connector profile resource.
---

# Resource: aws_appflow_connector_profile

Creates an AWS AppFlow connector profile.

For information about AppFlow flows, see the
[Amazon AppFlow API Reference][1]. For specific information about creating
AppFlow connector profile, see the [CreateConnectorProfile][2] page in the Amazon
AppFlow API Reference.

## Example Usage

### Amplitude

```terraform
resource "aws_appflow_connector_profile" "amplitude" {
  connection_mode = "Public"
  connector_profile_config {
    connector_profile_credentials {
      amplitude {
        api_key    = "0123456789abcdef0123456789abcdef"
        secret_key = "0123456789abcdef0123456789abcdef"
      }
    }
    connector_profile_properties {
      amplitude {
      }
    }
  }
  connector_profile_name = "amplitude-connector"
  connector_type         = "Amplitude"
}
```

### Zendesk

```terraform
data "aws_region" "current" {}

resource "aws_appflow_connector_profile" "zendesk" {
  connection_mode = "Public"
  connector_profile_config {
    connector_profile_credentials {
      zendesk {
        client_id     = "zendesk-client-id"
        client_secret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
        access_token  = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
        oauth_request {
          redirect_uri = "https://${data.aws_region.current.name}.console.aws.amazon.com/appflow/oauth"
        }
      }
    }
    connector_profile_properties {
      zendesk {
        instance_url = "https://subdomain.zendesk.com"
      }
    }
  }
  connector_profile_name = "zendesk-connector"
  connector_type         = "Zendesk"
}
```

## Argument Reference

The AppFlow connector profile argument layout is a complex structure.

### Top-Level Arguments

* `connection_mode` (Required) - Indicates the connection mode and specifies whether it is public or private. Private flows use AWS PrivateLink to route data over AWS infrastructure without exposing it to the public internet. One of: `Public`, `Private`.

* `connector_profile_config` (Required) - Defines the connector-specific [configuration and credentials](#connector-profile-config-arguments).

* `connector_profile_name ` (Required) - The name of the connector profile. The name is unique for each ConnectorProfile in your AWS account.

* `connector_type` (Required) - The type of connector. One of: `Amplitude`, `Datadog`, `Dynatrace`, `Googleanalytics`, `Honeycode`, `Infornexus`, `Marketo`, `Redshift`, `Salesforce`, `Servicenow`, `Singular`, `Slack`, `Snowflake`, `Trendmicro`, `Veeva`, `Zendesk`.

* `kms_arn` (Optional) - The ARN (Amazon Resource Name) of the Key Management Service (KMS) key you provide for encryption. This is required if you do not want to use the Amazon AppFlow-managed KMS key. If you don't provide anything here, Amazon AppFlow uses the Amazon AppFlow-managed KMS key.

#### Connector Profile Config Arguments

* `connector_profile_credentials` (Required) - The connector-specific [credentials](#connector-profile-credentials-arguments) required by each connector.

* `connector_profile_properties` (Required) - The connector-specific [properties](#connector-profile-properties-arguments) of the profile configuration.


##### Connector Profile Credentials Arguments

* `amplitude` (Optional) - The connector-specific [credentials](#amplitude-connector-profile-credentials-arguments) required when using Amplitude.

* `datadog` (Optional) - The connector-specific [credentials](#datadog-connector-profile-credentials-arguments) required when using Datadog.

* `dynatrace` (Optional) - The connector-specific [credentials](#dynatrace-connector-profile-credentials-arguments) required when using Dynatrace.

* `google_analytics` (Optional) - The connector-specific [credentials](#google-analytics-connector-profile-credentials-arguments) required when using Google Analytics.

* `honeycode` (Optional) - The connector-specific [credentials](#honeycode-connector-profile-credentials-arguments) required when using Amazon Honeycode.

* `infor_nexus` (Optional) - The connector-specific [credentials](#infor-nexus-connector-profile-credentials-arguments) required when using Infor Nexus.

* `marketo` (Optional) - The connector-specific [credentials](#marketo-connector-profile-credentials-arguments) required when using Marketo.

* `redshift` (Optional) - The connector-specific [credentials](#redshift-connector-profile-credentials-arguments) required when using Amazon Redshift.

* `salesforce` (Optional) - The connector-specific [credentials](#salesforce-connector-profile-credentials-arguments) required when using Salesforce.

* `service_now` (Optional) - The connector-specific [credentials](#servicenow-connector-profile-credentials-arguments) required when using ServiceNow.

* `singular` (Optional) - The connector-specific [credentials](#singular-connector-profile-credentials-arguments) required when using Singular.

* `slack` (Optional) - The connector-specific [credentials](#slack-connector-profile-credentials-arguments) required when using Slack.

* `snowflake` (Optional) - The connector-specific [credentials](#snowflake-connector-profile-credentials-arguments) required when using Snowflake.

* `trendmicro` (Optional) - The connector-specific [credentials](#trendmicro-connector-profile-credentials-arguments) required when using Trend Micro.

* `veeva` (Optional) - The connector-specific [credentials](#veeva-connector-profile-credentials-arguments) required when using Veeva.

* `zendesk` (Optional) - The connector-specific [credentials](#zendesk-connector-profile-credentials-arguments) required when using Zendesk.

##### Amplitude Connector Profile Credentials Arguments

* `api_key` (Required) - A unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.

* `secret_key` (Required) - The Secret Access Key portion of the credentials.

##### Datadog Connector Profile Credentials Arguments

* `api_key` (Required) - A unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.

* `application_key` (Required) - Application keys, in conjunction with your API key, give you full access to Datadog’s programmatic API. Application keys are associated with the user account that created them. The application key is used to log all requests made to the API.

##### Dynatrace Connector Profile Credentials Arguments

* `api_token` (Required) - The API tokens used by Dynatrace API to authenticate various API calls.

##### Google Analytics Connector Profile Credentials Arguments

* `access_token` (Optional) - The credentials used to access protected Google Analytics resources.

* `client_id` (Required) - The identifier for the desired client.

* `client_secret` (Required) - The client secret used by the OAuth client to authenticate to the authorization server.

* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint.

* `refresh_token` (Optional) - The credentials used to acquire new access tokens. This is required only for OAuth2 access tokens, and is not required for OAuth1 access tokens.

##### Honeycode Connector Profile Credentials Arguments

* `access_token` (Optional) - The credentials used to access protected Amazon Honeycode resources.

* `oauth_request` (Optional) - Used by select connectors for which the OAuth workflow is supported, such as Salesforce, Google Analytics, Marketo, Zendesk, and Slack.

* `refresh_token` (Optional) - The credentials used to acquire new access tokens.

##### Infor Nexus Connector Profile Credentials Arguments

* `access_key_id` (Required) - The Access Key portion of the credentials.

* `datakey` (Required) - The encryption keys used to encrypt data.

* `secret_access_key` (Required) - The secret key used to sign requests.

* `user_id` (Required) - The identifier for the user.

##### Marketo Connector Profile Credentials Arguments

* `access_token` (Optional) - The credentials used to access protected Marketo resources.

* `client_id` (Required) - The identifier for the desired client.

* `client_secret` (Required) - The client secret used by the OAuth client to authenticate to the authorization server.

* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint.

##### Redshift Connector Profile Credentials Arguments

* `password` (Required) - The password that corresponds to the user name.

* `username` (Required) - The name of the user.

##### Salesforce Connector Profile Credentials Arguments

* `access_token` (Optional) - The credentials used to access protected Salesforce resources.

* `client_credentials_arn` (Optional) - The secret manager ARN, which contains the client ID and client secret of the connected app.

* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint.

* `refresh_token` (Optional) - The credentials used to acquire new access tokens.

##### ServiceNow Connector Profile Credentials Arguments

* `password` (Required) - The password that corresponds to the user name.

* `username` (Required) - The name of the user.

##### Singular Connector Profile Credentials Arguments

* `api_key` (Required) - A unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.

##### Slack Connector Profile Credentials Arguments

* `access_token` (Optional) - The credentials used to access protected Slack resources.

* `client_id` (Required) - The identifier for the client.

* `client_secret` (Required) - The client secret used by the OAuth client to authenticate to the authorization server.

* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint.

##### Snowflake Connector Profile Credentials Arguments

* `password` (Required) - The password that corresponds to the user name.

* `username` (Required) - The name of the user.

##### Trendmicro Connector Profile Credentials Arguments

* `api_secret_key` (Required) - The Secret Access Key portion of the credentials.

##### Veeva Connector Profile Credentials Arguments

* `password` (Required) - The password that corresponds to the user name.

* `username` (Required) - The name of the user.

##### Zendesk Connector Profile Credentials Arguments

* `access_token` (Optional) - The credentials used to access protected Zendesk resources.

* `client_id` (Required) - The identifier for the desired client.

* `client_secret` (Required) - The client secret used by the OAuth client to authenticate to the authorization server.

* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint.

##### Connector Profile Properties Arguments

* `amplitude` (Optional) - The connector-specific [properties](#amplitude-connector-profile-properties-arguments) required by Amplitude. Configure as an empty block `{}`.

* `datadog` (Optional) - The connector-specific [properties](#datadog-connector-profile-properties-arguments) required by Datadog.

* `dynatrace` (Optional) - The connector-specific [properties](#dynatrace-connector-profile-properties-arguments) required by Dynatrace.

* `google_analytics` (Optional) - The connector-specific [properties](#google-connector-profile-properties-arguments) required by Google Analytics. Configure as an empty block `{}`.

* `honeycode` (Optional) - The connector-specific [properties](#honeycode-connector-profile-properties-arguments) required by Amazon Honeycode. Configure as an empty block `{}`.

* `infor_nexus` (Optional) - The connector-specific [properties](#infor-connector-profile-properties-arguments) required by Infor Nexus.

* `marketo` (Optional) - The connector-specific [properties](#marketo-connector-profile-properties-arguments) required by Marketo.

* `redshift` (Optional) - The connector-specific [properties](#redshift-connector-profile-properties-arguments) required by Amazon Redshift.

* `salesforce` (Optional) - The connector-specific [properties](#salesforce-connector-profile-properties-arguments) required by Salesforce.

* `service_now` (Optional) - The connector-specific [properties](#servicenow-connector-profile-properties-arguments) required by ServiceNow.

* `singular` (Optional) - The connector-specific [properties](#singular-connector-profile-properties-arguments) required by Singular. Configure as an empty block `{}`.

* `slack` (Optional) - The connector-specific [properties](#slack-connector-profile-properties-arguments) required by Slack.

* `snowflake` (Optional) - The connector-specific [properties](#snowflake-connector-profile-properties-arguments) required by Snowflake.

* `trendmicro` (Optional) - The connector-specific [properties](#trendmicro-connector-profile-properties-arguments) required by Trend Micro. Configure as an empty block `{}`.

* `veeva` (Optional) - The connector-specific [properties](#veeva-connector-profile-properties-arguments) required by Veeva.

* `zendesk` (Optional) - The connector-specific [properties](#zendesk-connector-profile-properties-arguments) required by Zendesk.

##### Datadog Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Datadog resource.

##### Dynatrace Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Dynatrace resource.

##### Infor Nexus Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Infor Nexus resource.

##### Marketo Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Marketo resource.

##### Redshift Connector Profile Properties Arguments

* `bucket_name` (Required) - A name for the associated Amazon S3 bucket.

* `bucket_prefix` (Optional) - The object key for the destination bucket in which Amazon AppFlow places the files.

* `database_url` (Required) - The JDBC URL of the Amazon Redshift cluster.

* `role_arn` (Required) - The Amazon Resource Name (ARN) of the IAM role.

##### Salesforce Connector Profile Properties Arguments

* `instance_url` (Optional) - The location of the Salesforce resource.

* `is_sandbox_environment` (Optional) - Indicates whether the connector profile applies to a sandbox or production environment.

##### ServiceNow Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the ServiceNow resource.

##### Slack Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Slack resource.

##### Snowflake Connector Profile Properties Arguments

* `account_name` (Optional) - The name of the account.

* `bucket_name` (Required) - The name of the Amazon S3 bucket associated with Snowflake.

* `bucket_prefix` (Optional) - The bucket path that refers to the Amazon S3 bucket associated with Snowflake.

* `private_link_service_name` (Optional) - The Snowflake Private Link service name to be used for private data transfers.

* `region` (Optional) - The AWS Region of the Snowflake account.

* `stage` (Required) - The name of the Amazon S3 stage that was created while setting up an Amazon S3 stage in the Snowflake account. This is written in the following format: `<Database>.<Schema>.<Stage Name>`.

* `warehouse` (Required) - The name of the Snowflake warehouse.

##### Veeva Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Veeva resource.

##### Zendesk Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Zendesk resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `credentials_arn` - The Amazon Resource Name (ARN) of the connector profile credentials.

* `connector_profile_arn` - The Amazon Resource Name (ARN) of the connector profile.

* `created_at` - Specifies when the connector profile was created.

* `last_updated_at` - Specifies when the connector profile was last updated.


## Import

AppFlow Connector Profile can be imported using the connector profile name, e.g.

```
$ terraform import aws_appflow_connector_profile.profile connector-profile-name
```

[1]: https://docs.aws.amazon.com/appflow/1.0/APIReference/Welcome.html
[2]: https://docs.aws.amazon.com/appflow/1.0/APIReference/API_CreateConnectorProfile.html

