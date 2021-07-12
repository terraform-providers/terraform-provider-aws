---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_oidc_thumbprint"
description: |-
  Retrieve CA fingerprint for an EKS OIDC endpoint in a given region 
---

# Data Source: aws_eks_oidc_thumbprint

Retrieve TLS certificate fingerprint for an EKS OIDC endpoint in a given region

## Example Usage

```hcl
data "aws_eks_cluster" "example" {
  name = "example"
}

data "aws_eks_oidc_thumbprint" "example" {
  region = "eu-west-1"
}

resource "aws_iam_openid_connect_provider" "oidc" {
  # Only available on Kubernetes version 1.13 and 1.14 clusters created or upgraded on or after September 3, 2019.
  url = data.aws_eks_cluster.example.identity[0].oidc[0].issuer
  client_id_list = [ "sts.amazonaws.com" ]
  thumbprint_list = [ data.aws_eks_oidc_thumbprint.example.sha1_hash ]
}
```

## Argument Reference

* `region` - (Required) The name of the region in which provisioning an EKS cluster

## Attributes Reference

* `sha1_hash` - The SHA1 fingerprint to be used in an IAM OpenID Connect provider configuration 
