---
layout: "aws"
page_title: "AWS: aws_eks_cluster"
sidebar_current: "docs-aws-resource-eks-cluster"
description: |-
  Manages an EKS Cluster
---

# aws_eks_cluster

Manages an EKS Cluster.

## Example Usage

```hcl
resource "aws_eks_cluster" "example" {
  name     = "example"
  role_arn = "${aws_iam_role.example.arn}"

  vpc_config {
    subnet_ids = ["${aws_subnet.example1.id}", "${aws_subnet.example2.id}"]
  }
}

output "endpoint" {
  value = "${aws_eks_cluster.example.endpoint}"
}

output "kubeconfig-certificate-authority-data" {
  value = "${aws_eks_cluster.example.certificate_authority.0.data}"
}
```

## Argument Reference

The following arguments are supported:

* `name` – (Required) Name of the cluster.
* `role_arn` - (Required) The Amazon Resource Name (ARN) of the IAM role that provides permissions for the Kubernetes control plane to make calls to AWS API operations on your behalf.
* `vpc_config` - (Required) Nested argument for the VPC associated with your cluster. Amazon EKS VPC resources have specific requirements to work properly with Kubernetes. For more information, see [Cluster VPC Considerations](https://docs.aws.amazon.com/eks/latest/userguide/network_reqs.html) and [Cluster Security Group Considerations](https://docs.aws.amazon.com/eks/latest/userguide/sec-group-reqs.html) in the Amazon EKS User Guide. Configuration detailed below.
* `version` – (Optional) Desired Kubernetes master version. If you do not specify a value, the latest available version is used.

### vpc_config

* `security_group_ids` – (Optional) List of security group IDs for the cross-account elastic network interfaces that Amazon EKS creates to use to allow communication between your worker nodes and the Kubernetes control plane.
* `subnet_ids` – (Required) List of subnet IDs. Must be in at least two different availability zones. Amazon EKS creates cross-account elastic network interfaces in these subnets to allow communication between your worker nodes and the Kubernetes control plane.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the cluster.
* `arn` - The Amazon Resource Name (ARN) of the cluster.
* `certificate_authority` - Nested attribute containing `certificate-authority-data` for your cluster.
  * `data` - The base64 encoded certificate data required to communicate with your cluster. Add this to the `certificate-authority-data` section of the `kubeconfig` file for your cluster.
* `endpoint` - The endpoint for your Kubernetes API server.
* `platform_version` - The platform version for the cluster.
* `version` - The Kubernetes server version for the cluster.
* `vpc_config` - Additional nested attributes:
  * `vpc_id` - The VPC associated with your cluster.

## Timeouts

`aws_eks_cluster` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

* `create` - (Default `15 minutes`) How long to wait for the EKS Cluster to be created.
* `update` - (Default `60 minutes`) How long to wait for the EKS Cluster to be updated.
* `delete` - (Default `15 minutes`) How long to wait for the EKS Cluster to be deleted.

## Import

EKS Clusters can be imported using the `name`, e.g.

```
$ terraform import aws_eks_cluster.my_cluster my_cluster
```
