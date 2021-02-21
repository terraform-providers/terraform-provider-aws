---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_policy_document"
description: |-
  Generates an IAM policy document in JSON format
---

# Data Source: aws_iam_policy_document

Generates an IAM policy document in JSON format for use with resources that expect policy documents such as [`aws_iam_policy`](/docs/providers/aws/r/iam_policy.html).

Using this data source to generate policy documents is *optional*. It is also valid to use literal JSON strings in your configuration or to use the `file` interpolation function to read a raw JSON policy document from a file.

~> **NOTE:** AWS's IAM policy document syntax allows for replacement of [policy variables](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_variables.html) within a statement using `${...}`-style notation, which conflicts with Terraform's interpolation syntax. In order to use AWS policy variables with this data source, use `&{...}` notation for interpolations that should be processed by AWS rather than by Terraform.

-> For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).

## Example Usage

### Basic Example

```hcl
data "aws_iam_policy_document" "example" {
  statement {
    sid = "1"

    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]

    resources = [
      "arn:aws:s3:::*",
    ]
  }

  statement {
    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:aws:s3:::${var.s3_bucket_name}",
    ]

    condition {
      test     = "StringLike"
      variable = "s3:prefix"

      values = [
        "",
        "home/",
        "home/&{aws:username}/",
      ]
    }
  }

  statement {
    actions = [
      "s3:*",
    ]

    resources = [
      "arn:aws:s3:::${var.s3_bucket_name}/home/&{aws:username}",
      "arn:aws:s3:::${var.s3_bucket_name}/home/&{aws:username}/*",
    ]
  }
}

resource "aws_iam_policy" "example" {
  name   = "example_policy"
  path   = "/"
  policy = data.aws_iam_policy_document.example.json
}
```

### Example Assume-Role Policy with Multiple Principals

You can specify multiple principal blocks with different types. You can also use this data source to generate an assume-role policy.

```hcl
data "aws_iam_policy_document" "event_stream_bucket_role_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["firehose.amazonaws.com"]
    }

    principals {
      type        = "AWS"
      identifiers = [var.trusted_role_arn]
    }

    principals {
      type        = "Federated"
      identifiers = ["arn:aws:iam::${var.account_id}:saml-provider/${var.provider_name}", "cognito-identity.amazonaws.com"]
    }
  }
}
```

### Example Using A Source Document

```hcl
data "aws_iam_policy_document" "source" {
  statement {
    actions   = ["ec2:*"]
    resources = ["*"]
  }

  statement {
    sid = "SidToOverride"

    actions   = ["s3:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "source_json_example" {
  source_json = data.aws_iam_policy_document.source.json

  statement {
    sid = "SidToOverride"

    actions = ["s3:*"]

    resources = [
      "arn:aws:s3:::somebucket",
      "arn:aws:s3:::somebucket/*",
    ]
  }
}
```

`data.aws_iam_policy_document.source_json_example.json` will evaluate to:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "ec2:*",
      "Resource": "*"
    },
    {
      "Sid": "SidToOverride",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::somebucket/*",
        "arn:aws:s3:::somebucket"
      ]
    }
  ]
}
```

### Example Using An Override Document

```hcl
data "aws_iam_policy_document" "override" {
  statement {
    sid = "SidToOverride"

    actions   = ["s3:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "override_json_example" {
  override_json = data.aws_iam_policy_document.override.json

  statement {
    actions   = ["ec2:*"]
    resources = ["*"]
  }

  statement {
    sid = "SidToOverride"

    actions = ["s3:*"]

    resources = [
      "arn:aws:s3:::somebucket",
      "arn:aws:s3:::somebucket/*",
    ]
  }
}
```

`data.aws_iam_policy_document.override_json_example.json` will evaluate to:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "ec2:*",
      "Resource": "*"
    },
    {
      "Sid": "SidToOverride",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": "*"
    }
  ]
}
```

### Example with Both Source and Override Documents

You can also combine `source_json` and `override_json` in the same document.

```hcl
data "aws_iam_policy_document" "source" {
  statement {
    sid       = "OverridePlaceholder"
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "override" {
  statement {
    sid       = "OverridePlaceholder"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "politik" {
  source_json   = data.aws_iam_policy_document.source.json
  override_json = data.aws_iam_policy_document.override.json
}
```

`data.aws_iam_policy_document.politik.json` will evaluate to:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "OverridePlaceholder",
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "*"
    }
  ]
}
```

### Example of Merging Source Documents

Multiple documents can be combined using the `source_policy_documents` or `override_policy_documents` attributes. `source_policy_documents` requires that all documents have unique Sids, while `override_policy_documents` will iteratively override matching Sids.

```hcl
data "aws_iam_policy_document" "source_one" {
  statement {
    actions   = ["ec2:*"]
    resources = ["*"]
  }

  statement {
    sid = "UniqueSidOne"

    actions   = ["s3:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "source_two" {
  statement {
    sid = "UniqueSidTwo"

    actions   = ["iam:*"]
    resources = ["*"]
  }

  statement {
    actions   = ["lambda:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "combined" {
  source_policy_documents = [
    data.aws_iam_policy_document.source_one.json,
    data.aws_iam_policy_document.source_two.json
  ]
}
```

`data.aws_iam_policy_document.combined.json` will evaluate to:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "ec2:*",
      "Resource": "*"
    },
    {
      "Sid": "UniqueSidOne",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": "*"
    },
    {
      "Sid": "UniqueSidTwo",
      "Effect": "Allow",
      "Action": "iam:*",
      "Resource": "*"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "lambda:*",
      "Resource": "*"
    }
  ]
}
```

### Example of Merging Override Documents

```hcl
data "aws_iam_policy_document" "policy_one" {
  statement {
    sid    = "OverridePlaceHolderOne"
    effect = "Allow"

    actions   = ["s3:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "policy_two" {
  statement {
    effect    = "Allow"
    actions   = ["ec2:*"]
    resources = ["*"]
  }

  statement {
    sid    = "OverridePlaceHolderTwo"
    effect = "Allow"

    actions   = ["iam:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "policy_three" {
  statement {
    sid    = "OverridePlaceHolderOne"
    effect = "Deny"

    actions   = ["logs:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "combined" {
  override_policy_documents = [
    data.aws_iam_policy_document.policy_one.json,
    data.aws_iam_policy_document.policy_two.json,
    data.aws_iam_policy_document.policy_three.json
  ]

  statement {
    sid    = "OverridePlaceHolderTwo"
    effect = "Deny"

    actions   = ["*"]
    resources = ["*"]
  }
}
```

`data.aws_iam_policy_document.combined.json` will evaluate to:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "OverridePlaceholderTwo",
      "Effect": "Allow",
      "Action": "iam:*",
      "Resource": "*"
    },
    {
      "Sid": "OverridePlaceholderOne",
      "Effect": "Deny",
      "Action": "logs:*",
      "Resource": "*"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "ec2:*",
      "Resource": "*"
    },
  ]
}
```

## Argument Reference

The following arguments are optional:

* `override_json` (Optional) - IAM policy document whose statements with non-blank `sid`s will override statements with the same `sid` from documents assigned to the `source_json`, `source_policy_documents`, and `override_policy_documents` arguments. Non-overriding statements will be added to the exported document.

~> **NOTE:** Statements without a `sid` cannot be overridden. In other words, a statement without a `sid` from documents assigned to the `source_json` or `source_policy_documents` arguments cannot be overridden by statements from documents assigned to the `override_json` or `override_policy_documents` arguments.

* `override_policy_documents` (Optional) - List of IAM policy documents that are merged together into the exported document. In merging, statements with non-blank `sid`s will override statements with the same `sid` from earlier documents in the list. Statements with non-blank `sid`s will also override statements with the same `sid` from documents provided in the `source_json` and `source_policy_documents` arguments.  Non-overriding statements will be added to the exported document.
* `policy_id` (Optional) - ID for the policy document.
* `source_json` (Optional) - IAM policy document used as a base for the exported policy document. Statements with the same `sid` from documents assigned to the `override_json` and `override_policy_documents` arguments will override source statements.
* `source_policy_documents` (Optional) - List of IAM policy documents that are merged together into the exported document. Statements defined in `source_policy_documents` or `source_json` must have unique `sid`s. Statements with the same `sid` from documents assigned to the `override_json` and `override_policy_documents` arguments will override source statements.
* `statement` (Optional) - Configuration block for a policy statement. Detailed below.
* `version` (Optional) - IAM policy document version. Valid values are `2008-10-17` and `2012-10-17`. Defaults to `2012-10-17`. For more information, see the [AWS IAM User Guide](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_version.html).

### `statement`

The following arguments are optional:

* `actions` (Optional) - List of actions that this statement either allows or denies. For example, `["ec2:RunInstances", "s3:*"]`.
* `condition` (Optional) - Configuration block for a condition. Detailed below.
* `effect` (Optional) - Whether this statement allows or denies the given actions. Valid values are `Allow` and `Deny`. Defaults to `Allow`.
* `not_actions` (Optional) - List of actions that this statement does *not* apply to. Use to apply a policy statement to all actions *except* those listed.
* `not_principals` (Optional) - Like `principals` except these are principals that the statement does *not* apply to.
* `not_resources` (Optional) - List of resource ARNs that this statement does *not* apply to. Use to apply a policy statement to all resources *except* those listed.
* `principals` (Optional) - Configuration block for principals. Detailed below.
* `resources` (Optional) - List of resource ARNs that this statement applies to. This is required by AWS if used for an IAM policy.
* `sid` (Optional) - Sid (statement ID) is an identifier for a policy statement.

### `condition`

A `condition` constrains whether a statement applies in a particular situation. Conditions can be specific to an AWS service. When using multiple `condition` blocks, they must *all* evaluate to true for the policy statement to apply. In other words, AWS evaluates the conditions as though with an "AND" boolean operation.

The following arguments are required:

* `test` (Required) Name of the [IAM condition operator](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_condition_operators.html) to evaluate.
* `values` (Required) Values to evaluate the condition against. If multiple values are provided, the condition matches if at least one of them applies. That is, AWS evaluates multiple values as though using an "OR" boolean operation.
* `variable` (Required) Name of a [Context Variable](http://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements.html#AvailableKeys) to apply the condition to. Context variables may either be standard AWS variables starting with `aws:` or service-specific variables prefixed with the service name.

### `principals` and `not_principals`

The `principals` and `not_principals` arguments define to whom a statement applies or does not apply, respectively.

~> **NOTE**: Even though the [IAM Documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html) states that `"Principal": "*"` and `"Principal": {"AWS": "*"}` are equivalent, those principal elements have different behavior in some situations, e.g. IAM Role Trust Policy. To have Terraform render JSON containing `"Principal": "*"`, use `type = "*"` and `identifiers = ["*"]`. To have Terraform render JSON containing `"Principal": {"AWS": "*"}`, use `type = "AWS"` and `identifiers = ["*"]`.

-> For more information about AWS principals, refer to the [AWS Identity and Access Management User Guide: AWS JSON policy elements: Principal](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html).

The following arguments are required:

* `identifiers` (Required) List of identifiers for principals. When `type` is `AWS`, these are IAM principal ARNs, e.g. `arn:aws:iam::12345678901:role/yak-role`.  When `type` is `Service`, these are AWS Service roles, e.g. `lambda.amazonaws.com`. When `type` is `Federated`, these are web identity users or SAML provider ARNs, e.g. `accounts.google.com` or `arn:aws:iam::12345678901:saml-provider/yak-saml-provider`. When `type` is `CanonicalUser`, these are [canonical user IDs](https://docs.aws.amazon.com/general/latest/gr/acct-identifiers.html#FindingCanonicalId), e.g. `79a59df900b949e55d96a1e698fbacedfd6e09d98eacf8f8d5218e7cd47ef2be`.
* `type` (Required) Type of principal. Valid values include `AWS`, `Service`, `Federated`, and `CanonicalUser`.

## Attributes Reference

The following attribute is exported:

* `json` - Standard JSON policy document rendered based on the arguments above.
