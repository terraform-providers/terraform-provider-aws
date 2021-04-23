---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_cluster"
description: |-
  Manages an EKS Cluster
---

# Resource: aws_eks_cluster

Manages an EKS Cluster.

## Example Usage

### Basic Usage

```hcl
resource "aws_eks_cluster" "example" {
  name     = "example"
  role_arn = aws_iam_role.example.arn

  vpc_config {
    subnet_ids = [aws_subnet.example1.id, aws_subnet.example2.id]
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Cluster handling.
  # Otherwise, EKS will not be able to properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
    aws_iam_role_policy_attachment.example-AmazonEKSClusterPolicy,
    aws_iam_role_policy_attachment.example-AmazonEKSVPCResourceController,
  ]
}

output "endpoint" {
  value = aws_eks_cluster.example.endpoint
}

output "kubeconfig-certificate-authority-data" {
  value = aws_eks_cluster.example.certificate_authority[0].data
}
```

### Example IAM Role for EKS Cluster

```hcl
resource "aws_iam_role" "example" {
  name = "eks-cluster-example"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "example-AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.example.name
}

# Optionally, enable Security Groups for Pods
# Reference: https://docs.aws.amazon.com/eks/latest/userguide/security-groups-for-pods.html
resource "aws_iam_role_policy_attachment" "example-AmazonEKSVPCResourceController" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
  role       = aws_iam_role.example.name
}
```

### Enabling Control Plane Logging

[EKS Control Plane Logging](https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html) can be enabled via the `enabled_cluster_log_types` argument. To manage the CloudWatch Log Group retention period, the [`aws_cloudwatch_log_group` resource](/docs/providers/aws/r/cloudwatch_log_group.html) can be used.

-> The below configuration uses [`depends_on`](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html) to prevent ordering issues with EKS automatically creating the log group first and a variable for naming consistency. Other ordering and naming methodologies may be more appropriate for your environment.

```hcl
variable "cluster_name" {
  default = "example"
  type    = string
}

resource "aws_eks_cluster" "example" {
  depends_on = [aws_cloudwatch_log_group.example]

  enabled_cluster_log_types = ["api", "audit"]
  name                      = var.cluster_name

  # ... other configuration ...
}

resource "aws_cloudwatch_log_group" "example" {
  name              = "/aws/eks/${var.cluster_name}/cluster"
  retention_in_days = 7

  # ... potentially other configuration ...
}
```

### Enabling IAM Roles for Service Accounts

Only available on Kubernetes version 1.13 and 1.14 clusters created or upgraded on or after September 3, 2019. For more information about this feature, see the [EKS User Guide](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html).

```hcl
resource "aws_eks_cluster" "example" {
  # ... other configuration ...
}

data "tls_certificate" "example" {
  url = aws_eks_cluster.example.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "example" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.example.certificates[0].sha1_fingerprint]
  url             = aws_eks_cluster.example.identity[0].oidc[0].issuer
}

data "aws_iam_policy_document" "example_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    condition {
      test     = "StringEquals"
      variable = "${replace(aws_iam_openid_connect_provider.example.url, "https://", "")}:sub"
      values   = ["system:serviceaccount:kube-system:aws-node"]
    }

    principals {
      identifiers = [aws_iam_openid_connect_provider.example.arn]
      type        = "Federated"
    }
  }
}

resource "aws_iam_role" "example" {
  assume_role_policy = data.aws_iam_policy_document.example_assume_role_policy.json
  name               = "example"
}
```

After adding inline IAM Policies (e.g. [`aws_iam_role_policy` resource](/docs/providers/aws/r/iam_role_policy.html)) or attaching IAM Policies (e.g. [`aws_iam_policy` resource](/docs/providers/aws/r/iam_policy.html) and [`aws_iam_role_policy_attachment` resource](/docs/providers/aws/r/iam_policy.html)) with the desired permissions to the IAM Role, annotate the Kubernetes service account (e.g. [`kubernetes_service_account` resource](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/service_account)) and recreate any pods.

## Argument Reference

The following arguments are supported:

* `name` – (Required) Name of the cluster.
* `role_arn` - (Required) The Amazon Resource Name (ARN) of the IAM role that provides permissions for the Kubernetes control plane to make calls to AWS API operations on your behalf. Ensure the resource configuration includes explicit dependencies on the IAM Role permissions by adding [`depends_on`](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html) if using the [`aws_iam_role_policy` resource](/docs/providers/aws/r/iam_role_policy.html) or [`aws_iam_role_policy_attachment` resource](/docs/providers/aws/r/iam_role_policy_attachment.html), otherwise EKS cannot delete EKS managed EC2 infrastructure such as Security Groups on EKS Cluster deletion.
* `vpc_config` - (Required) Nested argument for the VPC associated with your cluster. Amazon EKS VPC resources have specific requirements to work properly with Kubernetes. For more information, see [Cluster VPC Considerations](https://docs.aws.amazon.com/eks/latest/userguide/network_reqs.html) and [Cluster Security Group Considerations](https://docs.aws.amazon.com/eks/latest/userguide/sec-group-reqs.html) in the Amazon EKS User Guide. Configuration detailed below.
* `enabled_cluster_log_types` - (Optional) A list of the desired control plane logging to enable. For more information, see [Amazon EKS Control Plane Logging](https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html)
* `encryption_config` - (Optional) Configuration block with encryption configuration for the cluster. Only available on Kubernetes 1.13 and above clusters created after March 6, 2020. Detailed below.
* `kubernetes_network_config` - (Optional) Configuration block with kubernetes network configuration for the cluster. Detailed below. If removed, Terraform will only perform drift detection if a configuration value is provided.
* `tags` - (Optional) Key-value map of resource tags.
* `version` – (Optional) Desired Kubernetes master version. If you do not specify a value, the latest available version at resource creation is used and no upgrades will occur except those automatically triggered by EKS. The value must be configured and increased to upgrade the version when desired. Downgrades are not supported by EKS.

### encryption_config

The following arguments are supported in the `encryption_config` configuration block:

* `provider` - (Required) Configuration block with provider for encryption. Detailed below.
* `resources` - (Required) List of strings with resources to be encrypted. Valid values: `secrets`

#### provider

The following arguments are supported in the `provider` configuration block:

* `key_arn` - (Required) Amazon Resource Name (ARN) of the Key Management Service (KMS) customer master key (CMK). The CMK must be symmetric, created in the same region as the cluster, and if the CMK was created in a different account, the user must have access to the CMK. For more information, see [Allowing Users in Other Accounts to Use a CMK in the AWS Key Management Service Developer Guide](https://docs.aws.amazon.com/kms/latest/developerguide/key-policy-modifying-external-accounts.html).

### vpc_config

* `endpoint_private_access` - (Optional) Indicates whether or not the Amazon EKS private API server endpoint is enabled. Default is `false`.
* `endpoint_public_access` - (Optional) Indicates whether or not the Amazon EKS public API server endpoint is enabled. Default is `true`.
* `public_access_cidrs` - (Optional) List of CIDR blocks. Indicates which CIDR blocks can access the Amazon EKS public API server endpoint when enabled. EKS defaults this to a list with `0.0.0.0/0`. Terraform will only perform drift detection of its value when present in a configuration.
* `security_group_ids` – (Optional) List of security group IDs for the cross-account elastic network interfaces that Amazon EKS creates to use to allow communication between your worker nodes and the Kubernetes control plane.
* `subnet_ids` – (Required) List of subnet IDs. Must be in at least two different availability zones. Amazon EKS creates cross-account elastic network interfaces in these subnets to allow communication between your worker nodes and the Kubernetes control plane.

### kubernetes_network_config

The following arguments are supported in the `kubernetes_network_config` configuration block:

* `service_ipv4_cidr` - (Optional) The CIDR block to assign Kubernetes service IP addresses from. If you don't specify a block, Kubernetes assigns addresses from either the 10.100.0.0/16 or 172.20.0.0/16 CIDR blocks. We recommend that you specify a block that does not overlap with resources in other networks that are peered or connected to your VPC. You can only specify a custom CIDR block when you create a cluster, changing this value will force a new cluster to be created. The block must meet the following requirements:

    * Within one of the following private IP address blocks: 10.0.0.0/8, 172.16.0.0.0/12, or 192.168.0.0/16.

    * Doesn't overlap with any CIDR block assigned to the VPC that you selected for VPC.

    * Between /24 and /12.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the cluster.
* `arn` - The Amazon Resource Name (ARN) of the cluster.
* `certificate_authority` - Nested attribute containing `certificate-authority-data` for your cluster.
    * `data` - The base64 encoded certificate data required to communicate with your cluster. Add this to the `certificate-authority-data` section of the `kubeconfig` file for your cluster.
* `endpoint` - The endpoint for your Kubernetes API server.
* `identity` - Nested attribute containing identity provider information for your cluster. Only available on Kubernetes version 1.13 and 1.14 clusters created or upgraded on or after September 3, 2019.
    * `oidc` - Nested attribute containing [OpenID Connect](https://openid.net/connect/) identity provider information for the cluster.
        * `issuer` - Issuer URL for the OpenID Connect identity provider.
* `platform_version` - The platform version for the cluster.
* `status` - The status of the EKS cluster. One of `CREATING`, `ACTIVE`, `DELETING`, `FAILED`.
* `version` - The Kubernetes server version for the cluster.
* `vpc_config` - Additional nested attributes:
    * `cluster_security_group_id` - The cluster security group that was created by Amazon EKS for the cluster.
    * `vpc_id` - The VPC associated with your cluster.

## Timeouts

`aws_eks_cluster` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `30 minutes`) How long to wait for the EKS Cluster to be created.
* `update` - (Default `60 minutes`) How long to wait for the EKS Cluster to be updated.
Note that the `update` timeout is used separately for both `version` and `vpc_config` update timeouts.
* `delete` - (Default `15 minutes`) How long to wait for the EKS Cluster to be deleted.

## Import

EKS Clusters can be imported using the `name`, e.g.

```
$ terraform import aws_eks_cluster.my_cluster my_cluster
```
