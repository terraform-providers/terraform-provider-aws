---
layout: "aws"
page_title: "AWS: aws_eks_cluster"
sidebar_current: "docs-aws-datasource-eks-cluster"
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
  value = "${data.aws_eks_cluster.example.endpoint}"
}

output "kubeconfig-certificate-authority-data" {
  value = "${data.aws_eks_cluster.example.certificate_authority.0.data}"
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
* `endpoint` - The endpoint for your Kubernetes API server.
* `platform_version` - The platform version for the cluster.
* `role_arn` - The Amazon Resource Name (ARN) of the IAM role that provides permissions for the Kubernetes control plane to make calls to AWS API operations on your behalf.
* `version` - The Kubernetes server version for the cluster.
* `vpc_config` - Nested attribute containing VPC configuration for the cluster.
  * `security_group_ids` – List of security group IDs
  * `subnet_ids` – List of subnet IDs
  * `vpc_id` – The VPC associated with your cluster.
