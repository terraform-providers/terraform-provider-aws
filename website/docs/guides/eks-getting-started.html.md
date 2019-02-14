---
layout: "aws"
page_title: "EKS Getting Started Guide"
sidebar_current: "docs-aws-guide-eks-getting-started"
description: |-
  Using Terraform to configure AWS EKS.
---

# Getting Started with AWS EKS

The Amazon Web Services EKS service allows for simplified management of
[Kubernetes](https://kubernetes.io/) servers. While the service itself is
quite simple from an operator perspective, understanding how it interconnects
with other pieces of the AWS service universe and how to configure local
Kubernetes clients to manage clusters can be helpful.

While the [EKS User Guide](https://docs.aws.amazon.com/eks/latest/userguide/)
provides much of the up-to-date information about getting started with the service
from a generic standpoint, this guide provides a Terraform configuration based
introduction.

This guide will show how to deploy a sample architecture using Terraform. The
guide assumes some basic familiarity with Kubernetes but does not
assume any pre-existing deployment. It also assumes that you are familiar
with the usual Terraform plan/apply workflow; if you're new to Terraform
itself, refer first to [the Getting Started guide](/intro/getting-started/install.html).

It is worth noting that there are other valid ways to use these services and
resources that make different tradeoffs. We encourage readers to consult the official documentation for the respective services and resources for additional context and
best-practices. This guide can still serve as an introduction to the main resources
associated with these services, even if you choose a different architecture.

<!-- TOC depthFrom:2 -->

- [Guide Overview](#guide-overview)
- [Preparation](#preparation)
- [Create Sample Architecture in AWS](#create-sample-architecture-in-aws)
    - [Cluster Name Variable](#cluster-name-variable)
    - [Base VPC Networking](#base-vpc-networking)
    - [Kubernetes Masters](#kubernetes-masters)
        - [EKS Master Cluster IAM Role](#eks-master-cluster-iam-role)
        - [EKS Master Cluster Security Group](#eks-master-cluster-security-group)
        - [EKS Master Cluster](#eks-master-cluster)
    - [Configuring kubectl for EKS](#configuring-kubectl-for-eks)
    - [Kubernetes Worker Nodes](#kubernetes-worker-nodes)
        - [Worker Node IAM Role and Instance Profile](#worker-node-iam-role-and-instance-profile)
        - [Worker Node Security Group](#worker-node-security-group)
        - [Worker Node Access to EKS Master Cluster](#worker-node-access-to-eks-master-cluster)
        - [Worker Node AutoScaling Group](#worker-node-autoscaling-group)
        - [Required Kubernetes Configuration to Join Worker Nodes](#required-kubernetes-configuration-to-join-worker-nodes)
- [Destroy Sample Architecture in AWS](#destroy-sample-architecture-in-aws)

<!-- /TOC -->

## Guide Overview

~> **Warning:** Following this guide will create objects in your AWS account
that will cost you money against your AWS bill.

The sample architecture introduced here includes the following resources:

* EKS Cluster: AWS managed Kubernetes cluster of master servers
* AutoScaling Group containing 2 m4.large instances based on the latest EKS Amazon Linux 2 AMI: Operator managed Kubernetes worker nodes for running Kubernetes service deployments
* Associated VPC, Internet Gateway, Security Groups, and Subnets: Operator managed networking resources for the EKS Cluster and worker node instances
* Associated IAM Roles and Policies: Operator managed access resources for EKS and worker node instances

## Preparation

In order to follow this guide you will need an AWS account and to have
Terraform installed.
[Configure your credentials](/docs/providers/aws/index.html#authentication)
so that Terraform is able to act on your behalf.

For simplicity here, we will assume you are already using a set of IAM
credentials with suitable access to create AutoScaling, EC2, EKS, and IAM
resources. If you are not sure and are working in an AWS account used only for
development, the simplest approach to get started is to use credentials with
full administrative access to the target AWS account.

If you are planning to locally use the standard Kubernetes client, `kubectl`,
it must be at least version 1.10 to support `exec` authentication with usage
of `aws-iam-authenticator`. For additional information about installation
and configuration of these applications, see their official documentation.

Relevant Links:

* [Kubernetes Client Install Guide](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [AWS IAM Authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator)

## Create Sample Architecture in AWS

~> **NOTE:** We recommend using this guide to build a separate Terraform
configuration (for easy tear down) and more importantly running it in a
separate AWS account as your production infrastructure. While it is
self-contained and should not affect existing infrastructure, its always best
to be cautious!

~> **NOTE:** If you would rather see the full sample Terraform configuration
for this guide rather than the individual pieces, it can be found at:
https://github.com/terraform-providers/terraform-provider-aws/tree/master/examples/eks-getting-started

### Cluster Name Variable

The below sample Terraform configurations reference a variable called
`cluster-name` (`var.cluster-name`) which is used for consistency. Feel free
to substitute your own cluster name or create the variable configuration:

```hcl
variable "cluster-name" {
  default = "terraform-eks-demo"
  type    = "string"
}
```

### Base VPC Networking

EKS requires the usage of [Virtual Private Cloud](https://aws.amazon.com/vpc/) to
provide the base for its networking configuration.

~> **NOTE:** The usage of the specific `kubernetes.io/cluster/*` resource tags below are required for EKS and Kubernetes to discover and manage networking resources.

The below will create a 10.0.0.0/16 VPC, two 10.0.X.0/24 subnets, an internet
gateway, and setup the subnet routing to route external traffic through the
internet gateway:

```hcl
# This data source is included for ease of sample architecture deployment
# and can be swapped out as necessary.
data "aws_availability_zones" "available" {}

resource "aws_vpc" "demo" {
  cidr_block = "10.0.0.0/16"

  tags = "${
    map(
     "Name", "terraform-eks-demo-node",
     "kubernetes.io/cluster/${var.cluster-name}", "shared",
    )
  }"
}

resource "aws_subnet" "demo" {
  count = 2

  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = "${aws_vpc.demo.id}"

  tags = "${
    map(
     "Name", "terraform-eks-demo-node",
     "kubernetes.io/cluster/${var.cluster-name}", "shared",
    )
  }"
}

resource "aws_internet_gateway" "demo" {
  vpc_id = "${aws_vpc.demo.id}"

  tags = {
    Name = "terraform-eks-demo"
  }
}

resource "aws_route_table" "demo" {
  vpc_id = "${aws_vpc.demo.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.demo.id}"
  }
}

resource "aws_route_table_association" "demo" {
  count = 2

  subnet_id      = "${aws_subnet.demo.*.id[count.index]}"
  route_table_id = "${aws_route_table.demo.id}"
}
```

### Kubernetes Masters

This is where the EKS service comes into play. It requires a few operator
managed resources beforehand so that Kubernetes can properly manage other
AWS services as well as allow inbound networking communication from your
local workstation (if desired) and worker nodes.

#### EKS Master Cluster IAM Role

The below is an example IAM role and policy to allow the EKS service to
manage or retrieve data from other AWS services. It is also possible to create
these policies with the [`aws_iam_policy_document` data source](/docs/providers/aws/d/iam_policy_document.html)

For the latest required policy, see the [EKS User Guide](https://docs.aws.amazon.com/eks/latest/userguide/).

```hcl
resource "aws_iam_role" "demo-cluster" {
  name = "terraform-eks-demo-cluster"

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

resource "aws_iam_role_policy_attachment" "demo-cluster-AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = "${aws_iam_role.demo-cluster.name}"
}

resource "aws_iam_role_policy_attachment" "demo-cluster-AmazonEKSServicePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSServicePolicy"
  role       = "${aws_iam_role.demo-cluster.name}"
}
```

#### EKS Master Cluster Security Group

This security group controls networking access to the Kubernetes masters.
We will later configure this with an ingress rule to allow traffic from the
worker nodes.

```hcl
resource "aws_security_group" "demo-cluster" {
  name        = "terraform-eks-demo-cluster"
  description = "Cluster communication with worker nodes"
  vpc_id      = "${aws_vpc.demo.id}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "terraform-eks-demo"
  }
}

# OPTIONAL: Allow inbound traffic from your local workstation external IP
#           to the Kubernetes. You will need to replace A.B.C.D below with
#           your real IP. Services like icanhazip.com can help you find this.
resource "aws_security_group_rule" "demo-cluster-ingress-workstation-https" {
  cidr_blocks       = ["A.B.C.D/32"]
  description       = "Allow workstation to communicate with the cluster API Server"
  from_port         = 443
  protocol          = "tcp"
  security_group_id = "${aws_security_group.demo-cluster.id}"
  to_port           = 443
  type              = "ingress"
}
```

#### EKS Master Cluster

This resource is the actual Kubernetes master cluster. It can take a few minutes to
provision in AWS.

```hcl
resource "aws_eks_cluster" "demo" {
  name            = "${var.cluster-name}"
  role_arn        = "${aws_iam_role.demo-cluster.arn}"

  vpc_config {
    security_group_ids = ["${aws_security_group.demo-cluster.id}"]
    subnet_ids         = ["${aws_subnet.demo.*.id}"]
  }

  depends_on = [
    "aws_iam_role_policy_attachment.demo-cluster-AmazonEKSClusterPolicy",
    "aws_iam_role_policy_attachment.demo-cluster-AmazonEKSServicePolicy",
  ]
}
```

### Configuring kubectl for EKS

-> This section only provides some example methods for configuring `kubectl` to communicate with EKS servers. Managing Kubernetes clients and configurations is outside the scope of this guide.

If you are planning on using `kubectl` to manage the Kubernetes cluster, now
might be a great time to configure your client. After configuration, you can
verify cluster access via `kubectl version` displaying server version
information in addition to local client version information.

The AWS CLI [`eks update-kubeconfig`](https://docs.aws.amazon.com/cli/latest/reference/eks/update-kubeconfig.html)
command provides a simple method to create or update configuration files.

If you would rather update your configuration manually, the below Terraform output
generates a sample `kubectl` configuration to connect to your cluster. This can
be placed into a Kubernetes configuration file, e.g. `~/.kube/config`

```hcl
locals {
  kubeconfig = <<KUBECONFIG


apiVersion: v1
clusters:
- cluster:
    server: ${aws_eks_cluster.demo.endpoint}
    certificate-authority-data: ${aws_eks_cluster.demo.certificate_authority.0.data}
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: aws
  name: aws
current-context: aws
kind: Config
preferences: {}
users:
- name: aws
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1alpha1
      command: aws-iam-authenticator
      args:
        - "token"
        - "-i"
        - "${var.cluster-name}"
KUBECONFIG
}

output "kubeconfig" {
  value = "${local.kubeconfig}"
}
```

### Kubernetes Worker Nodes

The EKS service does not currently provide managed resources for running
worker nodes. Here we will create a few operator managed resources so that
Kubernetes can properly manage other AWS services, networking access, and
finally a configuration that allows automatic scaling of worker nodes.

#### Worker Node IAM Role and Instance Profile

The below is an example IAM role and policy to allow the worker nodes to
manage or retrieve data from other AWS services. It is used by Kubernetes
to allow worker nodes to join the cluster. It is also possible to create
these policies with the [`aws_iam_policy_document` data source](/docs/providers/aws/d/iam_policy_document.html)

For the latest required policy, see the [EKS User Guide](https://docs.aws.amazon.com/eks/latest/userguide/).

```hcl
resource "aws_iam_role" "demo-node" {
  name = "terraform-eks-demo-node"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "demo-node-AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = "${aws_iam_role.demo-node.name}"
}

resource "aws_iam_role_policy_attachment" "demo-node-AmazonEKS_CNI_Policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = "${aws_iam_role.demo-node.name}"
}

resource "aws_iam_role_policy_attachment" "demo-node-AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = "${aws_iam_role.demo-node.name}"
}

resource "aws_iam_instance_profile" "demo-node" {
  name = "terraform-eks-demo"
  role = "${aws_iam_role.demo-node.name}"
}
```

#### Worker Node Security Group

This security group controls networking access to the Kubernetes worker nodes.

```hcl
resource "aws_security_group" "demo-node" {
  name        = "terraform-eks-demo-node"
  description = "Security group for all nodes in the cluster"
  vpc_id      = "${aws_vpc.demo.id}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = "${
    map(
     "Name", "terraform-eks-demo-node",
     "kubernetes.io/cluster/${var.cluster-name}", "owned",
    )
  }"
}

resource "aws_security_group_rule" "demo-node-ingress-self" {
  description              = "Allow node to communicate with each other"
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = "${aws_security_group.demo-node.id}"
  source_security_group_id = "${aws_security_group.demo-node.id}"
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "demo-node-ingress-cluster-https" {
  description              = "Allow worker Kubelets and pods to receive communication from the cluster control plane"
  from_port                = 443
  protocol                 = "tcp"
  security_group_id        = "${aws_security_group.demo-node.id}"
  source_security_group_id = "${aws_security_group.demo-cluster.id}"
  to_port                  = 443
  type                     = "ingress"
}

resource "aws_security_group_rule" "demo-node-ingress-cluster-others" {
  description              = "Allow worker Kubelets and pods to receive communication from the cluster control plane"
  from_port                = 1025
  protocol                 = "tcp"
  security_group_id        = "${aws_security_group.demo-node.id}"
  source_security_group_id = "${aws_security_group.demo-cluster.id}"
  to_port                  = 65535
  type                     = "ingress"
}
```

#### Worker Node Access to EKS Master Cluster

Now that we have a way to know where traffic from the worker nodes is coming
from, we can allow the worker nodes networking access to the EKS master cluster.

```hcl
resource "aws_security_group_rule" "demo-cluster-ingress-node-https" {
  description              = "Allow pods to communicate with the cluster API Server"
  from_port                = 443
  protocol                 = "tcp"
  security_group_id        = "${aws_security_group.demo-cluster.id}"
  source_security_group_id = "${aws_security_group.demo-node.id}"
  to_port                  = 443
  type                     = "ingress"
}
```

#### Worker Node AutoScaling Group

Now we have everything in place to create and manage EC2 instances that will
serve as our worker nodes in the Kubernetes cluster. This setup utilizes
an EC2 AutoScaling Group (ASG) rather than manually working with EC2 instances.
This offers flexibility to scale up and down the worker nodes on demand when
used in conjunction with AutoScaling policies (not implemented here).

First, let us create a data source to fetch the latest Amazon Machine Image
(AMI) that Amazon provides with an EKS compatible Kubernetes baked in. It will filter for and select an AMI compatible with the specific Kubernetes version being deployed.

```hcl
data "aws_ami" "eks-worker" {
  filter {
    name   = "name"
    values = ["amazon-eks-node-${aws_eks_cluster.demo.version}-v*"]
  }

  most_recent = true
  owners      = ["602401143452"] # Amazon EKS AMI Account ID
}
```

Next, lets create an AutoScaling Launch Configuration that uses all our
prerequisite resources to define how to create EC2 instances using them.

```hcl
# This data source is included for ease of sample architecture deployment
# and can be swapped out as necessary.
data "aws_region" "current" {}

# EKS currently documents this required userdata for EKS worker nodes to
# properly configure Kubernetes applications on the EC2 instance.
# We utilize a Terraform local here to simplify Base64 encoding this
# information into the AutoScaling Launch Configuration.
# More information: https://docs.aws.amazon.com/eks/latest/userguide/launch-workers.html
locals {
  demo-node-userdata = <<USERDATA
#!/bin/bash
set -o xtrace
/etc/eks/bootstrap.sh --apiserver-endpoint '${aws_eks_cluster.demo.endpoint}' --b64-cluster-ca '${aws_eks_cluster.demo.certificate_authority.0.data}' '${var.cluster-name}'
USERDATA
}

resource "aws_launch_configuration" "demo" {
  associate_public_ip_address = true
  iam_instance_profile        = "${aws_iam_instance_profile.demo-node.name}"
  image_id                    = "${data.aws_ami.eks-worker.id}"
  instance_type               = "m4.large"
  name_prefix                 = "terraform-eks-demo"
  security_groups             = ["${aws_security_group.demo-node.id}"]
  user_data_base64            = "${base64encode(local.demo-node-userdata)}"

  lifecycle {
    create_before_destroy = true
  }
}
```

Finally, we create an AutoScaling Group that actually launches EC2 instances
based on the AutoScaling Launch Configuration.

~> **NOTE:** The usage of the specific `kubernetes.io/cluster/*` resource tag below is required for EKS and Kubernetes to discover and manage compute resources.

```hcl
resource "aws_autoscaling_group" "demo" {
  desired_capacity     = 2
  launch_configuration = "${aws_launch_configuration.demo.id}"
  max_size             = 2
  min_size             = 1
  name                 = "terraform-eks-demo"
  vpc_zone_identifier  = ["${aws_subnet.demo.*.id}"]

  tag {
    key                 = "Name"
    value               = "terraform-eks-demo"
    propagate_at_launch = true
  }

  tag {
    key                 = "kubernetes.io/cluster/${var.cluster-name}"
    value               = "owned"
    propagate_at_launch = true
  }
}
```

~> **NOTE:** At this point, your Kubernetes cluster will have running masters
and worker nodes, _however_, the worker nodes will not be able to join the
Kubernetes cluster quite yet. The next section has the required Kubernetes
configuration to enable the worker nodes to join the cluster.

#### Required Kubernetes Configuration to Join Worker Nodes

-> While managing Kubernetes cluster and client configurations are beyond the scope of this guide, we provide an example of how to apply the required Kubernetes [`ConfigMap`](http://kubernetes.io/docs/user-guide/configmap/) via `kubectl` below for completeness. See also the [Configuring kubectl for EKS](#configuring-kubectl-for-eks) section.

The EKS service does not provide a cluster-level API parameter or resource to
automatically configure the underlying Kubernetes cluster to allow worker nodes
to join the cluster via AWS IAM role authentication.

To output an example IAM Role authentication `ConfigMap` from your
Terraform configuration:

```hcl
locals {
  config_map_aws_auth = <<CONFIGMAPAWSAUTH


apiVersion: v1
kind: ConfigMap
metadata:
  name: aws-auth
  namespace: kube-system
data:
  mapRoles: |
    - rolearn: ${aws_iam_role.demo-node.arn}
      username: system:node:{{EC2PrivateDNSName}}
      groups:
        - system:bootstrappers
        - system:nodes
CONFIGMAPAWSAUTH
}

output "config_map_aws_auth" {
  value = "${local.config_map_aws_auth}"
}
```

* Run `terraform output config_map_aws_auth` and save the configuration into a file, e.g. `config_map_aws_auth.yaml`
* Run `kubectl apply -f config_map_aws_auth.yaml`
* You can verify the worker nodes are joining the cluster via: `kubectl get nodes --watch`

At this point, you should be able to utilize Kubernetes as expected!

## Destroy Sample Architecture in AWS

Assuming you have built the sample architecture in a separate configuration,
simply use `terraform destroy` to tear down all associated resources with this
guide.
