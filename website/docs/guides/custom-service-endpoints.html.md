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
<ul>
  <li><code>accessanalyzer</code></li>
  <li><code>acm</code></li>
  <li><code>acmpca</code></li>
  <li><code>amplify</code></li>
  <li><code>apigateway</code></li>
  <li><code>applicationautoscaling</code></li>
  <li><code>applicationinsights</code></li>
  <li><code>appmesh</code></li>
  <li><code>appstream</code></li>
  <li><code>appsync</code></li>
  <li><code>athena</code></li>
  <li><code>autoscaling</code></li>
  <li><code>autoscalingplans</code></li>
  <li><code>backup</code></li>
  <li><code>batch</code></li>
  <li><code>budgets</code></li>
  <li><code>cloud9</code></li>
  <li><code>cloudformation</code></li>
  <li><code>cloudfront</code></li>
  <li><code>cloudhsm</code></li>
  <li><code>cloudsearch</code></li>
  <li><code>cloudtrail</code></li>
  <li><code>cloudwatch</code></li>
  <li><code>cloudwatchevents</code></li>
  <li><code>cloudwatchlogs</code></li>
  <li><code>codeartifact</code></li>
  <li><code>codebuild</code></li>
  <li><code>codecommit</code></li>
  <li><code>codedeploy</code></li>
  <li><code>codepipeline</code></li>
  <li><code>codestarconnections</code></li>
  <li><code>codestarnotifications</code></li>
  <li><code>cognitoidentity</code></li>
  <li><code>cognitoidp</code></li>
  <li><code>configservice</code></li>
  <li><code>connect</code></li>
  <li><code>cur</code></li>
  <li><code>dataexchange</code></li>
  <li><code>datapipeline</code></li>
  <li><code>datasync</code></li>
  <li><code>dax</code></li>
  <li><code>devicefarm</code></li>
  <li><code>directconnect</code></li>
  <li><code>dlm</code></li>
  <li><code>dms</code></li>
  <li><code>docdb</code></li>
  <li><code>ds</code></li>
  <li><code>dynamodb</code></li>
  <li><code>ec2</code></li>
  <li><code>ecr</code></li>
  <li><code>ecrpublic</code></li>
  <li><code>ecs</code></li>
  <li><code>efs</code></li>
  <li><code>eks</code></li>
  <li><code>elasticache</code></li>
  <li><code>elasticbeanstalk</code></li>
  <li><code>elastictranscoder</code></li>
  <li><code>elb</code></li>
  <li><code>emr</code></li>
  <li><code>es</code></li>
  <li><code>firehose</code></li>
  <li><code>fms</code></li>
  <li><code>forecast</code></li>
  <li><code>fsx</code></li>
  <li><code>gamelift</code></li>
  <li><code>glacier</code></li>
  <li><code>globalaccelerator</code></li>
  <li><code>glue</code></li>
  <li><code>guardduty</code></li>
  <li><code>greengrass</code></li>
  <li><code>iam</code></li>
  <li><code>identitystore</code></li>
  <li><code>imagebuilder</code></li>
  <li><code>inspector</code></li>
  <li><code>iot</code></li>
  <li><code>iotanalytics</code></li>
  <li><code>iotevents</code></li>
  <li><code>kafka</code></li>
  <li><code>kinesis</code></li>
  <li><code>kinesisanalytics</code></li>
  <li><code>kinesisanalyticsv2</code></li>
  <li><code>kinesisvideo</code></li>
  <li><code>kms</code></li>
  <li><code>lakeformation</code></li>
  <li><code>lambda</code></li>
  <li><code>lexmodels</code></li>
  <li><code>licensemanager</code></li>
  <li><code>lightsail</code></li>
  <li><code>macie</code></li>
  <li><code>macie2</code></li>
  <li><code>managedblockchain</code></li>
  <li><code>marketplacecatalog</code></li>
  <li><code>mediaconnect</code></li>
  <li><code>mediaconvert</code></li>
  <li><code>medialive</code></li>
  <li><code>mediapackage</code></li>
  <li><code>mediastore</code></li>
  <li><code>mediastoredata</code></li>
  <li><code>mq</code></li>
  <li><code>mwaa</code></li>
  <li><code>neptune</code></li>
  <li><code>networkfirewall</code></li>
  <li><code>networkmanager</code></li>
  <li><code>opsworks</code></li>
  <li><code>organizations</code></li>
  <li><code>outposts</code></li>
  <li><code>personalize</code></li>
  <li><code>pinpoint</code></li>
  <li><code>pricing</code></li>
  <li><code>qldb</code></li>
  <li><code>quicksight</code></li>
  <li><code>ram</code></li>
  <li><code>rds</code></li>
  <li><code>redshift</code></li>
  <li><code>resourcegroups</code></li>
  <li><code>resourcegroupstaggingapi</code></li>  
  <li><code>route53</code></li>
  <li><code>route53domains</code></li>
  <li><code>route53resolver</code></li>
  <li><code>s3</code></li>
  <li><code>s3control</code></li>
  <li><code>s3outposts</code></li>
  <li><code>sagemaker</code></li>
  <li><code>sdb</code></li>
  <li><code>secretsmanager</code></li>
  <li><code>securityhub</code></li>
  <li><code>serverlessrepo</code></li>
  <li><code>servicecatalog</code></li>
  <li><code>servicediscovery</code></li>
  <li><code>servicequotas</code></li>
  <li><code>ses</code></li>
  <li><code>shield</code></li>
  <li><code>signer</code></li>
  <li><code>sns</code></li>
  <li><code>sqs</code></li>
  <li><code>ssm</code></li>
  <li><code>ssoadmin</code></li>
  <li><code>stepfunctions</code></li>
  <li><code>storagegateway</code></li>
  <li><code>sts</code></li>
  <li><code>swf</code></li>
  <li><code>synthetics</code></li>
  <li><code>timestreamwrite</code></li>
  <li><code>transfer</code></li>
  <li><code>waf</code></li>
  <li><code>wafregional</code></li>
  <li><code>wafv2</code></li>
  <li><code>worklink</code></li>
  <li><code>workmail</code></li>
  <li><code>workspaces</code></li>
  <li><code>xray</code></li>
</ul>
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
