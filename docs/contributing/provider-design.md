# Provider Design

_Please Note: This documentation is intended for Terraform AWS Provider code developers. Typical operators writing and applying Terraform configurations do not need to read or understand this material._

The Terraform AWS Provider follows the guidelines established in the [HashiCorp Provider Design Principles](https://www.terraform.io/docs/extend/hashicorp-provider-design-principles.html). That general documentation provides many high-level design points gleaned from years of experience with Terraform's design and implementation concepts. Sections below will expand on specific design details between that documentation and this provider, while others will capture other pertinent information that may not be covered there. Other pages of the contributing guide cover implementation details such as code, testing, and documentation specifics.

- [API and SDK Boundary](#api-and-sdk-boundary)
- [Infrastructure as Code Suitability](#infrastructure-as-code-suitability)
- [Resource Type Considerations](#resource-type-considerations)
    - [Authorization and Acceptance Resources](#authorization-and-acceptance-resources)
    - [Cross-Service Functionality](#cross-service-functionality)
    - [Data Sources](#data-sources)
        - [Plural Data Sources](#plural-data-sources)
        - [Singular Data Sources](#singular-data-sources)
    - [IAM Resource-Based Policy Resources](#iam-resource-based-policy-resources)
    - [Managing Resource Running State](#managing-resource-running-state)
    - [Task Execution and Waiter Resources](#task-execution-and-waiter-resources)
    - [Versioned Resources](#versioned-resources)
- [Other Considerations](#other-considerations)
    - [AWS Credential Exfiltration](#aws-credential-exfiltration)

## API and SDK Boundary

The AWS provider implements support for the [AWS](https://aws.amazon.com/) service APIs using the [AWS Go SDK](https://aws.amazon.com/sdk-for-go/). The API and SDK limits extend to the provider. In general, SDK operations manage the lifecycle of AWS components, such as creating, describing, updating, and deleting a database. Operations do not usually handle functionality within those components, such as executing a query on a database. If you are interested in other APIs/SDKs, we invite you to view the many Terraform Providers available, as each has a community of domain expertise.

Some examples of functionality that is not expected in this provider:

* Raw HTTP(S) handling. See the [Terraform HTTP Provider](https://registry.terraform.io/providers/hashicorp/http/latest) and [Terraform TLS Provider](https://registry.terraform.io/providers/hashicorp/tls/latest) instead.
* Kubernetes resource management beyond the EKS service APIs. See the [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest) instead.
* Active Directory or other protocol clients. See the [Terraform Active Directory Provider](https://registry.terraform.io/providers/hashicorp/ad/latest/docs) and other available provider instead.
* Functionality that requires additional software beyond the Terraform AWS Provider to be installed on the host executing Terraform. This currently includes the AWS CLI. See the [Terraform External Provider](https://registry.terraform.io/providers/hashicorp/external/latest) and other available providers instead.

## Infrastructure as Code Suitability

The provider maintainers' design goal is to cover as much of the AWS API as pragmatically possible. However, not every aspect of the API is compatible with an infrastructure-as-code (IaC) conception. If such limits affect you, we recommend that you open an AWS Support case and encourage others to do the same. Request that AWS components be made more self-contained and compatible with IaC. These AWS Support cases can also yield insights into the AWS service and API that are not well documented.

## Resource Type Considerations

Terraform resources work best as the smallest infrastructure blocks on which practitioners can build more complex configurations and abstractions, such as [Terraform Modules](https://www.terraform.io/docs/modules/). The general heuristic guiding when to implement a new Terraform resource for an aspect of AWS is whether the AWS service API provides create, read, update, and delete (CRUD) operations. However, not all AWS service API functionality falls cleanly into CRUD lifecycle management. In these situations, there is extra consideration necessary for properly mapping API operations to Terraform resources.

This section highlights design patterns when to consider an implementation within a singular Terraform resource or as separate Terraform resources.

Please note: the overall design and implementation across all AWS functionality is federated: individual services may implement concepts and use terminology differently. As such, this guide is not exhaustive. The aim is to provide general concepts and basic terminology that points contributors in the right direction, especially in understanding previous implementations.

### Authorization and Acceptance Resources

Some AWS services use an authorization-acceptance model for cross-account associations or access. Examples include:

* Direct Connect Association Proposals
* GuardDuty Member Invitations
* RAM Resource Share Associations
* Route 53 VPC Associations
* Security Hub Member Invitations

Depending on the API and components, AWS uses two basic ways of creating cross-region and cross-account associations. One way is to generate an invitation (or proposal) identifier from one AWS account to another. Then in the other AWS account, that identifier is used to accept the invitation. The second way is configuring a reference to another AWS account identifier. These may not require explicit acceptance on the receiving account to finish creating the association or begin working.

To model creating an association using an invitation or proposal, follow these guidelines.

* Follow the naming in the AWS service API to determine whether to use the term "invitation" or "proposal."
* For the originating account, create an "invitation" or "proposal" resource. Make sure that the AWS service API has operations for creating and reading invitations.
* For the responding account, create an "accepter" resource. Ensure that the API has operations for accepting, reading, and rejecting invitations in the responding account. Map the operations as follows:
    * Create: Accepts the invitation.
    * Read: Reads the invitation to determine its status. Note that in some APIs, invitations expire and disappear, complicating associations. If a resource does not find an invitation, the developer should implement a fall back to read the API resource associated with the invitation/proposal.
    * Delete: Rejects or otherwise deletes the invitation.

To model the second type of association, implicit associations, create an "association" resource and, optionally, an "authorization" resource. Map create, read, and delete to the corresponding operations in the AWS service API.

### Cross-Service Functionality

Many AWS service APIs build on top of other AWS services. Some examples of these include:

* EKS Node Groups managing Auto Scaling Groups
* Lambda Functions managing EC2 ENIs
* Transfer Servers managing EC2 VPC Endpoints

Some cross-service API implementations lack the management or description capabilities of the other service. The lack can make the Terraform resource implementation seem incomplete or unsuccessful in end-to-end configurations. Given the overall “resources should represent a single API object” goal from the [HashiCorp Provider Design Principles](https://www.terraform.io/docs/extend/hashicorp-provider-design-principles.html), a resource must only communicate with a single AWS service API. As such, maintainers will not approve cross-service resources.

The rationale behind this design decision includes the following:

* Unexpected IAM permissions being necessary for the resource. In high-security environments, all the service permissions may not be available or acceptable.
* Unexpected services generating CloudTrail logs for the resource.
* Needing extra and unexpected API endpoints configuration for organizations using custom endpoints, such as VPC endpoints.
* Unexpected changes to the AWS service internals for the cross-service implementations. Given that this functionality is not part of the primary service API, these details can change over time and may not be considered as a breaking change by the service team for an API upgrade.

A poignant real-world example of the last point involved a Lambda resource. The resource helped clean up extra resources (ENIs) due to a common misconfiguration. Practitioners found the functionality helpful since the issue was hard to diagnose. Years later, AWS updated the Lambda API. Immediately, practitioners reported that Terraform executions were failing. Downgrading the provider was not possible since many configurations depended on recent releases. For environments running many versions behind, forcing an upgrade with the fix would likely cause unrelated and unexpected changes. In the end, HashiCorp and AWS performed a large-scale outreach to help upgrade and fixing the misconfigurations. Provider maintainers and practitioners lost considerable time.

### Data Sources

A separate class of Terraform resource types are [data sources](https://www.terraform.io/docs/language/data-sources/). These are typically intended as a configuration method to lookup or fetch data in a read-only manner. Data sources should not have side effects on the remote system.

When discussing data sources, they are typically classified by the intended number of return objects or data. Singular data sources represent a one-to-one lookup or data operation. Plural data sources represent a one-to-many operation.

#### Plural Data Sources

These data sources are intended to return zero, one, or many results, usually associated with a managed resource type. Typically results are a set unless ordering guarantees are provided by the remote system. These should be named with a plural suffix (e.g. `s` or `es`) and should not include any specific attribute in the naming (e.g. prefer `aws_ec2_transit_gateways` instead of `aws_ec2_transit_gateway_ids`).

#### Singular Data Sources

These data sources are intended to return one result or an error. These should not include any specific attribute in the naming (e.g. prefer `aws_ec2_transit_gateway` instead of `aws_ec2_transit_gateway_id`).

### IAM Resource-Based Policy Resources

For some AWS components, the AWS API allows specifying an [IAM resource-based policy](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_identity-vs-resource.html), the IAM policy to associate with a component. Some examples include:

* ECR Repository Policies
* EFS File System Policies
* SNS Topic Policies

Provider developers should implement this capability in a new resource rather than adding it to the associated resource. Reasons for this include:

* Many of the policies must include the Amazon Resource Name (ARN) of the resource. Working around this requirement with custom difference handling within a self-contained resource is unnecessarily cumbersome.
* Some policies involving multiple resources need to cross-reference each other's ARNs. Without a separate resource, this introduces a configuration cycle.
* Splitting the resources allows operators to logically split their configurations into purely operational and security boundaries. This allows environments to have distinct practitioners roles and permissions for IAM versus infrastructure changes.

One rare exception to this guideline is where the policy is _required_ during resource creation.

### Managing Resource Running State

The AWS API provides the ability to start, stop, enable, or disable some AWS components. Some examples include:

* Batch Job Queues
* CloudFront Distributions
* RDS DB Event Subscriptions

In this situation, provider developers should implement this ability within the resource instead of creating a separate resource. Since a practitioner cannot practically manage interaction with a resource's states in Terraform's declarative configuration, developers should implement the state management in the resource. This design provides consistency and future-proofing even where updating a resource in the current API is not problematic.

### Task Execution and Waiter Resources

Some AWS operations are asynchronous. Terraform requests that AWS perform a task. Initially, AWS only notifies Terraform that it received the request. Terraform then requests the status while awaiting completion. Examples of this include:

* ACM Certificate validation
* EC2 AMI copying
* RDS DB Cluster Snapshot management

In this situation, provider developers should create a separate resource representing the task, assuming that the AWS service API provides operations to start the task and read its status. Adding the task functionality to the parent resource muddies its infrastructure-management purpose. The maintainers prefer this approach even though there is some duplication of an existing resource. For example, the provider has a resource for copying an EC2 AMI in addition to the EC2 AMI resource itself. This modularity allows practitioners to manage the result of the task resource with another resource.

For a related consideration, see the [Managing Resource Running State section](#managing-resource-running-state).

### Versioned Resources

AWS supports having multiple versions of some components. Examples of this include:

* ECS Task Definitions
* Lambda Functions
* Secrets Manager Secrets

In general, provider developers should create a separate resource to represent a single version. For example, the provider has both the `aws_secretsmanager_secret` and `aws_secretsmanager_secret_version` resources. However, in some cases, developers should handle versioning in the main resource.

In deciding when to create a separate resource, follow these guidelines:

* If AWS necessarily creates a version when you make a new AWS component, include version handling in the same Terraform resource. Creating an AWS component with one Terraform resource and later using a different resource for updates is confusing.
* If the AWS service API allows deleting versions and practitioners will want to delete versions, provider developers should implement a separate version resource.
* If the API only supports publishing new versions, either method is acceptable, however most current implementations are self-contained. Terraform's current configuration language does not natively support triggering resource updates or recreation across resources without a state value change. This can make the implementation more difficult for practitioners without special resource and configuration workarounds, such as a `triggers` attribute. If this changes in the future, then this guidance may be updated towards separate resources, following the [Task Execution and Waiter Resources](#task-execution-and-waiter-resources) guidance.

## Other Considerations

### AWS Credential Exfiltration

In the interest of security, the maintainers will not approve data sources that provide the ability to reference or export the AWS credentials of the running provider. There are valid use cases for this information, such as to execute AWS CLI calls as part of the same Terraform configuration. However, this mechanism may allow credentials to be discovered and used outside of Terraform. Some specific concerns include:

* The values may be visible in Terraform user interface output or logging, allowing anyone with user interface or log access to see the credentials.
* The values are currently stored in plaintext in the Terraform state, allowing anyone with access to the state file or another Terraform configuration that references the state access to the credentials.
* Any new related functionality, while opt-in to implement, is also opt-in to prevent via security controls or policies. Adopting a weaker default security posture requires advance notice and prevents organizations that implement those controls from updating to a version with any such functionality.
