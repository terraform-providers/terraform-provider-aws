---
subcategory: "ElasticSearch"
layout: "aws"
page_title: "AWS: aws_elasticsearch_domain"
description: |-
  Get information on an ElasticSearch Domain resource.
---

# Data Source: aws_elasticsearch_domain

Use this data source to get information about an Elasticsearch Domain

## Example Usage

```hcl
data "aws_elasticsearch_domain" "my_domain" {
  domain_name = "my-domain-name"
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` – (Required) Name of the domain.


## Attributes Reference

The following attributes are exported:

* `access_policies` – The policy document attached to the domain.
* `advanced_options` - Key-value string pairs to specify advanced configuration options.
* `advanced_security_options` - Status of the Elasticsearch domain's advanced security options. The block consists of the following attributes:
    * `enabled` - Whether advanced security is enabled.
    * `internal_user_database_enabled` - Whether the internal user database is enabled.
* `arn` – The Amazon Resource Name (ARN) of the domain.
* `cluster_config` - Cluster configuration of the domain.
    * `instance_type` - Instance type of data nodes in the cluster.
    * `instance_count` - Number of instances in the cluster.
    * `dedicated_master_enabled` - Indicates whether dedicated master nodes are enabled for the cluster.
    * `dedicated_master_type` - Instance type of the dedicated master nodes in the cluster.
    * `dedicated_master_count` - Number of dedicated master nodes in the cluster.
    * `zone_awareness_enabled` - Indicates whether zone awareness is enabled.
    * `zone_awareness_config` - Configuration block containing zone awareness settings.
        * `availability_zone_count` - Number of availability zones used.
    * `warm_enabled` - Indicates warm storage is enabled.
    * `warm_count` - The number of warm nodes in the cluster.
    * `warm_type` - The instance type for the Elasticsearch cluster's warm nodes.
* `cognito_options` - Domain Amazon Cognito Authentication options for Kibana.
    * `enabled` - Whether Amazon Cognito Authentication is enabled.
    * `user_pool_id` - The Cognito User pool used by the domain.
    * `identity_pool_id` - The Cognito Identity pool used by the domain.
    * `role_arn` - The IAM Role with the AmazonESCognitoAccess policy attached.
* `created` – Status of the creation of the domain.
* `deleted` – Status of the deletion of the domain.
* `domain_id` – Unique identifier for the domain.
* `ebs_options` - EBS Options for the instances in the domain.
    * `ebs_enabled` - Whether EBS volumes are attached to data nodes in the domain.
    * `volume_type` - The type of EBS volumes attached to data nodes.
    * `volume_size` - The size of EBS volumes attached to data nodes (in GB).
    * `iops` - The baseline input/output (I/O) performance of EBS volumes attached to data nodes.
* `elasticsearch_version` – ElasticSearch version for the domain.
* `encryption_at_rest` - Domain encryption at rest related options.
    * `enabled` - Whether encryption at rest is enabled in the domain.
    * `kms_key_id` - The KMS key id used to encrypt data at rest.
* `endpoint` – Domain-specific endpoint used to submit index, search, and data upload requests.
* `kibana_endpoint` - Domain-specific endpoint used to access the Kibana application.
* `log_publishing_options` - Domain log publishing related options.
    * `log_type` - The type of Elasticsearch log being published.
    * `cloudwatch_log_group_arn` - The CloudWatch Log Group where the logs are published.
    * `enabled` - Whether log publishing is enabled.
* `node_to_node_encryption` - Domain in transit encryption related options.
    * `enabled` - Whether node to node encryption is enabled.
* `processing` – Status of a configuration change in the domain.
* `snapshot_options` – Domain snapshot related options.
    * `automated_snapshot_start_hour` - Hour during which the service takes an automated daily snapshot of the indices in the domain.
* `tags` - The tags assigned to the domain.
* `vpc_options` - VPC Options for private Elasticsearch domains.
    * `availability_zones` - The availability zones used by the domain.
    * `security_group_ids` - The security groups used by the domain.
    * `subnet_ids` - The subnets used by the domain.
    * `vpc_id` - The VPC used by the domain.
