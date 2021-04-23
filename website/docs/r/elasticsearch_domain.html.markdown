---
subcategory: "ElasticSearch"
layout: "aws"
page_title: "AWS: aws_elasticsearch_domain"
description: |-
  Terraform resource for managing an AWS Elasticsearch Domain.
---

# Resource: aws_elasticsearch_domain

Manages an AWS Elasticsearch Domain.

## Example Usage

### Basic Usage

```hcl
resource "aws_elasticsearch_domain" "example" {
  domain_name           = "example"
  elasticsearch_version = "1.5"

  cluster_config {
    instance_type = "r4.large.elasticsearch"
  }

  snapshot_options {
    automated_snapshot_start_hour = 23
  }

  tags = {
    Domain = "TestDomain"
  }
}
```

### Access Policy

-> See also: [`aws_elasticsearch_domain_policy` resource](/docs/providers/aws/r/elasticsearch_domain_policy.html)

```hcl
variable "domain" {
  default = "tf-test"
}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_elasticsearch_domain" "example" {
  domain_name = var.domain

  # ... other configuration ...

  access_policies = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "es:*",
      "Principal": "*",
      "Effect": "Allow",
      "Resource": "arn:aws:es:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:domain/${var.domain}/*",
      "Condition": {
        "IpAddress": {"aws:SourceIp": ["66.193.100.22/32"]}
      }
    }
  ]
}
POLICY
}
```

### Log Publishing to CloudWatch Logs

```hcl
resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

resource "aws_cloudwatch_log_resource_policy" "example" {
  policy_name = "example"

  policy_document = <<CONFIG
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "es.amazonaws.com"
      },
      "Action": [
        "logs:PutLogEvents",
        "logs:PutLogEventsBatch",
        "logs:CreateLogStream"
      ],
      "Resource": "arn:aws:logs:*"
    }
  ]
}
CONFIG
}

resource "aws_elasticsearch_domain" "example" {
  # .. other configuration ...

  log_publishing_options {
    cloudwatch_log_group_arn = aws_cloudwatch_log_group.example.arn
    log_type                 = "INDEX_SLOW_LOGS"
  }
}
```

### VPC based ES

```hcl
variable "vpc" {}

variable "domain" {
  default = "tf-test"
}

data "aws_vpc" "selected" {
  tags = {
    Name = var.vpc
  }
}

data "aws_subnet_ids" "selected" {
  vpc_id = data.aws_vpc.selected.id

  tags = {
    Tier = "private"
  }
}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_security_group" "es" {
  name        = "${var.vpc}-elasticsearch-${var.domain}"
  description = "Managed by Terraform"
  vpc_id      = data.aws_vpc.selected.id

  ingress {
    from_port = 443
    to_port   = 443
    protocol  = "tcp"

    cidr_blocks = [
      data.aws_vpc.selected.cidr_block,
    ]
  }
}

resource "aws_iam_service_linked_role" "es" {
  aws_service_name = "es.amazonaws.com"
}

resource "aws_elasticsearch_domain" "es" {
  domain_name           = var.domain
  elasticsearch_version = "6.3"

  cluster_config {
    instance_type = "m4.large.elasticsearch"
  }

  vpc_options {
    subnet_ids = [
      data.aws_subnet_ids.selected.ids[0],
      data.aws_subnet_ids.selected.ids[1],
    ]

    security_group_ids = [aws_security_group.es.id]
  }

  advanced_options = {
    "rest.action.multi.allow_explicit_index" = "true"
  }

  access_policies = <<CONFIG
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "es:*",
			"Principal": "*",
			"Effect": "Allow",
			"Resource": "arn:aws:es:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:domain/${var.domain}/*"
		}
	]
}
CONFIG

  snapshot_options {
    automated_snapshot_start_hour = 23
  }

  tags = {
    Domain = "TestDomain"
  }

  depends_on = [aws_iam_service_linked_role.es]
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) Name of the domain.
* `access_policies` - (Optional) IAM policy document specifying the access policies for the domain
* `advanced_options` - (Optional) Key-value string pairs to specify advanced configuration options.
   Note that the values for these configuration options must be strings (wrapped in quotes) or they
   may be wrong and cause a perpetual diff, causing Terraform to want to recreate your Elasticsearch
   domain on every apply.
* `advanced_security_options` - (Optional) Options for [fine-grained access control](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/fgac.html). See below for more details.
* `ebs_options` - (Optional) EBS related options, may be required based on chosen [instance size](https://aws.amazon.com/elasticsearch-service/pricing/). See below.
* `encrypt_at_rest` - (Optional) Encrypt at rest options. Only available for [certain instance types](http://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/aes-supported-instance-types.html). See below.
* `node_to_node_encryption` - (Optional) Node-to-node encryption options. See below.
* `cluster_config` - (Optional) Cluster configuration of the domain, see below.
* `snapshot_options` - (Optional) Snapshot related options, see below.
* `vpc_options` - (Optional) VPC related options, see below. Adding or removing this configuration forces a new resource ([documentation](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-vpc.html#es-vpc-limitations)).
* `log_publishing_options` - (Optional) Options for publishing slow  and application logs to CloudWatch Logs. This block can be declared multiple times, for each log_type, within the same resource.
* `cognito_options` - (Optional) Options for authenticating Kibana with Cognito. See below.
* `elasticsearch_version` - (Optional) The version of Elasticsearch to deploy. Defaults to `1.5`
* `domain_endpoint_options` - (Optional) Domain endpoint HTTP(S) related options. See below.
* `tags` - (Optional) A map of tags to assign to the resource

The **advanced_security_options** block supports the following attributes:

* `enabled` - (Required, Forces new resource) Whether advanced security is enabled
* `internal_user_database_enabled` - (Optional, Default: false) Whether the internal user database is enabled. If not set, defaults to `false` by the AWS API.
* `master_user_options` - (Optional) Credentials for the master user: username and password, or ARN
    * `master_user_arn` - (Optional) ARN for the master user. Only specify if `internal_user_database_enabled` is not set or set to `false`)
    * `master_user_name` - (Optional) The master user's username, which is stored in the Amazon Elasticsearch Service domain's internal database. Only specify if `internal_user_database_enabled` is set to `true`.
    * `master_user_password` - (Optional) The master user's password, which is stored in the Amazon Elasticsearch Service domain's internal database. Only specify if `internal_user_database_enabled` is set to `true`.

**ebs_options** supports the following attributes:

* `ebs_enabled` - (Required) Whether EBS volumes are attached to data nodes in the domain.
* `volume_type` - (Optional) The type of EBS volumes attached to data nodes.
* `volume_size` - The size of EBS volumes attached to data nodes (in GiB).
**Required** if `ebs_enabled` is set to `true`.
* `iops` - (Optional) The baseline input/output (I/O) performance of EBS volumes attached to data nodes. Applicable only for the Provisioned IOPS EBS volume type.

**encrypt_at_rest** supports the following attributes:

* `enabled` - (Required) Whether to enable encryption at rest. If the `encrypt_at_rest` block is not provided then this defaults to `false`.
* `kms_key_id` - (Optional) The KMS key id to encrypt the Elasticsearch domain with. If not specified then it defaults to using the `aws/es` service KMS key.

**domain_endpoint_options** supports the following attributes:

* `enforce_https` - (Required) Whether or not to require HTTPS
* `tls_security_policy` - (Optional) The name of the TLS security policy that needs to be applied to the HTTPS endpoint. Valid values:  `Policy-Min-TLS-1-0-2019-07` and `Policy-Min-TLS-1-2-2019-07`. Terraform will only perform drift detection if a configuration value is provided.

**cluster_config** supports the following attributes:

* `instance_type` - (Optional) Instance type of data nodes in the cluster.
* `instance_count` - (Optional) Number of instances in the cluster.
* `dedicated_master_enabled` - (Optional) Indicates whether dedicated master nodes are enabled for the cluster.
* `dedicated_master_type` - (Optional) Instance type of the dedicated master nodes in the cluster.
* `dedicated_master_count` - (Optional) Number of dedicated master nodes in the cluster
* `zone_awareness_config` - (Optional) Configuration block containing zone awareness settings. Documented below.
* `zone_awareness_enabled` - (Optional) Indicates whether zone awareness is enabled, set to `true` for multi-az deployment. To enable awareness with three Availability Zones, the `availability_zone_count` within the `zone_awareness_config` must be set to `3`.
* `warm_enabled` - (Optional) Indicates whether to enable warm storage.
* `warm_count` - (Optional) The number of warm nodes in the cluster. Valid values are between `2` and `150`. `warm_count` can be only and must be set when `warm_enabled` is set to `true`.
* `warm_type` - (Optional) The instance type for the Elasticsearch cluster's warm nodes. Valid values are `ultrawarm1.medium.elasticsearch`, `ultrawarm1.large.elasticsearch` and `ultrawarm1.xlarge.elasticsearch`. `warm_type` can be only and must be set when `warm_enabled` is set to `true`.

**zone_awareness_config** supports the following attributes:

* `availability_zone_count` - (Optional) Number of Availability Zones for the domain to use with `zone_awareness_enabled`. Defaults to `2`. Valid values: `2` or `3`.

**node_to_node_encryption** supports the following attributes:

* `enabled` - (Required) Whether to enable node-to-node encryption. If the `node_to_node_encryption` block is not provided then this defaults to `false`.

**vpc_options** supports the following attributes:

AWS documentation: [VPC Support for Amazon Elasticsearch Service Domains](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-vpc.html)

**Note** you must have created the service linked role for the Elasticsearch service to use the `vpc_options`.
If you need to create the service linked role at the same time as the Elasticsearch domain then you must use `depends_on` to make sure that the role is created before the Elasticsearch domain.
See the [VPC based ES domain example](#vpc-based-es) above.

* `security_group_ids` - (Optional) List of VPC Security Group IDs to be applied to the Elasticsearch domain endpoints. If omitted, the default Security Group for the VPC will be used.
* `subnet_ids` - (Required) List of VPC Subnet IDs for the Elasticsearch domain endpoints to be created in.

Security Groups and Subnets referenced in these attributes must all be within the same VPC; this determines what VPC the endpoints are created in.

**snapshot_options** supports the following attribute:

* `automated_snapshot_start_hour` - (Required) Hour during which the service takes an automated daily snapshot of the indices in the domain.

**log_publishing_options** supports the following attribute:

* `log_type` - (Required) A type of Elasticsearch log. Valid values: INDEX_SLOW_LOGS, SEARCH_SLOW_LOGS, ES_APPLICATION_LOGS, AUDIT_LOGS
* `cloudwatch_log_group_arn` - (Required) ARN of the Cloudwatch log group to which log needs to be published.
* `enabled` - (Optional, Default: true) Specifies whether given log publishing option is enabled or not.

**cognito_options** supports the following attribute:

AWS documentation: [Amazon Cognito Authentication for Kibana](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-cognito-auth.html)

* `enabled` - (Optional, Default: false) Specifies whether Amazon Cognito authentication with Kibana is enabled or not
* `user_pool_id` - (Required) ID of the Cognito User Pool to use
* `identity_pool_id` - (Required) ID of the Cognito Identity Pool to use
* `role_arn` - (Required) ARN of the IAM role that has the AmazonESCognitoAccess policy attached

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the domain.
* `domain_id` - Unique identifier for the domain.
* `domain_name` - The name of the Elasticsearch domain.
* `endpoint` - Domain-specific endpoint used to submit index, search, and data upload requests.
* `kibana_endpoint` - Domain-specific endpoint for kibana without https scheme.
* `vpc_options.0.availability_zones` - If the domain was created inside a VPC, the names of the availability zones the configured `subnet_ids` were created inside.
* `vpc_options.0.vpc_id` - If the domain was created inside a VPC, the ID of the VPC.

## Timeouts

`aws_elasticsearch_domain` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `update` - (Optional, Default: `60m`) How long to wait for updates.

## Import

Elasticsearch domains can be imported using the `domain_name`, e.g.

```
$ terraform import aws_elasticsearch_domain.example domain_name
```
