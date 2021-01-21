---
subcategory: "SSM"
layout: "aws"
page_title: "AWS: aws_ssm_document"
description: |-
  Provides an SSM Document resource
---

# Resource: aws_ssm_document

Provides an SSM Document resource

~> **NOTE on updating SSM documents:** Only documents with a schema version of 2.0
or greater can update their content once created, see [SSM Schema Features][1]. To update a document with an older
schema version you must recreate the resource.

## Example Usage

```hcl
resource "aws_ssm_document" "foo" {
  name          = "test_document"
  document_type = "Command"

  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the document.
* `attachments_source` - (Optional) One or more configuration blocks describing attachments sources to a version of a document. Defined below.
* `content` - (Required) The JSON or YAML content of the document.
* `document_format` - (Optional, defaults to JSON) The format of the document. Valid document types include: `JSON` and `YAML`
* `document_type` - (Required) The type of the document. Valid document types include: `Automation`, `Command`, `Package`, `Policy`, and `Session`
* `permissions` - (Optional) Additional Permissions to attach to the document. See [Permissions](#permissions) below for details.
* `target_type` - (Optional) The target type which defines the kinds of resources the document can run on. For example, /AWS::EC2::Instance. For a list of valid resource types, see AWS Resource Types Reference (http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html)
* `tags` - (Optional) A map of tags to assign to the object.

## attachments_source

The `attachments_source` block supports the following:

* `key` - (Required) The key describing the location of an attachment to a document. Valid key types include: `SourceUrl` and `S3FileUrl`
* `values` - (Required) The value describing the location of an attachment to a document
* `name` - (Optional) The name of the document attachment file

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `created_date` - The date the document was created.
* `description` - The description of the document.
* `schema_version` - The schema version of the document.
* `default_version` - The default version of the document.
* `document_version` - The document version.
* `hash` - The sha1 or sha256 of the document content
* `hash_type` - "Sha1" "Sha256". The hashing algorithm used when hashing the content.
* `latest_version` - The latest version of the document.
* `owner` - The AWS user account of the person who created the document.
* `status` - "Creating", "Active" or "Deleting". The current status of the document.
* `parameter` - The parameters that are available to this document.
* `platform_types` - A list of OS platforms compatible with this SSM document, either "Windows" or "Linux".

[1]: http://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-ssm-docs.html#document-schemas-features

## Permissions

The permissions attribute specifies how you want to share the document. If you share a document privately,
you must specify the AWS user account IDs for those people who can use the document. If you share a document
publicly, you must specify All as the account ID.

The permissions mapping supports the following:

* `type` - The permission type for the document. The permission type can be `Share`.
* `account_ids` - The AWS user accounts that should have access to the document. The account IDs can either be a group of account IDs or `All`.

## Import

SSM Documents can be imported using the name, e.g.

```
$ terraform import aws_ssm_document.example example
```

The `attachments_source` argument does not have an SSM API method for reading the attachment information detail after creation. If the argument is set in the Terraform configuration on an imported resource, Terraform will always show a difference. To workaround this behavior, either omit the argument from the Terraform configuration or use [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to hide the difference, e.g.

```hcl
resource "aws_ssm_document" "test" {
  name          = "test_document"
  document_type = "Package"

  attachments_source {
    key    = "SourceUrl"
    values = ["s3://${aws_s3_bucket.object_bucket.bucket}/test.zip"]
  }

  # There is no AWS SSM API for reading attachments_source info directly
  lifecycle {
    ignore_changes = [attachments_source]
  }
}
```
