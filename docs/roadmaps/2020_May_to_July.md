# Roadmap: May - July 2020

Each quarter the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top 10 Community Issues, [core services](../CORE_SERVICES.md), and internal priorities. When community pull requests exist for a given item, we will prioritize working with the original authors to include their contributions. If the author can no longer take on the implementation, HashiCorp will complete any additional work needed.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap.

To make contribution easier, we’ll be using the [`Help Wanted`](https://github.com/hashicorp/terraform-provider-aws/labels/help%20wanted) tag to point to issues we’d like to include in this quarter’s series of releases.

This quarter (May-July ‘20) we will be prioritizing the following areas of work:

### 3.0.0

Milestone: [v3.0.0](https://github.com/hashicorp/terraform-provider-aws/milestone/70)

Each year the TF AWS Provider team releases a major version. [Major releases](https://www.terraform.io/docs/extend/best-practices/versioning.html#example-major-number-increments) include code removals, deprecations, and breaking changes. A corresponding “upgrade guide” will be published alongside the release.

We'll be updating the linked milestone as we work to finalize and complete v3.0.0.

### WAFv2

Issue: [#11046](https://github.com/hashicorp/terraform-provider-aws/issues/11046)

_AWS WAFv2 is a web application firewall that lets you monitor the HTTP and HTTPS requests that are forwarded to Amazon CloudFront, an Amazon API Gateway API, or an Application Load Balancer._

Support for WAFv2 functionality will be wholly separate from WAF “Classic”. We’ll focus on enabling community contributions to WAFv2 first. If there is not a community contribution, HashiCorp will work to add the missing resource or data source.

Support for WAFv2 will include:

#### Resources

* aws_wafv2_ip_set
* aws_wafv2_regex_pattern_set
* aws_wafv2_rule_group
* aws_wafv2_web_acl
* aws_wafv2_web_acl_association

#### Data Sources

* aws_wafv2_ip_set
* aws_wafv2_regex_pattern_set
* aws_wafv2_rule_group
* aws_wafv2_web_acl

### Amazon Lex

Issue: [#905](https://github.com/hashicorp/terraform-provider-aws/issues/905)

_Amazon Lex is a service for building conversational interfaces into any application using voice and text. Amazon Lex provides the advanced deep learning functionalities of automatic speech recognition (ASR) for converting speech to text, and natural language understanding (NLU) to recognize the intent of the text, to enable you to build applications with highly engaging user experiences and lifelike conversational interactions._

We’ll focus on enabling community contributions to Lex first. If there is not a community contribution, HashiCorp will work to add the missing resource or data source.

Support for Amazon Lex will include:

#### Resources

* aws_lex_slot_type
* aws_lex_intent
* aws_lex_bot
* aws_lex_bot_alias

#### Data Sources

* aws_lex_slot_type
* aws_lex_intent
* aws_lex_bot
* aws_lex_bot_alias

### AWS Certificate Manager

Issue: [#8531](https://github.com/hashicorp/terraform-provider-aws/issues/8531)

_AWS Certificate Manager is a service that allows you to easily provision, manage, and deploy public and private Secure Sockets Layer/Transport Layer Security (SSL/TLS) certificates for use with AWS services and your internal connected resources._

After evaluating the issue linked above, we concluded that the ACM resource was in need of a redesign. We’ll be prioritizing redesigning and updating the resource while we tackle the open bug reports and enhancements. Our research and redesign work will be tracked [here](https://github.com/hashicorp/terraform-provider-aws/issues/13053).

## Research Topics

Research topics include features, architectural changes, and ideas that we are pursuing in the longer term that may significantly impact the core user experience of the AWS provider. Research topics are discovery only and are not guaranteed to be included in a future release.

### Global Default Tags

Issue: [#7926](https://github.com/hashicorp/terraform-provider-aws/issues/7926)

We’ve been evaluating how users approach tagging their infrastructure in Terraform and the systems and practices that may interact with TF when it comes to tagging. The [initial discussions](https://github.com/hashicorp/terraform/issues/20866) led us to prioritize functionality that allows users to ignore specific tags globally in the AWS provider. As a complement to that feature, we are exploring the ability to supply global default tags to resources defined by the AWS Provider.

We are interested in your thoughts and feedback about this proposal and encourage you to comment on the issue linked above or schedule time with @maryelizbeth via the link on her [GitHub profile](https://github.com/maryelizbeth) to discuss.

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
