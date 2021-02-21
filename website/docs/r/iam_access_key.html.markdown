---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_access_key"
description: |-
  Provides an IAM access key. This is a set of credentials that allow API requests to be made as an IAM user.
---

# Resource: aws_iam_access_key

Provides an IAM access key. This is a set of credentials that allow API requests to be made as an IAM user.

## Example Usage

```hcl
resource "aws_iam_access_key" "lb" {
  user    = aws_iam_user.lb.name
  pgp_key = "keybase:some_person_that_exists"
}

resource "aws_iam_user" "lb" {
  name = "loadbalancer"
  path = "/system/"
}

resource "aws_iam_user_policy" "lb_ro" {
  name = "test"
  user = aws_iam_user.lb.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

output "secret" {
  value = aws_iam_access_key.lb.encrypted_secret
}
```

```hcl
resource "aws_iam_user" "test" {
  name = "test"
  path = "/test/"
}

resource "aws_iam_access_key" "test" {
  user = aws_iam_user.test.name
}

output "aws_iam_smtp_password_v4" {
  value = aws_iam_access_key.test.ses_smtp_password_v4
}
```

## Argument Reference

The following arguments are supported:

* `user` - (Required) The IAM user to associate with this access key.
* `pgp_key` - (Optional) Either a base-64 encoded PGP public key, or a
  keybase username in the form `keybase:some_person_that_exists`, for use
  in the `encrypted_secret` output attribute.
* `status` - (Optional) The access key status to apply. Defaults to `Active`.
Valid values are `Active` and `Inactive`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_date` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the access key was created.
* `id` - The access key ID.
* `user` - The IAM user associated with this access key.
* `key_fingerprint` - The fingerprint of the PGP key used to encrypt the secret. This attribute is not available for imported resources.
* `secret` - The secret access key. This attribute is not available for imported resources. Note that this will be written to the state file. If you use this, please protect your backend state file judiciously. Alternatively, you may supply a `pgp_key` instead, which will prevent the secret from being stored in plaintext, at the cost of preventing the use of the secret key in automation.
* `encrypted_secret` - The encrypted secret, base64 encoded, if `pgp_key` was specified. This attribute is not available for imported resources. The encrypted secret may be decrypted using the command line, for example: `terraform output -raw encrypted_secret | base64 --decode | keybase pgp decrypt`.
* `ses_smtp_password_v4` - The secret access key converted into an SES SMTP password by applying [AWS's documented Sigv4 conversion algorithm](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/smtp-credentials.html#smtp-credentials-convert). This attribute is not available for imported resources. As SigV4 is region specific, valid Provider regions are `ap-south-1`, `ap-southeast-2`, `eu-central-1`, `eu-west-1`, `us-east-1` and `us-west-2`. See current [AWS SES regions](https://docs.aws.amazon.com/general/latest/gr/rande.html#ses_region).

## Import

IAM Access Keys can be imported using the identifier, e.g.

```
$ terraform import aws_iam_access_key.example AKIA1234567890
```

Resource attributes such as `encrypted_secret`, `key_fingerprint`, `pgp_key`, `secret`, and `ses_smtp_password_v4` are not available for imported resources as this information cannot be read from the IAM API.
