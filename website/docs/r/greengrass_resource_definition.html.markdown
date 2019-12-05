---
subcategory: "Greengrass"
layout: "aws"
page_title: "AWS: aws_greengrass_resource_definition"
description: |-
    Creates and manages an AWS IoT Greengrass Resource Definition
---

# Resource: aws_greengrass_resource_definition

## Example Usage

```hcl
resource "aws_greengrass_resource_definition" "test" {
	name = "resource_definition"
	resource_definition_version {
		resource {
			id = "test_id"
			name = "test_name"
			data_container {
				local_device_resource_data {
					source_path = "/dev/source"

					group_owner_setting {
						auto_add_group_owner = false
						group_owner = "user"
					}
				}
			}
		}
	}
}
```

## Argument Reference
* `name` - (Required) The name of the resource definition.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the resource definition.
* `resource_definition_version` - (Optional) Object. Information about a resource definition version.

The `resource_definition_version` object has such arguments.
* `resource` - (Optional) List. Component of resource definition.

The `resource` object has such arguments:
* `id` - (Required) String. The resource ID, used to refer to a resource in the Lambda function configuration. Max length is 128 characters with pattern ''[a-zA-Z0-9:_-]+''. This must be unique within a Greengrass group.
* `name` - (Required) String. The descriptive resource name, which is displayed on the AWS IoT Greengrass console. Max length 128 characters with pattern ''[a-zA-Z0-9:_-]+''. This must be unique within a Greengrass group.
* `data_container` - (Required) Object. A container of data for all resource types.

The `data_container` object has such arguments. The container takes only one of the following supported resource data types.
* `local_device_resource_data` - (Optional) Object. Attributes that define the local device resource.
* `local_volume_resource_data` - (Optional) Object. Attributes that define the local volume resource.
* `s3_machine_learning_model_resource_data` - (Optional) Object. Attributes that define an Amazon S3 machine learning resource.
* `sagemaker_machine_learning_model_resource_data` - (Optional) Object. Attributes that define an Amazon SageMaker machine learning resource
* `secrets_manager_secret_resource_data` - (Optional) Object. Attributes that define a secret resource, which references a secret from AWS Secrets Manager.

The `local_device_resource_data` object has such arguments.
* `source_path` - (Optional) String. The local absolute path of the device resource. The source path for a device resource can refer only to a character device or block device under ''/dev''.
* `group_owner_setting` - (Optional) Object. Group/owner related settings for local resources. 

The `local_volume_resource_data` object has such arguments.
* `source_path` - (Optional) String.  The local absolute path of the volume resource on the host. The source path for a volume resource type cannot start with ''/sys''.
* `destination_path` - (Optional) String. The absolute local path of the resource inside the Lambda environment.
* `group_owner_setting` - (Optional) Object. Allows you to configure additional group privileges for the Lambda process. This field is optional.

The `s3_machine_learning_model_resource_data` object has such arguments.
* `destination_path` - (Optional) String. The absolute local path of the resource inside the Lambda environment.
* `s3_uri` - (Optional) String. The URI of the source model in an S3 bucket. The model package must be in tar.gz or .zip format

The `sagemaker_machine_learning_model_resource_data` object has such arguments.
* `destination_path` - (Optional) String. The absolute local path of the resource inside the Lambda environment.
* `sagemaker_job_arn` - (Optional) String. The ARN of the Amazon SageMaker training job that represents the source model.

The `secrets_manager_secret_resource_data` object has such arguments.
* `secret_arn` - (Optional) String. The ARN of the Secrets Manager secret to make available on the core. The value of the secret's latest version (represented by the ''AWSCURRENT'' staging label) is included by default.
* `additional_staging_labels_to_download` - (Optional) List of String. The staging labels whose values you want to make available on the core, in addition to ''AWSCURRENT''.

The `group_owner_setting` object has such arguments.
* `auto_add_group_owner` - (Optional) Bool. If true, AWS IoT Greengrass automatically adds the specified Linux OS group owner of the resource to the Lambda process privileges. Thus the Lambda process will have the file access permissions of the added Linux group.

* `group_owner` - (Optional) String. The name of the Linux OS group whose privileges will be added to the Lambda process. This field is optional.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:
* `arn` - The ARN of the group
* `latest_definition_version_arn` - The ARN of latest resource definition version

## Environment variables
If you use `resource_definition_version` object you should set `AMZN_CLIENT_TOKEN` as environmental variable.

## Import
IoT Greengrass Logger Definition can be imported using the `id`, e.g.
```
$ terraform import aws_greengrass_resource_definition.definition <resource_definition_id>
``` 