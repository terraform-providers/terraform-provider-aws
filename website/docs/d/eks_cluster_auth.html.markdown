---
layout: "aws"
page_title: "AWS: aws_eks_cluster_auth"
sidebar_current: "docs-aws-datasource-eks-cluster-auth"
description: |-
  Get an authentication token to communicate with an EKS Cluster
---

# Data Source: aws_eks_cluster

Get an authentication token to communicate with an EKS cluster.

Uses IAM credentials from the AWS provider to generate a temporary token that is compatible with
[AWS IAM Authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator) authentication.
This can be used to authenticate to an EKS cluster or to a cluster that has the AWS IAM Authenticator
server configured.

## Example Usage

```hcl

data "aws_eks_cluster" "example" {
  name = "example"
}

data "aws_eks_cluster_auth" "example" {
  name = "example"
}

provider "kubernetes" {
  host                   = "${data.aws_eks_cluster.cluster.endpoint}"
  cluster_ca_certificate = "${base64decode(data.aws_eks_cluster.cluster.certificate_authority.0.data)}"
  token                  = "${data.aws_eks_cluster_auth.example.token}"
  load_config_file       = false
}
```

## Argument Reference

* `name` - (Required) The name of the cluster

## Attributes Reference

* `token` - The token to use to authenticate with the cluster.
