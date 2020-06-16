---
layout: "aws"
page_title: "AWS: aws_personalize_dataset_group"
description: |-
  Creates a dataset group.
---

# Resource: aws_personalize_dataset_group

Creates a dataset group. A dataset group contains related datasets that supply data for training a model.

## Example Usage

### Basic dataset group

```hcl
resource "aws_personalize_dataset_group" "group" {
  name = "mydatasetgroup"
}
```

### Dataset group with KMS

```hcl
resource "aws_kms_key" "key" {}

resource "aws_iam_role" "role" {
  name               = "mydataset-assume"
  assume_role_policy = "${data.aws_iam_policy_document.assume_policy.json}"
}

data "aws_iam_policy_document" "assume_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["personalize.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "policy" {
  name   = "mydataset-policy"
  policy = "${data.aws_iam_policy_document.policy.json}"
}

resource "aws_iam_role_policy_attachment" "attach" {
  role       = "${aws_iam_role.role.name}"
  policy_arn = "${aws_iam_policy.policy.arn}"
}

data "aws_iam_policy_document" "policy" {
  statement {
    actions = [
      "kms:*",
    ]

    resources = [
      "${aws_kms_key.key.arn}",
    ]
  }
}

resource "aws_personalize_dataset_group" "group" {
  name = "mydataset"

  kms {
    key_arn  = "${aws_kms_key.key.arn}"
    role_arn = "${aws_iam_role.role.arn}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name`        (Required) - The name of the dataset group
* `kms`         (Optional) - KMS configuration to encrypt the datasets (documented below)

The `kms` object supports the following:

* `key_arn`     (Required) - The Amazon Resource Name (ARN) of a KMS key used to encrypt the datasets
* `role_arn`    (Required) - The ARN of the IAM role that has permissions to access the KMS key

## Import

Personalize dataset groups can be imported using the dataset group name or the full ARN.

### Import by name
```
$ terraform import aws_personalize_dataset_group.group mydataset
```

### Import by ARN
```
$ terraform import aws_personalize_dataset_group.group arn:aws:personalize:eu-west-1:xxxxxxxxxxxx:dataset-group/mydataset
```