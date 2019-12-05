---
subcategory: "Greengrass"
layout: "aws"
page_title: "AWS: aws_greengrass_function_definition"
description: |-
    Creates and manages an AWS IoT Greengrass Function Definition
---

# Resource: aws_greengrass_function_definition

## Example Usage

```hcl
resource "aws_greengrass_function_definition" "test" {
	name = "function_definition"
	function_definition_version {
		default_config {
			isolation_mode = "GreengrassContainer"
			run_as {
				gid = 1
				uid = 1
			}
		}
		function {
			function_arn = "arn:aws:lambda:us-west-2:504366010147:function:test_lambda_wv8l0glb:test"
			id = "test_id"

			function_configuration {
				encoding_type = "json"
				exec_args = "arg"
				executable = "exec_func"
				memory_size = 1
				pinned = false
				timeout = 2

				environment {
					access_sysfs = false
					variables = {
						"var" = "val",
					}

					execution {
						isolation_mode = "GreengrassContainer"
						run_as {
							gid = 2
							uid = 2
						}
					}

					resource_access_policy {
						permission = "rw"
						resource_id = "1"
					}
				}
			}
		}
	}
}
```

## Argument Reference
* `name` - (Required) The name of the function definition.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the function definition.
* `function_definition_version` - (Optional) Object. Information about a function definition version.

The `function_definition_version` object has such arguments.
* `default_config` - (Optional) Object. The default configuration that applies to all Lambda functions in this function definition version. Individual Lambda functions can override these settings.
* `function` - (Optional) List of Object. A list of Lambda functions in this function definition version/

The `default_config` object has such arguments:
* `isolation_mode` - (Optional) String. 
* `run_as` - (Optional) Object.

The `function` object has such arguments:
* `function_arn` - (Required) String. The ARN of the Lambda function.
* `id` - (Required) String. A descriptive or arbitrary ID for the core. This value must be unique within the core definition version. Max length is 128 characters with pattern [a-zA-Z0-9:_-]+.
* `function_configuration` - (Optional). The configuration of the Lambda function.

The `function_configuration` object has such arguments.
* `encoding_type` - (Optional) String. The expected encoding type of the input payload for the function. The default is ''json''.
* `exec_args` - (Optional) String. The execution arguments.
* `executable` - (Optional) String. The name of the function executable.
* `memory_size` - (Optional) Number. The memory size, in KB, which the function requires. This setting is not applicable and should be cleared when you run the Lambda function without containerization.
* `pinned` - (Optional) Bool. True if the function is pinned. Pinned means the function is long-lived and starts when the core starts.
* `timeout` - (Optional) Number. The allowed function execution time, after which Lambda should terminate the function. This timeout still applies to pinned Lambda functions for each request.
* `environment` - (Optional) Object. The environment configuration of the function.

The `environment` has such attributes.
* `access_sysfs` - (Optional) Bool. If true, the Lambda function is allowed to access the host's /sys folder. Use this when the Lambda function needs to read device information from /sys. This setting applies only when you run the Lambda function in a Greengrass container.
* `execution` - (Optional) Object. Configuration related to executing the Lambda function
* `resource_access_policy` - (Optional) List of objects. A list of the resources, with their permissions, to which the Lambda function will be granted access. A Lambda function can have at most 10 resources. ResourceAccessPolicies apply only when you run the Lambda function in a Greengrass container.
* `variables` - (Optional) Map. Environment variables for the Lambda function's configuration.

The `execution` object has such arguments:
* `isolation_mode` - (Optional) String. 
* `run_as` - (Optional) Object.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:
* `arn` - The ARN of the group
* `latest_definition_version_arn` - The ARN of latest function definition version

## Environment variables
If you use `function_definition_version` object you should set `AMZN_CLIENT_TOKEN` as environmental variable.

## Import
IoT Greengrass Logger Definition can be imported using the `id`, e.g.
```
$ terraform import aws_greengrass_function_definition.definition <function_definition_id>
``` 