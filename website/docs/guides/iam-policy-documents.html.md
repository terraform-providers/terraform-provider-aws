---
layout: "aws"
page_title: "AWS IAM Policy Documents with Terraform"
sidebar_current: "docs-aws-guide-iam-policy-documents"
description: |-
  Using Terraform to configure AWS IAM policy documents.
---

# AWS IAM Policy Documents with Terraform

AWS leverages a standard JSON Identity and Access Management (IAM) policy document format across many services to control authorization to resources and API actions. This guide is designed to highlight some recommended configuration patterns with how Terraform and the AWS provider can build these policy documents.

The example policy documents and resources in this guide are for illustrative purposes only. Full documentation about the IAM policy format and supported elements can be found in the [AWS IAM User Guide](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements.html).

~> **NOTE:** Some AWS services only allow a subset of the policy elements or policy variables. For more information, see the AWS User Guide for the service you are configuring.

~> **NOTE:** [IAM policy variables](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_variables.html), e.g. `${aws:username}`, use the same configuration syntax (`${...}`) as Terraform interpolation. When implementing IAM policy documents with these IAM variables, you may receive syntax errors from Terraform. You can escape the dollar character within your Terraform configration to prevent the error, e.g. `$${aws:username}`.

<!-- TOC depthFrom:2 -->

- [Choosing a Configuration Method](#choosing-a-configuration-method)
- [Recommended Configuration Method Examples](#recommended-configuration-method-examples)
    - [aws_iam_policy_document Data Source](#aws_iam_policy_document-data-source)
    - [Multiple Line Heredoc Syntax](#multiple-line-heredoc-syntax)
- [Other Configuration Method Examples](#other-configuration-method-examples)
    - [Single Line String Syntax](#single-line-string-syntax)
    - [file() Interpolation Function](#file-interpolation-function)
    - [template_file Data Source](#template_file-data-source)

<!-- /TOC -->

## Choosing a Configuration Method

Terraform offers flexibility when creating configurations to match the architectural structure of teams and infrastructure. In most situations, using native functionality within Terraform and its providers will be the simplest to understand, eliminating context switching with other tooling, file sprawl, or differing file formats. Configuration examples of the available methods can be found later in the guide.

The recommended approach to building AWS IAM policy documents within Terraform is the highly customizable [`aws_iam_policy_document` data source](#aws_iam_policy_document-data-source). A short list of benefits over other methods include:

- Native Terraform configuration - no need to worry about JSON formatting or syntax
- Policy layering - create policy documents that combine and/or overwrite other policy documents
- Built-in policy error checking

Otherwise in simple cases, such as a statically defined [assume role policy for an IAM role](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_permissions-to-switch.html), Terraform's [multiple line heredoc syntax](#multiple-line-heredoc-syntax) allows the easiest formatting without any indirection of a separate data source configuration or file.

Additional methods are available, such [single line string syntax](#single-line-string-syntax), the [file() interpolation function](#file-interpolation-function), and the [template_file data source](#template_file-data-source), however their usage is discouraged due to their complexity.

## Recommended Configuration Method Examples

These configuration methods are the simplest and most powerful within Terraform.

### aws_iam_policy_document Data Source

-> For complete implementation information and examples, see the [`aws_iam_policy_document` data source documentation](/docs/providers/aws/d/iam_policy_document.html).

```hcl
data "aws_iam_policy_document" "example" {
  statement {
    actions   = ["*"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "example" {
  # ... other configuration ...

  policy = "${data.aws_iam_policy_document.example.json}"
}
```

### Multiple Line Heredoc Syntax

Interpolation is available within the heredoc string if necessary.

For example:

```hcl
resource "aws_iam_policy" "example" {
  # ... other configuration ...
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
POLICY
}
```

## Other Configuration Method Examples

These other configuration methods are provided only for reference and not meant to be an authoritative source of information.

### Single Line String Syntax

Single line IAM policy documents can be constructed with regular string syntax. Interpolation is available within the string if necessary. Since the double quotes within the IAM policy JSON conflict with Terraform's double quotes for declaring a string, they need to be escaped (`\"`).

For example:

```hcl
resource "aws_iam_policy" "example" {
  # ... other configuration ...

  policy = "{\"Version\": \"2012-10-17\", \"Statement\": {\"Effect\": \"Allow\", \"Action\": \"*\", \"Resource\": \"*\"}}"
}
```

### file() Interpolation Function

To decouple the IAM policy JSON from the Terraform configuration, Terraform has a built-in [`file()` interpolation function](/docs/configuration/interpolation.html#file-path-), which can read the contents of a local file into the configuration. Interpolation is _not_ available when using the `file()` function by itself.

For example, creating a file called `policy.json` with the contents:

```json
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
```

Those contents can be read into the Terraform configuration via:

```hcl
resource "aws_iam_policy" "example" {
  # ... other configuration ...

  policy = "${file("policy.json")}"
}
```

### template_file Data Source

To enable interpolation in decoupled files, the [`template_file` data source](/docs/providers/template/d/file.html) is available.

For example, creating a file called `policy.json.tpl` with the contents:

```json
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "${resource}"
  }
}
```

Those contents can be read and interpolated into the Terraform configuration via:

```hcl
data "template_file" "example" {
  template = "${file("policy.json.tpl")}"

  vars {
    resource = "${aws_vpc.example.arn}"
  }
}

resource "aws_iam_policy" "example" {
  # ... other configuration ...

  policy = "${data.template_file.example.rendered}"
}
```
