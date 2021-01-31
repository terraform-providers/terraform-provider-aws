# Roadmap:  November 2020 - January 2021

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top 10 Community Issues, [Core Services](docs/CORE_SERVICES.md), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors, are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. To view all the items we've prioritized for this quarter, please see the [Roadmap milestone](https://github.com/hashicorp/terraform-provider-aws/milestone/138).

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur .

From [August through October](docs/roadmaps/2020_August_to_October.md), we committed to adding support for EventBridge, ImageBuilder , LakeFormation and Serverless Application Repository as new service offerings. We were able to deliver EventBridge within that time frame. Unfortunately for a number of reasons we weren’t able to release ImageBuilder, LakeFormation and Serverless Application Repository. That said, they are in progress and on track for release in early November.

From October-January ‘21, we will be prioritizing the following areas of work:

## New Services

### AWS SSO Permission Sets
Issue: [#15108](https://github.com/hashicorp/terraform-provider-aws/issues/15108)

_[AWS SSO](https://docs.aws.amazon.com/singlesignon/latest/APIReference/welcome.html) account assignment APIs enable you to build automation to create and update permissions that align with your company's common job functions. You can then assign the permissions to users and groups to entitle them for access in their required accounts. For example, you can give your developers broad control over resources in developer accounts, and limit that control to authorized operations personnel in production accounts. The new AWS CloudFormation support enables you to automate account assignments as you build new accounts. You can also use the APIs to decode user and group names from the unique identifiers that appear in AWS CloudTrail logs._

Support for AWS SSO Permission Sets will include:

New Resource(s):
- aws_sso_permission_set
- aws_sso_permission_set_policy
- aws_sso_permission_set_policy_attachment
- aws_sso_account_assignment

## Issues & Enhancements

### Core Service Reliability
Core Services are areas of high usage or strategic importance for our users. We strive to offer rock solid reliability in these areas. This quarter we will have a focus on RDS and Elasticache (which we are also promoting to Core Service status) to address some common pain points in their usage and ensure they continue to meet our standards.

#### RDS

- [#15177](https://github.com/hashicorp/terraform-provider-aws/issues/15177): Subsequent plan/apply forces global cluster recreation when source cluster's storage_encrypted=true
- [#15583](https://github.com/hashicorp/terraform-provider-aws/issues/15583):  aws db parameter group ... converts keys and values to lowercase and fails 'apply' due to aws_db_parameter_group changes
- [#1198](https://github.com/hashicorp/terraform-provider-aws/issues/1198): Unable to ignore changes to RDS minor engine version
- [#9401](https://github.com/hashicorp/terraform-provider-aws/issues/9401): Destroy/recreate DB instance on minor version update rather than updating
- [#2635](https://github.com/hashicorp/terraform-provider-aws/issues/2635): RDS - storage_encrypted = true does not work
- [#467](https://github.com/hashicorp/terraform-provider-aws/issues/467): With aws_db_instance when you remove the snapshot_identifier it wants to force a new resource
- [#10197](https://github.com/hashicorp/terraform-provider-aws/issues/10197): AWS aurora unexpected state 'configuring-iam-database-auth' when modifying the `iam_database_authentication_enabled` flag
- [#13891](https://github.com/hashicorp/terraform-provider-aws/issues/13891): RDS Cluster is not reattached to Global Cluster after failing deletion

#### Elasticache
The Elasticache work will begin with a research spike to ensure that the we can solve the following issues without introducing breaking changes into the provider:  

- [#14959](https://github.com/hashicorp/terraform-provider-aws/issues/14959): Research Spike: Elasticache Service Fixes and Improvements
- [#12708](https://github.com/hashicorp/terraform-provider-aws/issues/12708): resource/aws_elasticache_replication_group: Add MultiAZ support
- ~[#13517](https://github.com/hashicorp/terraform-provider-aws/issues/13517): Feature Request: `aws_elasticache_cluster` allow auto-minor-version-upgrade to be set~ This parameter is not enabled in the AWS API.
- [#5118](https://github.com/hashicorp/terraform-provider-aws/issues/5118): support setting primary/replica AZ attributes inside NodeGroupConfiguration for RedisClusterModelEnabled

### Workflow Improvements

We’ll also be tackling some of the top reported issues in the provider that are causing disruptions to high priority workflows: 

- [#14373](https://github.com/hashicorp/terraform-provider-aws/issues/14373): cloudfront: support for cache and origin request policies
- [#11584](https://github.com/hashicorp/terraform-provider-aws/issues/11584): Add ability to manage VPN tunnel options
- [#13986](https://github.com/hashicorp/terraform-provider-aws/issues/13986): Feature request: Managed prefix lists
- [#8009](https://github.com/hashicorp/terraform-provider-aws/issues/8009): S3 settings on aws_dms_endpoint conflict with "extra_connection_attributes"
- [#11220](https://github.com/hashicorp/terraform-provider-aws/issues/11220): Set account recovery preference
- [#12272](https://github.com/hashicorp/terraform-provider-aws/issues/12272): CloudWatch composite alarms
- [#4058](https://github.com/hashicorp/terraform-provider-aws/issues/4058): Support Firewall Manager Policies
- [#10931](https://github.com/hashicorp/terraform-provider-aws/issues/10931): Resource aws_sns_topic_subscription, new argument redrive_policy
- [#11098](https://github.com/hashicorp/terraform-provider-aws/issues/11098): Support for AWS Config Conformance Packs
- [#6674](https://github.com/hashicorp/terraform-provider-aws/issues/6674): Feature Request: Security Hub
- [#3891](https://github.com/hashicorp/terraform-provider-aws/issues/3891): Adding custom cognito user pool attribute forces new resource
- [#2245](https://github.com/hashicorp/terraform-provider-aws/issues/2245): AWS security groups not being destroyed
- [#8114](https://github.com/hashicorp/terraform-provider-aws/issues/8114): Cognito User Pool UI Customization
- [#11348](https://github.com/hashicorp/terraform-provider-aws/issues/11348): Add Type to AWS SFN State Machine
- [#11586](https://github.com/hashicorp/terraform-provider-aws/issues/11586): Faulty Read of Client VPN Network associations break state

### Technical Debt Theme

Last quarter we made considerable progress in improving the stability of our Acceptance Test suite. We were able to reduce our consistent test failures by 50% in Commercial, and fixed hundreds of tests in GovCloud. We believe that keeping our focus in this area in the next quarter is the way forward that provides the most value. With another quarter of focus we are looking to have a test suite free of problematic tests, along with optimizations which should improve the speeds of the suite.

### Research Topics

Research topics include features, architectural changes, and ideas that we are pursuing in the longer term that may significantly impact the core user experience of the AWS provider. Research topics are discovery only and are not guaranteed to be included in a future release.

We are interested in your thoughts and feedback about the proposals below and encourage you to comment on the linked issues or schedule time with @maryelizbeth via the link on her GitHub profile to discuss.

#### Default Tags Implementation Design
Issue: [#7926](https://github.com/hashicorp/terraform-provider-aws/issues/7926)

After completing user research and an internal review of our research conclusions, we will begin conducting engineering research and publish an RFC to address the implementation of this feature. Once the RFC has been approved, we will update the community with our plans for Default Tags. 

#### API Calls/IAM Actions Per Terraform Resource (Minimum IAM)
Issue: [#9154](https://github.com/hashicorp/terraform-provider-aws/issues/9154)

To address security concerns and best practices we are considering how Terraform could surface minimally viable IAM policies for taking actions on resources or executing a TF plan. This is in the early stages of research and we are particularly interested in whether or not this would be useful and the resources or services areas for which it is most valuable.

#### Lifecycle: Retain [Add 'retain' attribute to the Terraform lifecycle meta-parameter]
Issue: [#902](https://github.com/hashicorp/terraform-provider-aws/issues/902)

Some resources (e.g. log groups) are intended to be created but never destroyed. Terraform currently does not have a lifecycle attribute for retaining such resources. We are curious as to whether or not retaining resources is a workflow that meets the needs of our community and if so, how and where we might make use of that in the AWS Provider.

### Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
