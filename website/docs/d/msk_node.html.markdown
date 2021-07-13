---
subcategory: "Managed Streaming for Kafka (MSK)"
layout: "aws"
page_title: "AWS: aws_msk_node"
description: |-
  Get information on an Amazon MSK Cluster node
---

# Data Source: aws_msk_node

Get information on an Amazon MSK Cluster node.

## Example Usage

```hcl
data "aws_msk_node" "example" {
  cluster_arn     = "example"
  broker_endpoint = "b-1.kafka-primary.0000.00.kafka.us-east-1.amazonaws.com"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_arn` - (Required) Arn of the cluster.
* `broker_endpoint` - (Optional) Broker endpoint of the node.
* `broker_id` - (Optional) Broker ID of the node.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the MSK cluster node.
* `attached_eni_id` - The attached elastic network interface of the broker
* `client_subnet` - The client subnet to which this broker node belongs
* `client_vpc_ip_address` - The virtual private cloud (VPC) of the client
* `cluster_arn` - Amazon Resource Name (ARN) of the MSK cluster.
* `instance_type` - The instance type
* `kafka_version` - The version of Apache Kafka
