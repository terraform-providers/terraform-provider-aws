# Roadmap: August - October 2020

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top 10 Community Issues, [core services](../CORE_SERVICES.md), and internal priorities. When community pull requests exist for a given item, we will prioritize working with the original authors to include their contributions. If the author can no longer take on the implementation, HashiCorp will complete any additional work needed.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. To view all the items we've prioritized for this quarter, please see the [Roadmap milestone](https://github.com/hashicorp/terraform-provider-aws/milestone/138).

Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist we will create the resources and implementation ourselves.

From [May through July](2020_May_to_July.md), we committed to adding support for WAFv2 and Lex. We are happy to report that WAFv2 support is now complete and we will be finishing support for Lex immediately following the release of v3.0.0. Additionally, we planned to refactor ACM and will include the redesigned resource as part of the 3.0.0 package. Lastly, we spent some time researching Default Tags and want to extend a thank you to all the folks that volunteered to assist. We’ve written a proposal for the feature that is being reviewed internally and we hope to include the functionality in the product in the future.

From August-October ‘20, we will be prioritizing the following areas of work:

## New Services

### EventBridge

Issue: [#9330](https://github.com/hashicorp/terraform-provider-aws/issues/9330)

_[Amazon EventBridge](https://aws.amazon.com/eventbridge/) is a serverless event bus that makes it easy to connect applications together using data from your own applications, integrated Software-as-a-Service (SaaS) applications, and AWS services. EventBridge delivers a stream of real-time data from event sources, such as Zendesk, Datadog, or Pagerduty, and routes that data to targets like AWS Lambda. You can set up routing rules to determine where to send your data to build application architectures that react in real time to all of your data sources._

As EventBridge exists as an addition to existing CloudWatch APIs we will perform a research phase to determine whether Terraform support should exist as separate resources, additions to existing ones, or a combination of the two.

### EC2 Image Builder

Issue: [#11084](https://github.com/hashicorp/terraform-provider-aws/issues/11084)

_[EC2 Image Builder](https://aws.amazon.com/image-builder/) simplifies the creation, maintenance, validation, sharing, and deployment of Linux or Windows Server images for use with Amazon EC2 and on-premises._

Support for EC2 Image Builder will include:

New Resource(s):

- aws_imagebuilder_component
- aws_imagebuilder_distribution_configuration
- aws_imagebuilder_image
- aws_imagebuilder_image_pipeline
- aws_imagebuilder_image_recipe
- aws_imagebuilder_infrastructure_configuration

New Data Source(s):

- aws_imagebuilder_image

### AWS Lake Formation

Issue: [#9700](https://github.com/hashicorp/terraform-provider-aws/issues/9700)

_[AWS Lake Formation](https://aws.amazon.com/lake-formation) is a service that makes it easy to set up a secure data lake in days. A data lake is a centralized, curated, and secured repository that stores all your data, both in its original form and prepared for analysis. A data lake enables you to break down data silos and combine different types of analytics to gain insights and guide better business decisions._

Support for AWS Lake Formation will include:

New Resource(s):

- aws_lakeformation_resource
- aws_lakeformation_data_lake_settings
- aws_lakeformation_permissions

### AWS Serverless Application Repository

Issue: [#3981](https://github.com/hashicorp/terraform-provider-aws/issues/3981)

_The [AWS Serverless Application Repository](https://aws.amazon.com/serverless/serverlessrepo/) is a managed repository for serverless applications. It enables teams, organizations, and individual developers to store and share reusable applications, and easily assemble and deploy serverless architectures in powerful new ways. Using the Serverless Application Repository, you don't need to clone, build, package, or publish source code to AWS before deploying it. Instead, you can use pre-built applications from the Serverless Application Repository in your serverless architectures, helping you and your teams reduce duplicated work, ensure organizational best practices, and get to market faster. Integration with AWS Identity and Access Management (IAM) provides resource-level control of each application, enabling you to publicly share applications with everyone or privately share them with specific AWS accounts._

Support for AWS Serverless Application Repository will include:

New Resource(s):

- aws_serverlessapplicationrepository_cloudformation_stack

New Data Source(s):

- aws_serverlessapplicationrepository_application

## Issues and Enhancements

The issues below have gained substantial support via our community. As a result, we want to highlight our commitment to addressing them.

- [#12690](https://github.com/hashicorp/terraform-provider-aws/issues/12690) RDS Proxy Support
- [#11281](https://github.com/hashicorp/terraform-provider-aws/issues/11281) Home Directory Mappings Support for AWS Transfer User
- [#384](https://github.com/hashicorp/terraform-provider-aws/issues/384) Add support for CreateVPCAssociationAuthorization AWS API
- [#6562](https://github.com/hashicorp/terraform-provider-aws/issues/6562) Auto Scaling Plans (Dynamic/Predictive Auto Scaling Groups)
- [#5549](https://github.com/hashicorp/terraform-provider-aws/issues/5549) Terraform constantly updates resource policy on API Gateway
- [#11569](https://github.com/hashicorp/terraform-provider-aws/issues/11569) aws_transfer_server: support Elastic IPs
- [#5286](https://github.com/hashicorp/terraform-provider-aws/issues/5286) Point in time restore support for AWS RDS instances

## United States Federal Focus

We have added extra engineering and product capacity to enable us to provide the same compatibility and coverage assurances in the GovCloud, C2S, and SC2S regions as we currently do for Commercial AWS regions. Our attention on C2S/SC2S environments should result in better outcomes in other similar air gapped environments. Initially, we will be focusing on GovCloud and users should expect improved experiences within that region in the coming months.

## Technical Debt Theme

Each quarter we identify a technical debt theme for the team to focus on alongside new service additions, issue resolutions and enhancements. This quarter we are looking at spending time improving the reliability of our acceptance test framework. We have a number of flaky tests which add friction to the development cycle. Making these more consistent should improve the development experience for both contributors and maintainers.

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
