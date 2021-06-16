---
subcategory: "EventBridge (CloudWatch Events)"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_connection"
description: |-
  Provides an EventBridge connection resource.
---

# Resource: aws_cloudwatch_event_connection

Provides an EventBridge connection resource.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.


## Example Usage

```terraform
resource "aws_cloudwatch_event_connection" "test" {
  name               = "ngrok-connection"
  description        = "A connection description"
  authorization_type = "API_KEY"

  auth_parameters {
    api_key {
      key   = "x-signature"
      value = "1234"
    }
  }
}
```

## Example Usage Basic Authorization

```terraform
resource "aws_cloudwatch_event_connection" "test" {
  name               = "ngrok-connection"
  description        = "A connection description"
  authorization_type = "BASIC"

  auth_parameters {
    basic {
      username = "user"
      password = "Pass1234!"
    }
  }
}
```

## Example Usage OAuth Authorization

```terraform
resource "aws_cloudwatch_event_connection" "test" {
  name               = "ngrok-connection"
  description        = "A connection description"
  authorization_type = "BASIC"

  auth_parameters {
    oauth {
      authorization_endpoint = "https://auth.url.com/endpoint"
      http_method            = "GET"

      client_parameters {
        client_id     = "1234567890"
        client_secret = "Pass1234!"
      }

      oauth_http_parameters {
        body {
          key             = "body-parameter-key"
          value           = "body-parameter-value"
          is_value_secret = false
        }

        header {
          key             = "header-parameter-key"
          value           = "header-parameter-value"
          is_value_secret = false
        }

        query_string {
          key             = "query-string-parameter-key"
          value           = "query-string-parameter-value"
          is_value_secret = false
        }
      }
    }
  }
}
```

## Example Usage Invocation Http Parameters

```terraform
resource "aws_cloudwatch_event_connection" "test" {
  name               = "ngrok-connection"
  description        = "A connection description"
  authorization_type = "BASIC"

  auth_parameters {
    basic {
      username = "user"
      password = "Pass1234!"
    }

    invocation_http_parameters {
      body {
        key             = "body-parameter-key"
        value           = "body-parameter-value"
        is_value_secret = false
      }

      body {
        key             = "body-parameter-key2"
        value           = "body-parameter-value2"
        is_value_secret = true
      }

      header {
        key             = "header-parameter-key"
        value           = "header-parameter-value"
        is_value_secret = false
      }

      query_string {
        key             = "query-string-parameter-key"
        value           = "query-string-parameter-value"
        is_value_secret = false
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the new connection. Maximum of 64 characters consisting of numbers, lower/upper case letters, .,-,_.
* `description` - (Optional) Enter a description for the connection. Maximum of 512 characters.
* `authorization_type` - (Required) Choose the type of authorization to use for the connection. One of `API_KEY`,`BASIC`,`OAUTH_CLIENT_CREDENTIALS`.
* `auth_parameters` - (Required) Parameters used for authorization. A maximum of 1 are allowed. Documented below.
* `invocation_http_parameters` - (Optional) Invocation Http Parameters are additional credentials used to sign each Invocation of the ApiDestination created from this Connection. If the ApiDestination Rule Target has additional HttpParameters, the values will be merged together, with the Connection Invocation Http Parameters taking precedence. Secret values are stored and managed by AWS Secrets Manager. A maximum of 1 are allowed. Documented below.

`auth_parameters` support the following:

* `api_key` - (Optional) Parameters used for API_KEY authorization. An API key to include in the header for each authentication request. A maximum of 1 are allowed. Conflicts with `basic` and `oauth`. Documented below.
* `basic` - (Optional) Parameters used for BASIC authorization. A maximum of 1 are allowed. Conflicts with `api_key` and `oauth`. Documented below.
* `oauth` - (Optional) Parameters used for OAUTH_CLIENT_CREDENTIALS authorization. A maximum of 1 are allowed. Conflicts with `basic` and `api_key`. Documented below.

`api_key` support the following:

* `key` - (Required) Header Name.
* `value` - (Required) Header Value. Created and stored in AWS Secrets Manager.

`basic` support the following:

* `username` - (Required) A username for the authorization.
* `password` - (Required) A password for the authorization. Created and stored in AWS Secrets Manager.

`oauth` support the following:

* `authorization_endpoint` - (Required) A username for the authorization.
* `http_method` - (Required) A password for the authorization. Created and stored in AWS Secrets Manager.
* `client_parameters` - (Required) Contains the client parameters for OAuth authorization. Contains the following two parameters.
    * `client_id` - (Required) The client ID for the credentials to use for authorization. Created and stored in AWS Secrets Manager.
    * `client_secret` - (Required) The client secret for the credentials to use for authorization. Created and stored in AWS Secrets Manager.
* `oauth_http_parameters` - (Required) OAuth Http Parameters are additional credentials used to sign the request to the authorization endpoint to exchange the OAuth Client information for an access token. Secret values are stored and managed by AWS Secrets Manager. A maximum of 1 are allowed. Documented below.

`invocation_http_parameters` and `oauth_http_parameters` support the following:

* `body` - (Optional) Contains additional body string parameters for the connection. You can include up to 100 additional body string parameters per request. Each additional parameter counts towards the event payload size, which cannot exceed 64 KB. Each parameter can contain the following:
    * `key` - (Required) The key for the parameter.
    * `value` - (Required) The value associated with the key. Created and stored in AWS Secrets Manager if is secret.
    * `is_value_secret` - (Optional) Specified whether the value is secret.

* `header` - (Optional) Contains additional header parameters for the connection. You can include up to 100 additional body string parameters per request. Each additional parameter counts towards the event payload size, which cannot exceed 64 KB. Each parameter can contain the following:
    * `key` - (Required) The key for the parameter.
    * `value` - (Required) The value associated with the key. Created and stored in AWS Secrets Manager if is secret.
    * `is_value_secret` - (Optional) Specified whether the value is secret.

* `query_string` - (Optional) Contains additional query string parameters for the connection. You can include up to 100 additional body string parameters per request. Each additional parameter counts towards the event payload size, which cannot exceed 64 KB. Each parameter can contain the following:
    * `key` - (Required) The key for the parameter.
    * `value` - (Required) The value associated with the key. Created and stored in AWS Secrets Manager if is secret.
    * `is_value_secret` - (Optional) Specified whether the value is secret.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the connection.
* `secret_arn` - The Amazon Resource Name (ARN) of the secret created from the authorization parameters specified for the connection.


## Import

EventBridge Connection can be imported using the `name`, e.g.

```console
$ terraform import aws_cloudwatch_event_connection.test ngrok-connection
```
