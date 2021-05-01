# Roadmap:  February 2021 - April 2021

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top 10 Community Issues, [Core Services](docs/CORE_SERVICES.md), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors, are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. To view all the items we've prioritized for this quarter, please see the [Roadmap milestone](https://github.com/hashicorp/terraform-provider-aws/milestone/138).

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur .

From [November through January](docs/roadmaps/2020_November_to_January.md), we added support for (among other things):

- SSO Permission Sets
- EC2 Managed Prefix Lists
- Firewall Manager Policies
- SASL/SCRAM Authentication for MSK
- ImageBuilder
- LakeFormation
- Serverless Application Repository
- Cloudwatch Composite Alarms

As well as partnering with AWS to provide launch day support for:

- Network Firewall
- Code Signing for Lambda
- Container Images for Lambda
- Gateway Load Balancer
- Spot Launch for EKS Managed Node Groups

From February-April ‘21, we will be prioritizing the following areas of work:

## Provider Functionality: Default Tags

Issue: [#7926](https://github.com/hashicorp/terraform-provider-aws/issues/7926)

Default Tags builds on the workflows in Ignore Tags to provide additional control over the ways Terraform manages tagging capabilities. Users will be able to specify lists of tags to apply to all resources in a configuration at the provider level. Our goal in offering this use case is to assist in tidying up configuration files, decreasing development efforts, and streamlining cost allocation and resource attribution within organizations of all sizes. 

## New Services

### CloudWatch Synthetics
Issue: [#11145](https://github.com/hashicorp/terraform-provider-aws/issues/11145)

_[CloudWatch Synthetics](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Synthetics_Canaries.html) You can use Amazon CloudWatch Synthetics to create canaries, configurable scripts that run on a schedule, to monitor your endpoints and APIs. Canaries follow the same routes and perform the same actions as a customer, which makes it possible for you to continually verify your customer experience even when you don't have any customer traffic on your applications. By using canaries, you can discover issues before your customers do._

Support for CloudWatch Synthetics will include:

New Resource(s):
- aws_synthetics_canary

New Datasource(s):
- aws_synthetics_canary_runs

### Managed Workflows for Apache Airflow

Issue: [#16432](https://github.com/hashicorp/terraform-provider-aws/issues/16432)

_[Managed Workflows for Apache Airflow](https://aws.amazon.com/blogs/aws/introducing-amazon-managed-workflows-for-apache-airflow-mwaa/) Amazon Managed Workflows for Apache Airflow (MWAA) is a managed orchestration service for Apache Airflow1 that makes it easier to set up and operate end-to-end data pipelines in the cloud at scale. Apache Airflow is an open-source tool used to programmatically author, schedule, and monitor sequences of processes and tasks referred to as “workflows.” With Managed Workflows, you can use Airflow and Python to create workflows without having to manage the underlying infrastructure for scalability, availability, and security. Managed Workflows automatically scales its workflow execution capacity to meet your needs, and is integrated with AWS security services to help provide you with fast and secure access to data._

Support for Amazon Managed Workflows for Apache Airflow will include:

New Resource(s):

- aws_mwaa_environment

## Core Service Reliability
Core Services are areas of high usage or strategic importance for our users. We strive to offer rock solid reliability in these areas. This quarter we will have a focus on RDS and Elasticache (which we are also promoting to Core Service status) to address some common pain points in their usage and ensure they continue to meet our standards.

### RDS

- [#15177](https://github.com/hashicorp/terraform-provider-aws/issues/15177): Subsequent plan/apply forces global cluster recreation when source cluster's storage_encrypted=true
- [#15583](https://github.com/hashicorp/terraform-provider-aws/issues/15583):  aws db parameter group ... converts keys and values to lowercase and fails 'apply' due to aws_db_parameter_group changes
- [#1198](https://github.com/hashicorp/terraform-provider-aws/issues/1198): Unable to ignore changes to RDS minor engine version
- [#9401](https://github.com/hashicorp/terraform-provider-aws/issues/9401): Destroy/recreate DB instance on minor version update rather than updating
- [#2635](https://github.com/hashicorp/terraform-provider-aws/issues/2635): RDS - storage_encrypted = true does not work
- [#467](https://github.com/hashicorp/terraform-provider-aws/issues/467): With aws_db_instance when you remove the snapshot_identifier it wants to force a new resource
- [#10197](https://github.com/hashicorp/terraform-provider-aws/issues/10197): AWS aurora unexpected state 'configuring-iam-database-auth' when modifying the `iam_database_authentication_enabled` flag
- [#13891](https://github.com/hashicorp/terraform-provider-aws/issues/13891): RDS Cluster is not reattached to Global Cluster after failing deletion

## Technical Debt Theme

Last quarter we continued to improve the stability of our Acceptance Test suite. Following on from that work we will begin to integrate our Pull Request workflow with our Acceptance testing suite with a goal of being able to determine which tests to run, trigger, and view results of Acceptance Test runs on GitHub. This will improve our time to merge incoming PR's and further protect against regressions.

We also spent time last quarter improving our documentation to give contributors more explicit guidance on best practice patterns for [data conversion](https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/data-handling-and-conversion.md) and [error handling](https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/error-handling.md).  

## Research Topics

Research topics include features, architectural changes, and ideas that we are pursuing in the longer term that may significantly impact the core user experience of the AWS provider. Research topics are discovery only and are not guaranteed to be included in a future release.

We are interested in your thoughts and feedback about the proposals below and encourage you to comment on the linked issues or schedule time with @maryelizbeth via the link on her GitHub profile to discuss.

### API Calls/IAM Actions Per Terraform Resource (Minimum IAM)
Issue: [#9154](https://github.com/hashicorp/terraform-provider-aws/issues/9154)

To address security concerns and best practices we are considering how Terraform could surface minimally viable IAM policies for taking actions on resources or executing a TF plan. This is in the early stages of research and we are particularly interested in whether or not this would be useful and the resources or services areas for which it is most valuable.

### Lifecycle: Retain [Add 'retain' attribute to the Terraform lifecycle meta-parameter]
Issue: [#902](https://github.com/hashicorp/terraform-provider-aws/issues/902)

Some resources (e.g. log groups) are intended to be created but never destroyed. Terraform currently does not have a lifecycle attribute for retaining such resources. We are curious as to whether or not retaining resources is a workflow that meets the needs of our community and if so, how and where we might make use of that in the AWS Provider.

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
