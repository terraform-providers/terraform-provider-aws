---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_clusters"
description: |-
  Retrieve information about several filtered EKS Clusters 
---

# Data Source: aws_eks_clusters

Retrieve information about several filtered EKS Clusters .

## Example Usage

```hcl
data "aws_eks_clusters" "tagged_clusters" {
  filter {
    name   = "tag:tagKey"
    values = ["tagValue"]
  }
}

data "aws_iam_policy_document" "assume_role" {
  dynamic "statement" {
    for_each = data.eks_clusters.tagged_clusters
    content {
      effect  = "Allow"
      actions = ["sts:AssumeRoleWithWebIdentity"]

      principals {
        type        = "Federated"
        identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/${trimprefix(statement.value.identity.0.oidc.0.issuer, "https://")}"]
      }
    }
  }
}
```

## Argument Reference

* `filter` - (Required) Configuration block(s) for filtering. Consisting of:
  * `name` - The name of the filter field. Currently only `tag:tagKey` names are supported.
  * `values` - Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

* `clusters` - The filtered cluster list, a list of the following:
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
