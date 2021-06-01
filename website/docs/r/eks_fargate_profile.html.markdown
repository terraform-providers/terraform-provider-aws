---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_fargate_profile"
description: |-
  Manages an EKS Fargate Profile
---

# Resource: aws_eks_fargate_profile

Manages an EKS Fargate Profile.

## Example Usage

```hcl
resource "aws_eks_fargate_profile" "example" {
  cluster_name           = aws_eks_cluster.example.name
  fargate_profile_name   = "example"
  pod_execution_role_arn = aws_iam_role.example.arn
  subnet_ids             = aws_subnet.example[*].id

  selector {
    namespace = "example"
  }
}
```

### Example IAM Role for EKS Fargate Profile

```hcl
resource "aws_iam_role" "example" {
  name = "eks-fargate-profile-example"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "eks-fargate-pods.amazonaws.com"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "example-AmazonEKSFargatePodExecutionRolePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSFargatePodExecutionRolePolicy"
  role       = aws_iam_role.example.name
}
```

## Argument Reference

The following arguments are required:

* `cluster_name` – (Required) Name of the EKS Cluster.
* `fargate_profile_name` – (Required) Name of the EKS Fargate Profile.
* `pod_execution_role_arn` – (Required) Amazon Resource Name (ARN) of the IAM Role that provides permissions for the EKS Fargate Profile.
* `selector` - (Required) Configuration block(s) for selecting Kubernetes Pods to execute with this EKS Fargate Profile. Detailed below.
* `subnet_ids` – (Required) Identifiers of private EC2 Subnets to associate with the EKS Fargate Profile. These subnets must have the following resource tag: `kubernetes.io/cluster/CLUSTER_NAME` (where `CLUSTER_NAME` is replaced with the name of the EKS Cluster).

The following arguments are optional:

* `tags` - (Optional) Key-value map of resource tags.

### selector Configuration Block

The following arguments are required:

* `namespace` - (Required) Kubernetes namespace for selection.

The following arguments are optional:

* `labels` - (Optional) Key-value map of Kubernetes labels for selection.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the EKS Fargate Profile.
* `id` - EKS Cluster name and EKS Fargate Profile name separated by a colon (`:`).
* `status` - Status of the EKS Fargate Profile.

## Timeouts

`aws_eks_fargate_profile` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `10 minutes`) How long to wait for the EKS Fargate Profile to be created.
* `delete` - (Default `10 minutes`) How long to wait for the EKS Fargate Profile to be deleted.

## Import

EKS Fargate Profiles can be imported using the `cluster_name` and `fargate_profile_name` separated by a colon (`:`), e.g.

```
$ terraform import aws_eks_fargate_profile.my_fargate_profile my_cluster:my_fargate_profile
```
