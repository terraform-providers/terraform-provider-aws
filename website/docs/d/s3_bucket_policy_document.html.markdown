---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_policy_document"
description: |-
    Provides details about a specific S3 bucket
---

# Data Source: aws_s3_bucket_policy_document

Creates a policy document for a bucket without using heredocs.

This resource is useful for creating any kind of bucket policy document without using heredocs.

## Example Usage

### Public Bucket Redirect

Creates a public bucket that redirects to another bucket with `redirect_all_requests_to` and a non-heredoc public bucket policy.

```
data "aws_iam_policy_document" "public-read" {
  statement {
    sid    = "PublicReadGetObject"
    effect = "Allow"
    actions = [
      "s3:GetObject"
    ]

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    resources = [
      "arn:aws:s3:::${var.bucket_name}/*"
    ]
  }
}

resource "aws_s3_bucket" "s3_www_bucket" {
  bucket = var.bucket_name
  acl    = "public-read"

  website {
    redirect_all_requests_to = "http://example.com"
  }

  tags = var.tags
}


resource "aws_s3_bucket_policy" "public-read" {
  bucket = aws_s3_bucket.s3_www_bucket.bucket
  policy = data.aws_iam_policy_document.public-read.json
}
```

## Argument Reference

The following arguments are supported:

* `policy_id` (Optional) - An ID for the policy document.
* `source_json` (Optional) - An IAM policy document to import as a base for the
  current policy document.  Statements with non-blank `sid`s in the current
  policy document will overwrite statements with the same `sid` in the source
  json.  Statements without an `sid` cannot be overwritten.
* `override_json` (Optional) - An IAM policy document to import and override the
  current policy document.  Statements with non-blank `sid`s in the override
  document will overwrite statements with the same `sid` in the current document.
  Statements without an `sid` cannot be overwritten.
* `statement` (Optional) - A nested configuration block (described below)
  configuring one *statement* to be included in the policy document.
* `version` (Optional) - IAM policy document version. Valid values: `2008-10-17`, `2012-10-17`. Defaults to `2012-10-17`. For more information, see the [AWS IAM User           Guide](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_version.html).

Each document configuration may have one or more `statement` blocks, which
each accept the following arguments:

* `sid` (Optional) - An ID for the policy statement.
* `effect` (Optional) - Either "Allow" or "Deny", to specify whether this
  statement allows or denies the given actions. The default is "Allow".
* `actions` (Optional) - A list of actions that this statement either allows
  or denies. For example, ``["ec2:RunInstances", "s3:*"]``.
* `not_actions` (Optional) - A list of actions that this statement does *not*
  apply to. Used to apply a policy statement to all actions *except* those
  listed.
* `resources` (Optional) - A list of resource ARNs that this statement applies
  to. This is required by AWS if used for an IAM policy.
* `not_resources` (Optional) - A list of resource ARNs that this statement
  does *not* apply to. Used to apply a policy statement to all resources
  *except* those listed.
* `principals` (Optional) - A nested configuration block (described below)
  specifying a principal (or principal pattern) to which this statement applies.
* `not_principals` (Optional) - Like `principals` except gives principals that
  the statement does *not* apply to.
* `condition` (Optional) - A nested configuration block (described below)
  that defines a further, possibly-service-specific condition that constrains
  whether this statement applies.

Each policy may have either zero or more `principals` blocks or zero or more
`not_principals` blocks, both of which each accept the following arguments:

* `type` (Required) The type of principal. For AWS ARNs this is "AWS".  For AWS services (e.g. Lambda), this is "Service". For Federated access the type is "Federated".
* `identifiers` (Required) List of identifiers for principals. When `type`
  is "AWS", these are IAM user or role ARNs.  When `type` is "Service", these are AWS Service roles e.g. `lambda.amazonaws.com`. When `type` is "Federated", these are web      identity users or SAML provider ARNs.

For further examples or information about AWS principals then please refer to the [documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/                             reference_policies_elements_principal.html).

Each policy statement may have zero or more `condition` blocks, which each
accept the following arguments:

* `test` (Required) The name of the
  [IAM condition operator](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_condition_operators.html)
  to evaluate.
* `variable` (Required) The name of a
  [Context Variable](http://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements.html#AvailableKeys)
  to apply the condition to. Context variables may either be standard AWS
  variables starting with `aws:`, or service-specific variables prefixed with
  the service name.
* `values` (Required) The values to evaluate the condition against. If multiple
  values are provided, the condition matches if at least one of them applies.
  (That is, the tests are combined with the "OR" boolean operation.)

When multiple `condition` blocks are provided, they must *all* evaluate to true
for the policy statement to apply. (In other words, the conditions are combined
with the "AND" boolean operation.)

## Context Variable Interpolation

The IAM policy document format allows context variables to be interpolated
into various strings within a statement. The native IAM policy document format
uses `${...}`-style syntax that is in conflict with Terraform's interpolation
syntax, so this data source instead uses `&{...}` syntax for interpolations that
should be processed by AWS rather than by Terraform.

## Wildcard Principal

In order to define wildcard principal (a.k.a. anonymous user) use `type = "*"` and
`identifiers = ["*"]`. In that case the rendered json will contain `"Principal": "*"`.
Note, that even though the [IAM Documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html)
states that `"Principal": "*"` and `"Principal": {"AWS": "*"}` are equivalent,
those principals have different behavior for IAM Role Trust Policy. Therefore
Terraform will normalize the principal field only in above-mentioned case and principals
like `type = "AWS"` and `identifiers = ["*"]` will be rendered as `"Principal": {"AWS": "*"}`.

## Attributes Reference

The following attribute is exported:

* `json` - The above arguments serialized as a standard JSON policy document.


