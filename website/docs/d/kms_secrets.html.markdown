---
subcategory: "KMS"
layout: "aws"
page_title: "AWS: aws_kms_secrets"
description: |-
    Decrypt multiple secrets from data encrypted with the AWS KMS service
---

# Data Source: aws_kms_secrets

Decrypt multiple secrets from data encrypted with the AWS KMS service.

~> **NOTE**: Using this data provider will allow you to conceal secret data within your resource definitions but does not take care of protecting that data in all Terraform logging and state output. Please take care to secure your secret data beyond just the Terraform configuration.

## Example Usage

If you do not already have a `CiphertextBlob` from encrypting a KMS secret, you can use the below commands to obtain one using the [AWS CLI kms encrypt](https://docs.aws.amazon.com/cli/latest/reference/kms/encrypt.html) command. This requires you to have your AWS CLI setup correctly and replace the `--key-id` with your own. Alternatively you can use `--plaintext 'master-password'` (CLIv1) or `--plaintext fileb://<(echo -n 'master-password')` (CLIv2) instead of reading from a file.

-> If you have a newline character at the end of your file, it will be decrypted with this newline character intact. For most use cases this is undesirable and leads to incorrect passwords or invalid values, as well as possible changes in the plan. Be sure to use `echo -n` if necessary.

```sh
$ echo -n 'master-password' > plaintext-password
$ aws kms encrypt --key-id ab123456-c012-4567-890a-deadbeef123 --plaintext fileb://plaintext-password --encryption-context foo=bar --output text --query CiphertextBlob
AQECAHgaPa0J8WadplGCqqVAr4HNvDaFSQ+NaiwIBhmm6qDSFwAAAGIwYAYJKoZIhvcNAQcGoFMwUQIBADBMBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDI+LoLdvYv8l41OhAAIBEIAfx49FFJCLeYrkfMfAw6XlnxP23MmDBdqP8dPp28OoAQ==
```

That encrypted output can now be inserted into Terraform configurations without exposing the plaintext secret directly.

```hcl
data "aws_kms_secrets" "example" {
  secret {
    # ... potentially other configuration ...
    name    = "master_password"
    payload = "AQECAHgaPa0J8WadplGCqqVAr4HNvDaFSQ+NaiwIBhmm6qDSFwAAAGIwYAYJKoZIhvcNAQcGoFMwUQIBADBMBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDI+LoLdvYv8l41OhAAIBEIAfx49FFJCLeYrkfMfAw6XlnxP23MmDBdqP8dPp28OoAQ=="

    context = {
      foo = "bar"
    }
  }

  secret {
    # ... potentially other configuration ...
    name    = "master_username"
    payload = "AQECAHgaPa0J8WadplGCqqVAr4HNvDaFSQ+NaiwIBhmm6qDSFwAAAGIwYAYJKoZIhvcNAQcGoFMwUQIBADBMBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDI+LoLdvYv8l41OhAAIBEIAfx49FFJCLeYrkfMfAw6XlnxP23MmDBdqP8dPp28OoAQ=="
  }
}

resource "aws_rds_cluster" "example" {
  # ... other configuration ...
  master_password = data.aws_kms_secrets.example.plaintext["master_password"]
  master_username = data.aws_kms_secrets.example.plaintext["master_username"]
}
```

## Argument Reference

The following arguments are supported:

* `secret` - (Required) One or more encrypted payload definitions from the KMS service. See the Secret Definitions below.

### Secret Definitions

Each `secret` supports the following arguments:

* `name` - (Required) The name to export this secret under in the attributes.
* `payload` - (Required) Base64 encoded payload, as returned from a KMS encrypt operation.
* `context` - (Optional) An optional mapping that makes up the Encryption Context for the secret.
* `grant_tokens` (Optional) An optional list of Grant Tokens for the secret.

For more information on `context` and `grant_tokens` see the [KMS
Concepts](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `plaintext` - Map containing each `secret` `name` as the key with its decrypted plaintext value
