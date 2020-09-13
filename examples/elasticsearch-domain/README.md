# Elasticsearch Domain Example

This example creates an Elasticsearch Domain with [fine-grained access control](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/fgac.html) enabled

## Fine-Grained Access Control Requirements
- Elasticsearch 6.7 or later
- Encryption of data at rest and node-to-node encryption enabled
- Require HTTPS for all traffic to the domain enabled
Source: https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/fgac.html#fgac-enabling

**Note:** Fine-grained access control can only be enabled on new domains. Once enabled, it cannot be disabled.

## Configure AWS Provider

To run, configure your AWS provider as described in https://www.terraform.io/docs/providers/aws/index.html

## Run the Example

For planning phase

```bash
terraform plan
```

For apply phase

```bash
terraform apply
```

## Remove the Example

To remove the stack

```bash
 terraform destroy
```