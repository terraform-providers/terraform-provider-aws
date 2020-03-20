---
subcategory: "Managed Streaming for Kafka (MSK)"
layout: "aws"
page_title: "AWS: aws_msk_cluster"
description: |-
  Get information on an Amazon MSK Cluster
---

# Data Source: aws_msk_cluster

Get information on an Amazon MSK Cluster.

## Example Usage

```hcl
data "aws_msk_cluster" "example" {
  cluster_name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_name` - (Required) Name of the cluster.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the MSK cluster.
* `bootstrap_brokers` - A comma separated list of one or more hostname:port pairs of Kafka brokers suitable to boostrap connectivity to the Kafka cluster.
* `bootstrap_brokers_tls` - A comma separated list of one or more DNS names (or IPs) and TLS port pairs kafka brokers suitable to boostrap connectivity to the kafka cluster.
* `kafka_version` - Apache Kafka version.
* `number_of_broker_nodes` - Number of broker nodes in the cluster.
* `tags` - Map of key-value pairs assigned to the cluster.
* `zookeeper_connect_string` - A comma separated list of one or more hostname:port pairs to use to connect to the Apache Zookeeper cluster.
