# EC2 Transit Gateway Cross-Account Peering Attachment

This example demonstrates how to peer two Transit Gateways in different regions. The peer transit gateway can be in your account or a different AWS account. Refer to AWS documentation for supported regions.

See [more in the Transit Gateway Peering Attachment documentation](https://docs.aws.amazon.com/vpc/latest/tgw/tgw-peering.html).

## Running this example

Either `cp terraform.template.tfvars terraform.tfvars` and modify that new file accordingly or provide variables via CLI:

```
terraform apply \
	-var="aws_first_access_key=AAAAAAAAAAAAAAAAAAA" \
	-var="aws_first_secret_key=SuperSecretKeyForAccount1" \
	-var="aws_second_access_key=BBBBBBBBBBBBBBBBBBB" \
	-var="aws_second_secret_key=SuperSecretKeyForAccount2" \
	-var="aws_first_region=us-east-2" \
	-var="aws_second_region=us-west-2"
```

## Prerequisites

- This example requires two AWS accounts within the same AWS Organizations Organization
- Ensure Resource Access Manager is enabled in your organization. For more information, see the [Resource Access Manager User Guide](https://docs.aws.amazon.com/ram/latest/userguide/getting-started-sharing.html).
