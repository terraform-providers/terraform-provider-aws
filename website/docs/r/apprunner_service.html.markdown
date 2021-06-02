---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_service"
description: |-
  Manages an App Runner Service.
---

# Resource: aws_apprunner_service

Manages an App Runner Service.

## Example Usage

### Service with a Code Repository Source

```terraform
resource "aws_apprunner_service" "example" {
  service_name = "example"

  source_configuration {
    authentication_configuration {
      connection_arn = aws_apprunner_connection.example.arn
    }
    code_repository {
      code_configuration {
        code_configuration_values {
          build_command = "python setup.py develop"
          port          = "8000"
          runtime       = "python3"
          start_command = "python runapp.py"
        }
        configuration_source = "API"
      }
      repository_url = "https://github.com/example/my-example-python-app"
      source_code_version {
        type  = "BRANCH"
        value = "main"
      }
    }
  }

  tags = {
    Name = "example-apprunner-service"
  }
}
```

### Service with an Image Repository Source

```terraform
resource "aws_apprunner_service" "example" {
  service_name = "example"

  source_configuration {
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "public.ecr.aws/jg/hello:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }

  tags = {
    Name = "example-apprunner-service"
  }
}
```

## Argument Reference

The following arguments are required:

* `service_name` - (Forces new resource) Name of the service.
* `source_configuration` - The source to deploy to the App Runner service. Can be a code or an image repository. See [Source Configuration](#source-configuration) below for more details.

The following arguments are optional:

* `auto_scaling_configuration_arn` - ARN of an App Runner automatic scaling configuration resource that you want to associate with your service. If not provided, App Runner associates the latest revision of a default auto scaling configuration.
* `encryption_configuration` - (Forces new resource) An optional custom encryption key that App Runner uses to encrypt the copy of your source repository that it maintains and your service logs. By default, App Runner uses an AWS managed CMK. See [Encryption Configuration](#encryption-configuration) below for more details.
* `health_check_configuration` - (Forces new resource) Settings of the health check that AWS App Runner performs to monitor the health of your service. See [Health Check Configuration](#health-check-configuration) below for more details.
* `instance_configuration` - The runtime configuration of instances (scaling units) of the App Runner service. See [Instance Configuration](#instance-configuration) below for more details.
* `tags` - Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Encryption Configuration

The `encryption_configuration` block supports the following argument:

* `kms_key` - (Required) The ARN of the KMS key used for encryption.

### Health Check Configuration

The `health_check_configuration` block supports the following arguments:

* `healthy_threshold` - (Optional) The number of consecutive checks that must succeed before App Runner decides that the service is healthy. Defaults to 1. Minimum value of 1. Maximum value of 20.
* `interval` - (Optional) The time interval, in seconds, between health checks. Defaults to 5. Minimum value of 1. Maximum value of 20.
* `path` - (Optional) The URL to send requests to for health checks. Defaults to `/`. Minimum length of 0. Maximum length of 51200.
* `protocol` - (Optional) The IP protocol that App Runner uses to perform health checks for your service. Valid values: `TCP`, `HTTP`. Defaults to `TCP`. If you set protocol to `HTTP`, App Runner sends health check requests to the HTTP path specified by `path`.
* `timeout` - (Optional) The time, in seconds, to wait for a health check response before deciding it failed. Defaults to 2. Minimum value of  1. Maximum value of 20.
* `unhealthy_threshold` - (Optional) The number of consecutive checks that must fail before App Runner decides that the service is unhealthy. Defaults to 5. Minimum value of  1. Maximum value of 20.

### Instance Configuration

The `instance_configuration` block supports the following arguments:

* `cpu` - (Optional) The number of CPU units reserved for each instance of your App Runner service represented as a String. Defaults to `1024`. Valid values: `1024|2048|(1|2) vCPU`.
* `instance_role_arn` - (Required) The Amazon Resource Name (ARN) of an IAM role that provides permissions to your App Runner service. These are permissions that your code needs when it calls any AWS APIs.
* `memory` - (Optional) The amount of memory, in MB or GB, reserved for each instance of your App Runner service. Defaults to `2048`. Valid values: `2048|3072|4096|(2|3|4) GB`.

### Source Configuration

The `source_configuration` block supports the following arguments:

~>**Note:** Either `code_repository` or `image_repository` must be specified (but not both).

* `authentication_configuration` - (Optional) Describes resources needed to authenticate access to some source repositories. See [Authentication Configuration](#authentication-configuration) below for more details.
* `auto_deployments_enabled` - (Optional) Whether continuous integration from the source repository is enabled for the App Runner service. If set to `true`, each repository change (source code commit or new image version) starts a deployment. Defaults to `true`.
* `code_repository` - (Optional) Description of a source code repository. See [Code Repository](#code-repository) below for more details.
* `image_repository` - (Optional) Description of a source image repository. See [Image Repository](#image-repository) below for more details.

### Authentication Configuration

The `authentication_configuration` block supports the following arguments:

* `access_role_arn` - (Optional) ARN of the IAM role that grants the App Runner service access to a source repository. Required for ECR image repositories (but not for ECR Public)
* `connection_arn` - (Optional) ARN of the App Runner connection that enables the App Runner service to connect to a source repository. Required for GitHub code repositories.

### Code Repository

The `code_repository` block supports the following arguments:

* `code_configuration` - (Optional) Configuration for building and running the service from a source code repository. See [Code Configuration](#code-configuration) below for more details.
* `repository_url` - (Required) The location of the repository that contains the source code.
* `source_code_version` - (Required) The version that should be used within the source code repository. See [Source Code Version](#source-code-version) below for more details.

### Image Repository

The `image_repository` block supports the following arguments:

* `image_configuration` - (Optional) Configuration for running the identified image. See [Image Configuration](#image-configuration) below for more details.
* `image_identifier` - (Required) The identifier of an image. For an image in Amazon Elastic Container Registry (Amazon ECR), this is an image name. For the
  image name format, see Pulling an image in the Amazon ECR User Guide.
* `image_repository_type` - (Required) The type of the image repository. This reflects the repository provider and whether the repository is private or public. Valid values: `ECR` , `ECR_PUBLIC`.

### Code Configuration

The `code_configuration` block supports the following arguments:

* `code_configuration_values` - (Optional) Basic configuration for building and running the App Runner service. Use this parameter to quickly launch an App Runner service without providing an apprunner.yaml file in the source code repository (or ignoring the file if it exists). See [Code Configuration Values](#code-configuration-values) below for more details.
* `configuration_source` - (Required) The source of the App Runner configuration. Valid values: `REPOSITORY`, `API`. Values are interpreted as follows:
    * `REPOSITORY` - App Runner reads configuration values from the apprunner.yaml file in the
    source code repository and ignores the CodeConfigurationValues parameter.
    * `API` - App Runner uses configuration values provided in the CodeConfigurationValues
    parameter and ignores the apprunner.yaml file in the source code repository.

### Code Configuration Values

The `code_configuration_values` blocks supports the following arguments:

* `build_command` - (Optional) The command App Runner runs to build your application.
* `port` - (Optional) The port that your application listens to in the container. Defaults to `"8080"`.
* `runtime` - (Required) A runtime environment type for building and running an App Runner service. Represents a programming language runtime. Valid values: `python3`, `nodejs12`.
* `runtime_environment_variables` - (Optional) Environment variables available to your running App Runner service. A map of key/value pairs. Keys with a prefix of `AWSAPPRUNNER` are reserved for system use and aren't valid.
* `start_command` - (Optional) The command App Runner runs to start your application.

### Image Configuration

The `image_configuration` block supports the following arguments:

* `port` - (Optional) The port that your application listens to in the container. Defaults to `"8080"`.
* `runtime_environment_variables` - (Optional) Environment variables available to your running App Runner service. A map of key/value pairs. Keys with a prefix of `AWSAPPRUNNER` are reserved for system use and aren't valid.
* `start_command` - (Optional) A command App Runner runs to start the application in the source image. If specified, this command overrides the Docker image’s default start command.

### Source Code Version

The `source_code_version` block supports the following arguments:

* `type` - (Required) The type of version identifier. For a git-based repository, branches represent versions. Valid values: `BRANCH`.
* `value`- (Required) A source code version. For a git-based repository, a branch name maps to a specific version. App Runner uses the most recent commit to the branch.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the App Runner service.
* `service_id` - An alphanumeric ID that App Runner generated for this service. Unique within the AWS Region.
* `service_url` - A subdomain URL that App Runner generated for this service. You can use this URL to access your service web application.
* `status` - The current state of the App Runner service.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

App Runner Services can be imported by using the `arn`, e.g.

```
$ terraform import aws_apprunner_service.example arn:aws:apprunner:us-east-1:1234567890:service/example/0a03292a89764e5882c41d8f991c82fe
```
