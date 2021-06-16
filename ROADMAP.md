# Roadmap:  May 2021 - July 2021

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top 10 Community Issues, [Core Services](docs/CORE_SERVICES.md), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. To view all the items we've prioritized for this quarter, please see the [Roadmap milestone](https://github.com/hashicorp/terraform-provider-aws/milestone/138).

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur .

In the period spanning February to April 2021 846 Pull Requests were opened in the provider and 947 were merged, adding support for:

- Default Tags
- CloudFront Origin Request and Cache Policies
- Cloudwatch Synthetics
- SecurityHub
- Elasticache Global Replication Group
- ACM Private Certificate Authorities
- Managed Workflows for Apache Airflow
- Managed Add Ons for EKS
- ECR Cross Region Replication
- SNS FIFO Topics
- EC2 Autoscaling Warm Pools

Among many other enhancements, bug fixes and resolutions to technical debt items.

From May-July ‘21, we will be prioritizing the following areas of work:

## New Services

### AWS Amplify
Issue: [#6917](https://github.com/hashicorp/terraform-provider-aws/issues/6917)

_AWS Amplify is a set of tools and services that can be used together or on their own, to help front-end web and mobile developers build scalable full stack applications, powered by AWS. With Amplify, you can configure app backends and connect your app in minutes, deploy static web apps in a few clicks, and easily manage app content outside the AWS console._
Support for AWS Amplify will include:

New Resource(s):
- aws_amplify_app
- aws_amplify_backend_environment
- aws_amplify_branch
- aws_amplify_domain_association
- aws_amplify_webhook

### Amazon Timestream

Issue: [#15421](https://github.com/hashicorp/terraform-provider-aws/issues/15421)

_Amazon Timestream is a fast, scalable, and serverless time series database service for IoT and operational applications that makes it easy to store and analyze trillions of events per day up to 1,000 times faster and at as little as 1/10th the cost of relational databases. Amazon Timestream saves you time and cost in managing the lifecycle of time series data, and its purpose-built query engine lets you access and analyze recent and historical data together with a single query. Amazon Timestream has built-in time series analytics functions, helping you identify trends and patterns in near real-time. Amazon Timestream is serverless and automatically scales up or down to adjust capacity and performance, so you don’t need to manage the underlying infrastructure, freeing you to focus on building your applications._

Support for Amazon Timestream will include:

New Resource(s):

- aws_timestreamwrite_database

### AWS AppConfig

Issue: [#11973](https://github.com/hashicorp/terraform-provider-aws/issues/11973)

_Use AWS AppConfig, a capability of AWS Systems Manager, to create, manage, and quickly deploy application configurations. You can use AWS AppConfig with applications hosted on Amazon Elastic Compute Cloud (Amazon EC2) instances, AWS Lambda, containers, mobile applications, or IoT devices._

Support for AWS AppConfig will include:

New Resource(s)
- aws_appconfig_application
- aws_appconfig_configuration_profile
- aws_appconfig_deployment_strategy
- aws_appconfig_environment
- aws_appconfig_deployment


## Enhancements to Existing Services
- [AWS Transfer Server: Attach VPC security groups at creation](https://github.com/hashicorp/terraform-provider-aws/issues/15788)
- [EC2 Launch Templates](https://github.com/hashicorp/terraform-provider-aws/issues/4264)
- [AWS ElasticSearch Native SAML for Kibana](https://github.com/hashicorp/terraform-provider-aws/issues/16259)

## Core Service Reliability

Core Services are areas of high usage or strategic importance for our users. We strive to offer rock solid reliability in these areas. This quarter we will have a focus on S3. We will be preparing a wholesale re-design of the `aws_s3_bucket` resource that we are planning to introduce in our major version release (v4.0) this year. Our focus will be on understanding how we can better break up the currently monolithic S3 bucket resource.

## Research Topics

Research topics include features, architectural changes, and ideas that we are pursuing in the longer term that may significantly impact the core user experience of the AWS provider. Research topics are discovery only and are not guaranteed to be included in a future release.

We are interested in your thoughts and feedback about the proposals below and encourage you to comment on the linked issues or schedule time with @maryelizbeth via the link on her GitHub profile to discuss.

- Major Version Planning (v4.0) including wholesale redesign of the aws_s3_bucket resource to break it up into more manageable resources.
- AWS Go SDK v2 Adoption
- Test Discovery - Enable the automation of running the correct subset of acceptance tests for a given PR.

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
