---
layout: "aws"
page_title: "AWS: cloudhsm_v2_cluster"
sidebar_current: "docs-aws-datasource-cloudhsm-v2-cluster"
description: |-
  Get information on a CloudHSM v2 cluster.
---

# Data Source: aws_cloudhsm_v2_cluster

Use this data source to get information about a CloudHSM v2 cluster

## Example Usage

```hcl
data "aws_cloudhsm_v2_cluster" "cluster" {
  cluster_id = "cluster-testclusterid"
}
```
## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The id of Cloud HSM v2 cluster.
* `cluster_state` - (Optional) The state of the cluster to be found.

## Attributes Reference

The following attributes are exported:

* `vpc_id` - The id of the VPC that the CloudHSM cluster resides in.
* `security_group_id` - The ID of the security group associated with the CloudHSM cluster.
* `subnet_ids` - The IDs of subnets in which cluster operates.
* `cluster_certificates` - The list of cluster certificates.
  * `cluster_certificates.0.cluster_certificate` - The cluster certificate issued (signed) by the issuing certificate authority (CA) of the cluster's owner.
  * `cluster_certificates.0.cluster_csr` - The certificate signing request (CSR). Available only in UNINITIALIZED state.
  * `cluster_certificates.0.aws_hardware_certificate` - The HSM hardware certificate issued (signed) by AWS CloudHSM.
  * `cluster_certificates.0.hsm_certificate` - The HSM certificate issued (signed) by the HSM hardware.
  * `cluster_certificates.0.manufacturer_hardware_certificate` - The HSM hardware certificate issued (signed) by the hardware manufacturer.
The number of available cluster certificates may vary depending on state of the cluster.
