---
layout: "aws"
page_title: "Decrypting AWS Lambda environment variables with KMS"
sidebar_current: "decrypt-lambda-env-vars-kms"
description: |-
  How to guide for using KMS encrypted variables in your Lambdas
---

# Context

[AWS Lambda](https://aws.amazon.com/lambda/) allows you to set environment variables. The value of these environment variables can be a [KMS](https://aws.amazon.com/kms/) encrypted payload, which you then [dynamically decrypt within your Lambda execution](https://docs.aws.amazon.com/lambda/latest/dg/tutorial-env_console.html). The only requirement is that the permissions are configured correctly in order for the Lambda to KMS decrypt that payload.

The AWS Lambda console exposes a checkbox titled **Enable helpers for encryption in transit**. This checkbox simply reveals a) a dropdown menu to select a KMS key, and b) a button called **Encrypt**. After selecting the KMS key of your choice, all the **Encrypt** button does is a call to KMS to encrypt the value of your environment variable. The resulting payload is hidden from you. After saving your Lambda with the encrypted environment variable, you can use KMS decode within your code to retrieve its plaintext value. This means you can safely store secrets on your AWS Lambda configuration. This guide is an example of how to take advantage of the above workflow with your Terraform code.
 
## TLDR

1) KMS encrypt your secrets.
2) Put them in your Lambda Terraform code, in the environment variables section.
3) Use KMS decrypt in your code and use the secrets.

Read on for a more detailed guide.

## Encrypting your environment variables values with KMS

First, you'll need to KMS encrypt the secrets that you want to put in environment variables. You can do something like this:

```
aws kms encrypt --key-id abcdefga-1234-4321-1111-1234567890ab --plaintext 'test env var encryption with kms' --output text --query CiphertextBlob
```

You can also encrypt a file -- check the `aws kms` help dialog for more options on encrypting with KMS.

The call above will return something like this:

```
AQICAHgO7ccccccccRfEHaaaaaaaaaaa/CgHyxWbbbbbbbbbbb/ccccccccccccccccccccccccccccccccG9w0BBwagbzBVVVVVgGCSqGSIb3DQEcccccccceBglghkgBZQMBBBBBM7G2BQyebrcccccccc1VH09B6AbeSSSSSvxFK5TZ61ccccccccC2HDVHDDDDDWCuPiwcCBLLLLdVO9E97cccccccc==
```

## Using the encrypted values in your Terraform

Here's an example implementation of using the encrypted payload in your Terraform. The following code creates a role with full KMS access and a Lambda using that role. The Lambda is specified with an environment variable called `foo` which has our encrypted payload as its value.

```
resource "aws_iam_policy" "kms-policy" {
  name        = "KMSFullAccess"
  path        = "/"
  description = "Give full KMS access"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "kms:*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda-policy-attach" {
    role       = "${aws_iam_role.iam_for_lambda.name}"
    policy_arn = "${aws_iam_policy.kms-policy.arn}"
}

resource "aws_lambda_function" "test_lambda" {
  filename         = "lambda_function_payload.zip"
  function_name    = "test-env-vars-kms"
  role             = "${aws_iam_role.iam_for_lambda.arn}"
  handler          = "lambda_function.lambda_handler"
  source_code_hash = "${base64sha256(file("lambda_function_payload.zip"))}"
  runtime          = "python3.6"

  ## We will put the payload in here:
  environment {
    variables = {
      foo = "AQICAHgO7ccccccccRfEHaaaaaaaaaaa/CgHyxWbbbbbbbbbbb/ccccccccccccccccccccccccccccccccG9w0BBwagbzBVVVVVgGCSqGSIb3DQEcccccccceBglghkgBZQMBBBBBM7G2BQyebrcccccccc1VH09B6AbeSSSSSvxFK5TZ61ccccccccC2HDVHDDDDDWCuPiwcCBLLLLdVO9E97cccccccc=="
    }
  }
}
```

In your code, you'll need to decrypt the environment variable via KMS. Here's a sample implementation in Python:

```
import boto3
import os

from base64 import b64decode

ENCRYPTED = os.environ['foo']
# Decrypt code should run once and variables stored outside of the function
# handler so that these are decrypted once per container
DECRYPTED = boto3.client('kms').decrypt(CiphertextBlob=b64decode(ENCRYPTED))['Plaintext'].decode("utf-8")

def lambda_handler(event, context):
    return DECRYPTED 
```

You can see examples for other languages via the AWS Lambda console, by going to **Encryption Configuration**, checking **Enable helpers for encryption in transit** and then clicking on **Code**. This button is (at the time of this writing) found to the right of your environment variables, and will default to showing sample KMS decryption code for the language your Lambda is configured with.

## Summary

The required steps to use KMS encrypted variables in AWS Lambda using Terraform is:

1) Encrypt your secrets with `aws kms encrypt`.
2) Put your encrypted secrets in your Terraform files, in the environment variables section of your Lambda.
3) Make sure your Lambda IAM Role has permissions to `decrypt` on KMS.
4) Add code to your Lambda function to decrypt the secret.
5) Use the decrypted value in your code.

Make sure to use further security best practices to keep your secrets safe! :)

