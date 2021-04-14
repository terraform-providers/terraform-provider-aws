---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_addon"
description: |-
  Manages an EKS add-on
---

# Resource: aws_eks_addon

Manages an EKS add-on.

~> **Note:** Amazon EKS add-on can only be used with Amazon EKS Clusters
running version 1.18 with platform version eks.3 or later
because add-ons rely on the Server-side Apply Kubernetes feature,
which is only available in Kubernetes 1.18 and later.

## Example Usage

```terraform
resource "aws_eks_addon" "example" {
  cluster_name = aws_eks_cluster.example.name
  addon_name   = "vpc-cni"
}
```

### Example IAM Role for EKS Addon "vpc-cni" with AWS managed policy

```terraform
resource "aws_eks_cluster" "example" {
  # ... other configuration ...
}

data "tls_certificate" "example" {
  url = aws_eks_cluster.example.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "example" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.example.certificates[0].sha1_fingerprint]
  url             = aws_eks_cluster.example.identity[0].oidc[0].issuer
}

data "aws_iam_policy_document" "example_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    condition {
      test     = "StringEquals"
      variable = "${replace(aws_iam_openid_connect_provider.example.url, "https://", "")}:sub"
      values   = ["system:serviceaccount:kube-system:aws-node"]
    }

    principals {
      identifiers = [aws_iam_openid_connect_provider.example.arn]
      type        = "Federated"
    }
  }
}

resource "aws_iam_role" "example" {
  assume_role_policy = data.aws_iam_policy_document.example_assume_role_policy.json
  name               = "example-vpc-cni-role"
}

resource "aws_iam_role_policy_attachment" "example" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.example.name
}
```

## Argument Reference

The following arguments are required:

* `addon_name` – (Required) Name of the EKS add-on. The name must match one of
  the names returned by [list-addon](https://docs.aws.amazon.com/cli/latest/reference/eks/list-addons.html).
* `cluster_name` – (Required) Name of the EKS Cluster.

The following arguments are optional:

* `addon_version` – (Optional) The version of the EKS add-on. The version must
  match one of the versions returned by [describe-addon-versions](https://docs.aws.amazon.com/cli/latest/reference/eks/describe-addon-versions.html).
* `resolve_conflicts` - (Optional) Define how to resolve parameter value conflicts
  when migrating an existing add-on to an Amazon EKS add-on or when applying
  version updates to the add-on. Valid values are `NONE` and `OVERWRITE`.
* `tags` - (Optional) Key-value map of resource tags.
* `service_account_role_arn` - (Optional) The Amazon Resource Name (ARN) of an
  existing IAM role to bind to the add-on's service account. The role must be
  assigned the IAM permissions required by the add-on. If you don't specify
  an existing IAM role, then the add-on uses the permissions assigned to the node
  IAM role. For more information, see [Amazon EKS node IAM role](https://docs.aws.amazon.com/eks/latest/userguide/create-node-role.html)
  in the Amazon EKS User Guide.
  
  ~> **Note:** To specify an existing IAM role, you must have an IAM OpenID Connect (OIDC)
  provider created for your cluster. For more information, [see Enabling IAM roles
  for service accounts on your cluster](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html)
  in the Amazon EKS User Guide.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the EKS add-on.
* `id` - EKS Cluster name and EKS Addon name separated by a colon (`:`).
* `status` - Status of the EKS add-on.
* `created_at` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was created.
* `modified_at` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was updated.

## Import

EKS add-on can be imported using the `cluster_name` and `addon_name` separated by a colon (`:`), e.g.

```
$ terraform import aws_eks_addon.my_eks_addon my_cluster_name:my_addon_name
```
