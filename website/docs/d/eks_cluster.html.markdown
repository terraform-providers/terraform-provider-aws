---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_cluster"
description: |-
  Retrieve information about an EKS Cluster
---

# Data Source: aws_eks_cluster

Retrieve information about an EKS Cluster.

## Example Usage

```hcl
data "aws_eks_cluster" "example" {
  name = "example"
}

output "endpoint" {
  value = data.aws_eks_cluster.example.endpoint
}

output "kubeconfig-certificate-authority-data" {
  value = data.aws_eks_cluster.example.certificate_authority[0].data
}

# Only available on Kubernetes version 1.13 and 1.14 clusters created or upgraded on or after September 3, 2019.
output "identity-oidc-issuer" {
  value = data.aws_eks_cluster.example.identity[0].oidc[0].issuer
}
```

## Argument Reference

* `name` - (Required) The name of the cluster

## Attributes Reference

* `id` - The name of the cluster
* `arn` - The Amazon Resource Name (ARN) of the cluster.
* `certificate_authority` - Nested attribute containing `certificate-authority-data` for your cluster.
    * `data` - The base64 encoded certificate data required to communicate with your cluster. Add this to the `certificate-authority-data` section of the `kubeconfig` file for your cluster.
* `created_at` - The Unix epoch time stamp in seconds for when the cluster was created.
* `enabled_cluster_log_types` - The enabled control plane logs.
* `endpoint` - The endpoint for your Kubernetes API server.
* `identity` - Nested attribute containing identity provider information for your cluster. Only available on Kubernetes version 1.13 and 1.14 clusters created or upgraded on or after September 3, 2019. For an example using this information to enable IAM Roles for Service Accounts, see the [`aws_eks_cluster` resource documentation](/docs/providers/aws/r/eks_cluster.html).
    * `oidc` - Nested attribute containing [OpenID Connect](https://openid.net/connect/) identity provider information for the cluster.
        * `issuer` - Issuer URL for the OpenID Connect identity provider.
* `platform_version` - The platform version for the cluster.
* `role_arn` - The Amazon Resource Name (ARN) of the IAM role that provides permissions for the Kubernetes control plane to make calls to AWS API operations on your behalf.
* `status` - The status of the EKS cluster. One of `CREATING`, `ACTIVE`, `DELETING`, `FAILED`.
* `tags` - Key-value map of resource tags.
* `version` - The Kubernetes server version for the cluster.
* `vpc_config` - Nested list containing VPC configuration for the cluster.
    * `cluster_security_group_id` - The cluster security group that was created by Amazon EKS for the cluster.
    * `endpoint_private_access` - Indicates whether or not the Amazon EKS private API server endpoint is enabled.
    * `endpoint_public_access` - Indicates whether or not the Amazon EKS public API server endpoint is enabled.
    * `public_access_cidrs` - List of CIDR blocks. Indicates which CIDR blocks can access the Amazon EKS public API server endpoint.
    * `security_group_ids` – List of security group IDs
    * `subnet_ids` – List of subnet IDs
    * `vpc_id` – The VPC associated with your cluster.
