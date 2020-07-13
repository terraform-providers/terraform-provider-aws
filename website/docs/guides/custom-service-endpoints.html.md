---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Custom Service Endpoint Configuration"
description: |-
  Configuring the Terraform AWS Provider to connect to custom AWS service endpoints and AWS compatible solutions.
---

# Custom Service Endpoint Configuration

The Terraform AWS Provider configuration can be customized to connect to non-default AWS service endpoints and AWS compatible solutions. This may be useful for environments with specific compliance requirements, such as using [AWS FIPS 140-2 endpoints](https://aws.amazon.com/compliance/fips/), connecting to AWS Snowball, SC2S, or C2S environments, or local testing.

This guide outlines how to get started with customizing endpoints, the available endpoint configurations, and offers example configurations for working with certain local development and testing solutions.

~> **NOTE:** Support for connecting the Terraform AWS Provider with custom endpoints and AWS compatible solutions is offered as best effort. Individual Terraform resources may require compatibility updates to work in certain environments. Integration testing by HashiCorp during provider changes is exclusively done against default AWS endpoints at this time.

<!-- TOC depthFrom:2 -->

- [Getting Started with Custom Endpoints](#getting-started-with-custom-endpoints)
- [Available Endpoint Customizations](#available-endpoint-customizations)
- [Connecting to Local AWS Compatible Solutions](#connecting-to-local-aws-compatible-solutions)
    - [DynamoDB Local](#dynamodb-local)
    - [LocalStack](#localstack)

<!-- /TOC -->

## Getting Started with Custom Endpoints

To configure the Terraform AWS Provider to use customized endpoints, it can be done within `provider` declarations using the `endpoints` configuration block, e.g.

```hcl
provider "aws" {
  # ... potentially other provider configuration ...

  endpoints {
    dynamodb = "http://localhost:4569"
    s3       = "http://localhost:4572"
  }
}
```

If multiple, different Terraform AWS Provider configurations are required, see the [Terraform documentation on multiple provider instances](https://www.terraform.io/docs/configuration/providers.html#alias-multiple-provider-instances) for additional information about the `alias` provider configuration and its usage.

## Available Endpoint Customizations

The Terraform AWS Provider allows the following endpoints to be customized:

<!-- markdownlint-disable MD033 -->
<!--
    The division splits this long list into multiple columns without manually
    maintaining a table. The terraform.io Markdown parser previously allowed
    for Markdown within HTML elements, however the Terraform Registry parser
    is more accurate/strict, so we use raw HTML to maintain this list.
-->
<div style="column-width: 14em;">

- `acm`
- `acmpca`
- `amplify`
- `apigateway`
- `applicationautoscaling`
- `applicationinsights`
- `appmesh`
- `appstream`
- `appsync`
- `athena`
- `autoscaling`
- `autoscalingplans`
- `backup`
- `batch`
- `budgets`
- `cloud9`
- `cloudformation`
- `cloudfront`
- `cloudhsm`
- `cloudsearch`
- `cloudtrail`
- `cloudwatch`
- `cloudwatchevents`
- `cloudwatchlogs`
- `codebuild`
- `codecommit`
- `codedeploy`
- `codepipeline`
- `cognitoidentity`
- `cognitoidp`
- `configservice`
- `cur`
- `datapipeline`
- `datasync`
- `dax`
- `devicefarm`
- `directconnect`
- `dlm`
- `dms`
- `docdb`
- `ds`
- `dynamodb`
- `ec2`
- `ecr`
- `ecs`
- `efs`
- `eks`
- `elasticache`
- `elasticbeanstalk`
- `elastictranscoder`
- `elb`
- `emr`
- `es`
- `firehose`
- `fms`
- `forecast`
- `fsx`
- `gamelift`
- `glacier`
- `globalaccelerator`
- `glue`
- `guardduty`
- `greengrass`
- `iam`
- `inspector`
- `iot`
- `iotanalytics`
- `iotevents`
- `kafka`
- `kinesis_analytics` (**DEPRECATED** Use `kinesisanalytics` instead)
- `kinesis`
- `kinesisanalytics`
- `kinesisvideo`
- `kms`
- `lakeformation`
- `lambda`
- `lexmodels`
- `licensemanager`
- `lightsail`
- `macie`
- `managedblockchain`
- `mediaconnect`
- `mediaconvert`
- `medialive`
- `mediapackage`
- `mediastore`
- `mediastoredata`
- `mq`
- `neptune`
- `opsworks`
- `organizations`
- `personalize`
- `pinpoint`
- `pricing`
- `qldb`
- `quicksight`
- `r53` (**DEPRECATED** Use `route53` instead)
- `ram`
- `rds`
- `redshift`
- `resourcegroups`
- `route53`
- `route53resolver`
- `s3`
- `s3control`
- `sagemaker`
- `sdb`
- `secretsmanager`
- `securityhub`
- `serverlessrepo`
- `servicecatalog`
- `servicediscovery`
- `servicequotas`
- `ses`
- `shield`
- `sns`
- `sqs`
- `ssm`
- `stepfunctions`
- `storagegateway`
- `sts`
- `swf`
- `transfer`
- `waf`
- `wafregional`
- `worklink`
- `workspaces`
- `xray`

</div>
<!-- markdownlint-enable MD033 -->

## Connecting to Local AWS Compatible Solutions

~> **NOTE:** This information is not intended to be exhaustive for all local AWS compatible solutions or necessarily authoritative configurations for those documented. Check the documentation for each of these solutions for the most up to date information.

### DynamoDB Local

The Amazon DynamoDB service offers a downloadable version for writing and testing applications without accessing the DynamoDB web service. For more information about this solution, see the [DynamoDB Local documentation in the Amazon DynamoDB Developer Guide](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.html).

An example provider configuration:

```hcl
provider "aws" {
  access_key                  = "mock_access_key"
  region                      = "us-east-1"
  secret_key                  = "mock_secret_key"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    dynamodb = "http://localhost:8000"
  }
}
```

### LocalStack

[LocalStack](https://localstack.cloud/) provides an easy-to-use test/mocking framework for developing Cloud applications.

An example provider configuration:

```hcl
provider "aws" {
  access_key                  = "mock_access_key"
  region                      = "us-east-1"
  s3_force_path_style         = true
  secret_key                  = "mock_secret_key"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    apigateway     = "http://localhost:4567"
    cloudformation = "http://localhost:4581"
    cloudwatch     = "http://localhost:4582"
    dynamodb       = "http://localhost:4569"
    es             = "http://localhost:4578"
    firehose       = "http://localhost:4573"
    iam            = "http://localhost:4593"
    kinesis        = "http://localhost:4568"
    lambda         = "http://localhost:4574"
    route53        = "http://localhost:4580"
    redshift       = "http://localhost:4577"
    s3             = "http://localhost:4572"
    secretsmanager = "http://localhost:4584"
    ses            = "http://localhost:4579"
    sns            = "http://localhost:4575"
    sqs            = "http://localhost:4576"
    ssm            = "http://localhost:4583"
    stepfunctions  = "http://localhost:4585"
    sts            = "http://localhost:4592"
  }
}
```
