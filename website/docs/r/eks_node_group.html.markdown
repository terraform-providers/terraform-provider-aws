---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_node_group"
description: |-
  Manages an EKS Node Group
---

# Resource: aws_eks_node_group

Manages an EKS Node Group, which can provision and optionally update an Auto Scaling Group of Kubernetes worker nodes compatible with EKS. Additional documentation about this functionality can be found in the [EKS User Guide](https://docs.aws.amazon.com/eks/latest/userguide/managed-node-groups.html).

## Example Usage

```hcl
resource "aws_eks_node_group" "example" {
  cluster_name    = aws_eks_cluster.example.name
  node_group_name = "example"
  node_role_arn   = aws_iam_role.example.arn
  subnet_ids      = aws_subnet.example[*].id

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Node Group handling.
  # Otherwise, EKS will not be able to properly delete EC2 Instances and Elastic Network Interfaces.
  depends_on = [
    aws_iam_role_policy_attachment.example-AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.example-AmazonEKS_CNI_Policy,
    aws_iam_role_policy_attachment.example-AmazonEC2ContainerRegistryReadOnly,
  ]
}
```

### Ignoring Changes to Desired Size

You can utilize the generic Terraform resource [lifecycle configuration block](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html) with `ignore_changes` to create an EKS Node Group with an initial size of running instances, then ignore any changes to that count caused externally (e.g. Application Autoscaling).

```hcl
resource "aws_eks_node_group" "example" {
  # ... other configurations ...

  scaling_config {
    # Example: Create EKS Node Group with 2 instances to start
    desired_size = 2

    # ... other configurations ...
  }

  # Optional: Allow external changes without Terraform plan difference
  lifecycle {
    ignore_changes = [scaling_config[0].desired_size]
  }
}
```

### Example IAM Role for EKS Node Group

```hcl
resource "aws_iam_role" "example" {
  name = "eks-node-group-example"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "example-AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.example.name
}

resource "aws_iam_role_policy_attachment" "example-AmazonEKS_CNI_Policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.example.name
}

resource "aws_iam_role_policy_attachment" "example-AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.example.name
}
```

### Example Subnets for EKS Node Group

```hcl
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_subnet" "example" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.example.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.example.id

  tags = {
    "kubernetes.io/cluster/${aws_eks_cluster.example.name}" = "shared"
  }
}
```

## Argument Reference

The following arguments are required:

* `cluster_name` – (Required) Name of the EKS Cluster.
* `node_group_name` – (Required) Name of the EKS Node Group.
* `node_role_arn` – (Required) Amazon Resource Name (ARN) of the IAM Role that provides permissions for the EKS Node Group.
* `scaling_config` - (Required) Configuration block with scaling settings. Detailed below.
* `subnet_ids` – (Required) Identifiers of EC2 Subnets to associate with the EKS Node Group. These subnets must have the following resource tag: `kubernetes.io/cluster/CLUSTER_NAME` (where `CLUSTER_NAME` is replaced with the name of the EKS Cluster).

The following arguments are optional:

* `ami_type` - (Optional) Type of Amazon Machine Image (AMI) associated with the EKS Node Group. Defaults to `AL2_x86_64`. Valid values: `AL2_x86_64`, `AL2_x86_64_GPU`, `AL2_ARM_64`. Terraform will only perform drift detection if a configuration value is provided.
* `capacity_type` - (Optional) Type of capacity associated with the EKS Node Group. Valid values: `ON_DEMAND`, `SPOT`. Terraform will only perform drift detection if a configuration value is provided.
* `disk_size` - (Optional) Disk size in GiB for worker nodes. Defaults to `20`. Terraform will only perform drift detection if a configuration value is provided.
* `force_update_version` - (Optional) Force version update if existing pods are unable to be drained due to a pod disruption budget issue.
* `instance_types` - (Optional) List of instance types associated with the EKS Node Group. Defaults to `["t3.medium"]`. Terraform will only perform drift detection if a configuration value is provided.
* `labels` - (Optional) Key-value map of Kubernetes labels. Only labels that are applied with the EKS API are managed by this argument. Other Kubernetes labels applied to the EKS Node Group will not be managed.
* `launch_template` - (Optional) Configuration block with Launch Template settings. Detailed below.
* `release_version` – (Optional) AMI version of the EKS Node Group. Defaults to latest version for Kubernetes version.
* `remote_access` - (Optional) Configuration block with remote access settings. Detailed below.
* `tags` - (Optional) Key-value map of resource tags.
* `version` – (Optional) Kubernetes version. Defaults to EKS Cluster Kubernetes version. Terraform will only perform drift detection if a configuration value is provided.

### launch_template Configuration Block

~> **NOTE:** Either `id` or `name` must be specified.

* `id` - (Optional) Identifier of the EC2 Launch Template. Conflicts with `name`.
* `name` - (Optional) Name of the EC2 Launch Template. Conflicts with `id`.
* `version` - (Required) EC2 Launch Template version number. While the API accepts values like `$Default` and `$Latest`, the API will convert the value to the associated version number (e.g. `1`) on read and Terraform will show a difference on next plan. Using the `default_version` or `latest_version` attribute of the `aws_launch_template` resource or data source is recommended for this argument.

### remote_access Configuration Block

* `ec2_ssh_key` - (Optional) EC2 Key Pair name that provides access for SSH communication with the worker nodes in the EKS Node Group. If you specify this configuration, but do not specify `source_security_group_ids` when you create an EKS Node Group, port 22 on the worker nodes is opened to the Internet (0.0.0.0/0).
* `source_security_group_ids` - (Optional) Set of EC2 Security Group IDs to allow SSH access (port 22) from on the worker nodes. If you specify `ec2_ssh_key`, but do not specify this configuration when you create an EKS Node Group, port 22 on the worker nodes is opened to the Internet (0.0.0.0/0).

### scaling_config Configuration Block

* `desired_size` - (Required) Desired number of worker nodes.
* `max_size` - (Required) Maximum number of worker nodes.
* `min_size` - (Required) Minimum number of worker nodes.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the EKS Node Group.
* `id` - EKS Cluster name and EKS Node Group name separated by a colon (`:`).
* `resources` - List of objects containing information about underlying resources.
    * `autoscaling_groups` - List of objects containing information about AutoScaling Groups.
        * `name` - Name of the AutoScaling Group.
    * `remote_access_security_group_id` - Identifier of the remote access EC2 Security Group.
* `status` - Status of the EKS Node Group.

## Timeouts

`aws_eks_node_group` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `60 minutes`) How long to wait for the EKS Node Group to be created.
* `update` - (Default `60 minutes`) How long to wait for the EKS Node Group to be updated. Note that the `update` timeout is used separately for both configuration and version update operations.
* `delete` - (Default `60 minutes`) How long to wait for the EKS Node Group to be deleted.

## Import

EKS Node Groups can be imported using the `cluster_name` and `node_group_name` separated by a colon (`:`), e.g.

```
$ terraform import aws_eks_node_group.my_node_group my_cluster:my_node_group
```
