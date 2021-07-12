---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_node_group"
description: |-
  Retrieve information about an EKS Node Group
---

# Data Source: aws_eks_node_group

Retrieve information about an EKS Node Group.

## Example Usage

```hcl
data "aws_eks_node_group" "example" {
  cluster_name    = "example"
  node_group_name = "example"
}

output "node_group_arn" {
  value = "${data.aws_eks_node_group.example.node_role_arn}"
}

output "subnet_ids" {
  value = "${data.aws_eks_node_group.example.subnet_ids}"
}
```

## Argument Reference

* `cluster_name` - (Required) The name of the cluster.
* `node_group_name` - (Required) The name of the node group.

## Attributes Reference

* `id` - EKS Cluster name and EKS Node Group name separated by a colon (`:`).
* `ami_type` - Type of Amazon Machine Image (AMI) associated with the EKS Node Group.
* `arn` - Amazon Resource Name (ARN) of the EKS Node Group.
* `disk_size` - Disk size in GiB for worker nodes.
* `instance_types` - Set of instance types associated with the EKS Node Group.
* `labels` - Key-value map of Kubernetes labels. Only labels that are applied with the EKS API are managed by this argument. Other Kubernetes labels applied to the EKS Node Group will not be managed.
* `node_role_arn` – Amazon Resource Name (ARN) of the IAM Role that provides permissions for the EKS Node Group.
* `release_version` – AMI version of the EKS Node Group.
* `remote_access` - Configuration block with remote access settings.
    * `ec2_ssh_key` - EC2 Key Pair name that provides access for SSH communication with the worker nodes in the EKS Node Group.
    * `source_security_group_ids` - Set of EC2 Security Group IDs to allow SSH access (port 22) from on the worker nodes.
* `resources` - List of objects containing information about underlying resources.
    * `autoscaling_groups` - List of objects containing information about AutoScaling Groups.
        * `name` - Name of the AutoScaling Group.
    * `remote_access_security_group_id` - Identifier of the remote access EC2 Security Group.
* `scaling_config` - Configuration block with scaling settings.
    * `desired_size` - Desired number of worker nodes.
    * `max_size` - Maximum number of worker nodes.
    * `min_size` - Minimum number of worker nodes.
* `status` - Status of the EKS Node Group.
* `subnet_ids` – Identifiers of EC2 Subnets to associate with the EKS Node Group.
* `tags` - Key-value map of resource tags.
* `version` – Kubernetes version.
