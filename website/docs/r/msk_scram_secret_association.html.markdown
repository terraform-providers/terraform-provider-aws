---
subcategory: "Managed Streaming for Kafka (MSK)"
layout: "aws"
page_title: "AWS: aws_msk_scram_secret_association"
description: |-
  Associates SCRAM secrets with a Managed Streaming for Kafka (MSK) cluster.
---

# Resource: aws_msk_scram_secret_association

Associates SCRAM secrets stored in the Secrets Manager service with a Managed Streaming for Kafka (MSK) cluster.

-> **Note:** The following assumes the MSK cluster has SASL/SCRAM authentication enabled. See below for example usage or refer to the [Username/Password Authentication](https://docs.aws.amazon.com/msk/latest/developerguide/msk-password.html) section of the MSK Developer Guide for more details.

To set up username and password authentication for a cluster, create an [`aws_secretsmanager_secret` resource](/docs/providers/aws/r/secretsmanager_secret.html) and associate
a username and password with the secret with an [`aws_secretsmanager_secret_version` resource](/docs/providers/aws/r/secretsmanager_secret_version.html). When creating a secret for the cluster,
the `name` must have the prefix `AmazonMSK_` and you must either use an existing custom AWS KMS key or create a new
custom AWS KMS key for your secret with the [`aws_kms_key` resource](/docs/providers/aws/r/kms_key.html). It is important to note that a policy is required for the `aws_secretsmanager_secret`
resource in order for Kafka to be able to read it. This policy is attached automatically when the `aws_msk_scram_secret_association` is used,
however, this policy will not be in terraform and as such, will present a diff on plan/apply. For that reason, you must use the [`aws_secretsmanager_secret_policy`
resource](/docs/providers/aws/r/secretsmanager_secret_policy.html) as shown below in order to ensure that the state is in a clean state after the creation of secret and the association to the cluster.

## Example Usage

```hcl
resource "aws_msk_scram_secret_association" "example" {
  cluster_arn     = aws_msk_cluster.example.arn
  secret_arn_list = [aws_secretsmanager_secret.example.arn]

  depends_on = [aws_secretsmanager_secret_version.example]
}

resource "aws_msk_cluster" "example" {
  cluster_name = "example"
  # ... other configuration...
  client_authentication {
    sasl {
      scram = true
    }
  }
}

resource "aws_secretsmanager_secret" "example" {
  name       = "AmazonMSK_example"
  kms_key_id = aws_kms_key.example.key_id
}

resource "aws_kms_key" "example" {
  description = "Example Key for MSK Cluster Scram Secret Association"
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id     = aws_secretsmanager_secret.example.id
  secret_string = jsonencode({ username = "user", password = "pass" })
}

resource "aws_secretsmanager_secret_policy" "example" {
  secret_arn = aws_secretsmanager_secret.example.arn
  policy     = <<POLICY
{
  "Version" : "2012-10-17",
  "Statement" : [ {
    "Sid": "AWSKafkaResourcePolicy",
    "Effect" : "Allow",
    "Principal" : {
      "Service" : "kafka.amazonaws.com"
    },
    "Action" : "secretsmanager:getSecretValue",
    "Resource" : "${aws_secretsmanager_secret.example.arn}"
  } ]
}
POLICY
}
```

## Argument Reference

The following arguments are supported:

* `cluster_arn` - (Required, Forces new resource) Amazon Resource Name (ARN) of the MSK cluster.
* `secret_arn_list` - (Required) List of AWS Secrets Manager secret ARNs.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the MSK cluster.

## Import

MSK SCRAM Secret Associations can be imported using the `id` e.g.

```
$ terraform import aws_msk_scram_secret_association.example arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3
```
