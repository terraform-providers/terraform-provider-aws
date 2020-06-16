---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_download_bucket_object"
description: |-
    Downloads a bucket object to be used as a local file
---

# Data Source: aws_s3_download_bucket_object

The S3 Download Bucket Object Data Source allows download of a file to the disk to be used as a local file.

## Example Usage

The following example retrieves a zip file from S3 and uses that file for a Lambda function

```hcl
data "aws_s3_download_bucket_object" "lambda_file" {
  bucket   = "ourcorp-deploy-config"
  key      = "lambda_file.zip"
  filename = "lambda_file.zip"
}

resource "aws_lambda_function" "example" {
  filename      = data.aws_s3_download_bucket_object.lambda_file.filename
  function_name = "example"
  handler       = "lambda_handler.lambda_handler"
  runtime       = "python3.7"
}
```

This is especially useful when using a S3 file to deploy a Lambda function in multiple regions. Deployment from S3 can only occur in the same region as the lambda.

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to read the object from
* `key` - (Required) The full path to the object inside the bucket
* `version_id` - (Optional) Specific version ID of the object returned (defaults to latest version)
* `filename` - (Required) The path where the file should be downloaded
