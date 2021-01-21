---
subcategory: "ECS"
layout: "aws"
page_title: "AWS: aws_ecs_task_definition"
description: |-
  Manages a revision of an ECS task definition.
---

# Resource: aws_ecs_task_definition

Manages a revision of an ECS task definition to be used in `aws_ecs_service`.

## Example Usage

```hcl
resource "aws_ecs_task_definition" "service" {
  family                = "service"
  container_definitions = file("task-definitions/service.json")

  volume {
    name      = "service-storage"
    host_path = "/ecs/service-storage"
  }

  placement_constraints {
    type       = "memberOf"
    expression = "attribute:ecs.availability-zone in [us-west-2a, us-west-2b]"
  }
}
```

The referenced `task-definitions/service.json` file contains a valid JSON document,
which is shown below, and its content is going to be passed directly into the
`container_definitions` attribute as a string. Please note that this example
contains only a small subset of the available parameters.

```json
[
  {
    "name": "first",
    "image": "service-first",
    "cpu": 10,
    "memory": 512,
    "essential": true,
    "portMappings": [
      {
        "containerPort": 80,
        "hostPort": 80
      }
    ]
  },
  {
    "name": "second",
    "image": "service-second",
    "cpu": 10,
    "memory": 256,
    "essential": true,
    "portMappings": [
      {
        "containerPort": 443,
        "hostPort": 443
      }
    ]
  }
]
```

### With AppMesh Proxy

```hcl
resource "aws_ecs_task_definition" "service" {
  family                = "service"
  container_definitions = file("task-definitions/service.json")

  proxy_configuration {
    type           = "APPMESH"
    container_name = "applicationContainerName"
    properties = {
      AppPorts         = "8080"
      EgressIgnoredIPs = "169.254.170.2,169.254.169.254"
      IgnoredUID       = "1337"
      ProxyEgressPort  = 15001
      ProxyIngressPort = 15000
    }
  }
}
```

## Argument Reference

### Top-Level Arguments

* `family` - (Required) A unique name for your task definition.
* `container_definitions` - (Required) A list of valid [container
definitions](http://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_ContainerDefinition.html)
provided as a single valid JSON document. Please note that you should only
provide values that are part of the container definition document. For a
detailed description of what parameters are available, see the [Task Definition
Parameters](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html)
section from the official [Developer Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide).

~> **NOTE**: Proper escaping is required for JSON field values containing quotes (`"`) such as `environment` values. If directly setting the JSON, they should be escaped as `\"` in the JSON,  e.g. `"value": "I \"love\" escaped quotes"`. If using a Terraform variable value, they should be escaped as `\\\"` in the variable, e.g. `value = "I \\\"love\\\" escaped quotes"` in the variable and `"value": "${var.myvariable}"` in the JSON.

* `task_role_arn` - (Optional) The ARN of IAM role that allows your Amazon ECS container task to make calls to other AWS services.
* `execution_role_arn` - (Optional) The Amazon Resource Name (ARN) of the task execution role that the Amazon ECS container agent and the Docker daemon can assume.
* `network_mode` - (Optional) The Docker networking mode to use for the containers in the task. The valid values are `none`, `bridge`, `awsvpc`, and `host`.
* `ipc_mode` - (Optional) The IPC resource namespace to be used for the containers in the task The valid values are `host`, `task`, and `none`.
* `pid_mode` - (Optional) The process namespace to use for the containers in the task. The valid values are `host` and `task`.
* `volume` - (Optional) A set of [volume blocks](#volume-block-arguments) that containers in your task may use.
* `placement_constraints` - (Optional) A set of [placement constraints](#placement-constraints-arguments) rules that are taken into consideration during task placement. Maximum number of `placement_constraints` is `10`.
* `cpu` - (Optional) The number of cpu units used by the task. If the `requires_compatibilities` is `FARGATE` this field is required.
* `memory` - (Optional) The amount (in MiB) of memory used by the task. If the `requires_compatibilities` is `FARGATE` this field is required.
* `requires_compatibilities` - (Optional) A set of launch types required by the task. The valid values are `EC2` and `FARGATE`.
* `proxy_configuration` - (Optional) The [proxy configuration](#proxy-configuration-arguments) details for the App Mesh proxy.
* `inference_accelerator` - (Optional) Configuration block(s) with Inference Accelerators settings. Detailed below.
* `tags` - (Optional) Key-value map of resource tags

#### Volume Block Arguments

* `name` - (Required) The name of the volume. This name is referenced in the `sourceVolume`
parameter of container definition in the `mountPoints` section.
* `host_path` - (Optional) The path on the host container instance that is presented to the container. If not set, ECS will create a nonpersistent data volume that starts empty and is deleted after the task has finished.
* `docker_volume_configuration` - (Optional) Used to configure a [docker volume](#docker-volume-configuration-arguments)
* `efs_volume_configuration` - (Optional) Used to configure a [EFS volume](#efs-volume-configuration-arguments).

#### Docker Volume Configuration Arguments

For more information, see [Specifying a Docker volume in your Task Definition Developer Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/docker-volumes.html#specify-volume-config)

* `scope` - (Optional) The scope for the Docker volume, which determines its lifecycle, either `task` or `shared`.  Docker volumes that are scoped to a `task` are automatically provisioned when the task starts and destroyed when the task stops. Docker volumes that are scoped as `shared` persist after the task stops.
* `autoprovision` - (Optional) If this value is `true`, the Docker volume is created if it does not already exist. *Note*: This field is only used if the scope is `shared`.
* `driver` - (Optional) The Docker volume driver to use. The driver value must match the driver name provided by Docker because it is used for task placement.
* `driver_opts` - (Optional) A map of Docker driver specific options.
* `labels` - (Optional) A map of custom metadata to add to your Docker volume.

##### Example Usage

```hcl
resource "aws_ecs_task_definition" "service" {
  family                = "service"
  container_definitions = file("task-definitions/service.json")

  volume {
    name = "service-storage"

    docker_volume_configuration {
      scope         = "shared"
      autoprovision = true
      driver        = "local"

      driver_opts = {
        "type"   = "nfs"
        "device" = "${aws_efs_file_system.fs.dns_name}:/"
        "o"      = "addr=${aws_efs_file_system.fs.dns_name},rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2,noresvport"
      }
    }
  }
}
```

#### EFS Volume Configuration Arguments

For more information, see [Specifying an EFS volume in your Task Definition Developer Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/efs-volumes.html#specify-efs-config)

* `file_system_id` - (Required) The ID of the EFS File System.
* `root_directory` - (Optional) The directory within the Amazon EFS file system to mount as the root directory inside the host. If this parameter is omitted, the root of the Amazon EFS volume will be used. Specifying / will have the same effect as omitting this parameter. This argument is ignored when using `authorization_config`.
* `transit_encryption` - (Optional) Whether or not to enable encryption for Amazon EFS data in transit between the Amazon ECS host and the Amazon EFS server. Transit encryption must be enabled if Amazon EFS IAM authorization is used. Valid values: `ENABLED`, `DISABLED`. If this parameter is omitted, the default value of `DISABLED` is used.
* `transit_encryption_port` - (Optional) The port to use for transit encryption. If you do not specify a transit encryption port, it will use the port selection strategy that the Amazon EFS mount helper uses.
* `authorization_config` - (Optional) The authorization configuration details for the Amazon EFS file system.
    * `access_point_id` - The access point ID to use. If an access point is specified, the root directory value will be relative to the directory set for the access point. If specified, transit encryption must be enabled in the EFSVolumeConfiguration.
    * `iam` - Whether or not to use the Amazon ECS task IAM role defined in a task definition when mounting the Amazon EFS file system. If enabled, transit encryption must be enabled in the EFSVolumeConfiguration. Valid values: `ENABLED`, `DISABLED`. If this parameter is omitted, the default value of `DISABLED` is used.

##### Example Usage

```hcl
resource "aws_ecs_task_definition" "service" {
  family                = "service"
  container_definitions = file("task-definitions/service.json")

  volume {
    name = "service-storage"

    efs_volume_configuration {
      file_system_id          = aws_efs_file_system.fs.id
      root_directory          = "/opt/data"
      transit_encryption      = "ENABLED"
      transit_encryption_port = 2999
      authorization_config {
        access_point_id = aws_efs_access_point.test.id
        iam             = "ENABLED"
      }
    }
  }
}
```


#### Placement Constraints Arguments

* `type` - (Required) The type of constraint. Use `memberOf` to restrict selection to a group of valid candidates.
Note that `distinctInstance` is not supported in task definitions.
* `expression` -  (Optional) Cluster Query Language expression to apply to the constraint.
For more information, see [Cluster Query Language in the Amazon EC2 Container
Service Developer
Guide](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/cluster-query-language.html).

#### Proxy Configuration Arguments

* `container_name` - (Required) The name of the container that will serve as the App Mesh proxy.
* `properties` - (Required) The set of network configuration parameters to provide the Container Network Interface (CNI) plugin, specified a key-value mapping.
* `type` - (Optional) The proxy type. The default value is `APPMESH`. The only supported value is `APPMESH`.

#### Inference Accelerators Arguments

* `device_name` - (Required) The Elastic Inference accelerator device name. The deviceName must also be referenced in a container definition as a ResourceRequirement.
* `device_type` - (Required) The Elastic Inference accelerator type to use.

##### Example Usage

```hcl
resource "aws_ecs_task_definition" "test" {
  family                = "test"
  container_definitions = <<TASK_DEFINITION
[
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"environment": [
			{"name": "VARNAME", "value": "VARVAL"}
		],
		"essential": true,
		"image": "jenkins",
		"memory": 128,
		"name": "jenkins",
		"portMappings": [
			{
				"containerPort": 80,
				"hostPort": 8080
			}
		],
        "resourceRequirements":[
            {
                "type":"InferenceAccelerator",
                "value":"device_1"
            }
        ]
	}
]
TASK_DEFINITION

  inference_accelerator {
    device_name = "device_1"
    device_type = "eia1.medium"
  }
}
```

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Full ARN of the Task Definition (including both `family` and `revision`).
* `family` - The family of the Task Definition.
* `revision` - The revision of the task in a particular family.

## Import

ECS Task Definitions can be imported via their Amazon Resource Name (ARN):

```
$ terraform import aws_ecs_task_definition.example arn:aws:ecs:us-east-1:012345678910:task-definition/mytaskfamily:123
```
