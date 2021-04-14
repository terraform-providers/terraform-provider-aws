---
subcategory: "Managed Workflows for Apache Airflow (MWAA)"
layout: "aws"
page_title: "AWS: aws_mwaa_environment"
description: |-
  Creates a MWAA Environment
---

# Resource: aws_mwaa_environment

Creates a MWAA Environment resource.

## Example Usage

A MWAA Environment requires an IAM role (`aws_iam_role`), two subnets in the private zone (`aws_subnet`) and a versioned S3 bucket (`aws_s3_bucket`).

### Basic Usage

```terraform
resource "aws_mwaa_environment" "example" {
  dag_s3_path        = "dags/"
  execution_role_arn = aws_iam_role.example.arn
  name               = "example"

  network_configuration {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = aws_subnet.private[*].id
  }

  source_bucket_arn = aws_s3_bucket.example.arn
}
```

### Example with Airflow configuration options

```terraform
resource "aws_mwaa_environment" "example" {
  airflow_configuration_options = {
    "core.default_task_retries" = 16
    "core.parallelism"          = 1
  }

  dag_s3_path        = "dags/"
  execution_role_arn = aws_iam_role.example.arn
  name               = "example"

  network_configuration {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = aws_subnet.private[*].id
  }

  source_bucket_arn = aws_s3_bucket.example.arn
}
```

### Example with logging configurations

Note that Airflow task logs are enabled by default with the `INFO` log level.

```terraform
resource "aws_mwaa_environment" "example" {
  dag_s3_path        = "dags/"
  execution_role_arn = aws_iam_role.example.arn

  logging_configuration {
    dag_processing_logs {
      enabled   = true
      log_level = "DEBUG"
    }

    scheduler_logs {
      enabled   = true
      log_level = "INFO"
    }

    task_logs {
      enabled   = true
      log_level = "WARNING"
    }

    webserver_logs {
      enabled   = true
      log_level = "ERROR"
    }

    worker_logs {
      enabled   = true
      log_level = "CRITICAL"
    }
  }

  name = "example"

  network_configuration {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = aws_subnet.private[*].id
  }

  source_bucket_arn = aws_s3_bucket.example.arn
}
```

### Example with tags

```terraform
resource "aws_mwaa_environment" "example" {
  dag_s3_path        = "dags/"
  execution_role_arn = aws_iam_role.example.arn
  name               = "example"

  network_configuration {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = aws_subnet.private[*].id
  }

  source_bucket_arn = aws_s3_bucket.example.arn

  tags = {
    Name        = "example"
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are supported:

* `airflow_configuration_options` - (Optional) The `airflow_configuration_options` parameter specifies airflow override options. Check the [Official documentation](https://docs.aws.amazon.com/mwaa/latest/userguide/configuring-env-variables.html#configuring-env-variables-reference) for all possible configuration options.
* `airflow_version` - (Optional) Airflow version of your environment, will be set by default to the latest version that MWAA supports.
* `dag_s3_path` - (Required) The relative path to the DAG folder on your Amazon S3 storage bucket. For example, dags. For more information, see [Importing DAGs on Amazon MWAA](https://docs.aws.amazon.com/mwaa/latest/userguide/configuring-dag-import.html).
* `environment_class` - (Optional) Environment class for the cluster. Possible options are `mw1.small`, `mw1.medium`, `mw1.large`. Will be set by default to `mw1.small`. Please check the [AWS Pricing](https://aws.amazon.com/de/managed-workflows-for-apache-airflow/pricing/) for more information about the environment classes.
* `execution_role_arn` - (Required) The Amazon Resource Name (ARN) of the task execution role that the Amazon MWAA and its environment can assume. Check the [official AWS documentation](https://docs.aws.amazon.com/mwaa/latest/userguide/mwaa-create-role.html) for the detailed role specification.
* `kms_key` - (Optional) The Amazon Resource Name (ARN) of your KMS key that you want to use for encryption. Will be set to the ARN of the managed KMS key `aws/airflow` by default. Please check the [Official Documentation](https://docs.aws.amazon.com/mwaa/latest/userguide/custom-keys-certs.html) for more information.
* `logging_configuration` - (Optional) The Apache Airflow logs you want to send to Amazon CloudWatch Logs.
* `max_workers` - (Optional) The maximum number of workers that can be automatically scaled up. Value need to be between `1` and `25`. Will be `10` by default.
* `min_workers` - (Optional) The minimum number of workers that you want to run in your environment. Will be `1` by default.
* `name` - (Required) The name of the Apache Airflow Environment
* `network_configuration` - (Required) Specifies the network configuration for your Apache Airflow Environment. This includes two private subnets as well as security groups for the Airflow environment. Each subnet requires internet connection, otherwise the deployment will fail. See [Network configuration](#network) below for details.
* `plugins_s3_object_version` - (Optional) The plugins.zip file version you want to use.
* `plugins_s3_path` - (Optional) The relative path to the plugins.zip file on your Amazon S3 storage bucket. For example, plugins.zip. If a relative path is provided in the request, then plugins_s3_object_version is required. For more information, see [Importing DAGs on Amazon MWAA](https://docs.aws.amazon.com/mwaa/latest/userguide/configuring-dag-import.html).
* `requirements_s3_object_version` - (Optional) The requirements.txt file version you want to use.
* `requirements_s3_path` - (Optional) The relative path to the requirements.txt file on your Amazon S3 storage bucket. For example, requirements.txt. If a relative path is provided in the request, then requirements_s3_object_version is required. For more information, see [Importing DAGs on Amazon MWAA](https://docs.aws.amazon.com/mwaa/latest/userguide/configuring-dag-import.html).
* `source_bucket_arn` - (Required) The Amazon Resource Name (ARN) of your Amazon S3 storage bucket. For example, arn:aws:s3:::airflow-mybucketname.
* `webserver_access_mode` - (Optional) Specifies whether the webserver should be accessible over the internet or via your specified VPC. Possible options: `PRIVATE_ONLY` (default) and `PUBLIC_ONLY`.
* `weekly_maintenance_window_start` - (Optional) Specifies the start date for the weekly maintenance window.
* `tags` - (Optional) An array of key:value pairs to associate with the resource.

### Logging configurations

The `logging_configuration` block supports the following arguments.

* `dag_processing_logs` - (Optional) (Optional) Log configuration options for processing DAGs. See [Module logging configuration](#module-logging-configuration) for more information. Disabled by default.
* `scheduler_logs` - (Optional) Log configuration options for the schedulers. See [Module logging configuration](#module-logging-configuration) for more information. Disabled by default.
* `task_logs` - (Optional) Log configuration options for DAG tasks. See [Module logging configuration](#module-logging-configuration) for more information. Enabled by default with `INFO` log level.
* `webserver_logs` - (Optional) Log configuration options for the webservers. See [Module logging configuration](#module-logging-configuration) for more information. Disabled by default.
* `worker_logs` - (Optional) Log configuration options for the workers. See [Module logging configuration](#module-logging-configuration) for more information. Disabled by default.

### Module logging configuration

A configuration block to use for logging with respect to the various Apache Airflow services: DagProcessingLogs, SchedulerLogs, TaskLogs, WebserverLogs, and WorkerLogs. It supports the following arguments.

* `enabled` - (Required) Enabling or disabling the collection of logs
* `log_level` - (Optional) Logging level. Valid values: `CRITICAL`, `ERROR`, `WARNING`, `INFO`, `DEBUG`. Will be `INFO` by default.

### Network configuration

The `network_configuration` block supports the following arguments. More information about the required subnet and security group settings can be found in the [official AWS documentation](https://docs.aws.amazon.com/mwaa/latest/userguide/vpc-create.html).

* `security_group_ids` - (Required) Security groups IDs for the environment. At least one of the security group needs to allow MWAA resources to talk to each other, otherwise MWAA cannot be provisioned.
* `subnet_ids` - (Required)  The private subnet IDs in which the environment should be created. MWAA requires two subnets.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the MWAA Environment
* `created_at` - The Created At date of the MWAA Environment
* `logging_configuration.<LOG_TYPE>.cloud_watch_log_group_arn` - Provides the ARN for the CloudWatch group where the logs will be published
* `service_role_arn` - The Service Role ARN of the Amazon MWAA Environment
* `status` - The status of the Amazon MWAA Environment
* `webserver_url` - The webserver URL of the MWAA Environment


## Import

MWAA Environment can be imported using `Name` e.g.

```
$ terraform import aws_mwaa_environment.example MyAirflowEnvironment
```
