---
layout: "aws"
page_title: "AWS: cloudhsm_v2_cluster"
sidebar_current: "docs-aws-resource-cloudhsm-v2-cluster"
description: |-
  Provides a CloudHSM v2 resource.
---

# aws_cloudhsm_v2_cluster

Creates an Amazon CloudHSM v2 cluster.

For information about CloudHSM v2, see the
[AWS CloudHSM User Guide][1] and the [Amazon
CloudHSM API Reference][2].

~> **NOTE:** CloudHSM can take up to several minutes to be set up.
Practically no single attribute can be updated except TAGS.
If you need to delete a cluster, you have to remove its HSM modules first.
To initialize cluster you have to sign CSR and upload it.

## Example Usage

The following example below creates a CloudHSM cluster.

```hcl
provider "aws" {
  region = "${var.aws_region}"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "cloudhsm2_vpc" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "example-aws_cloudhsm_v2_cluster"
  }
}

resource "aws_subnet" "cloudhsm2_subnets" {
  count                   = 2
  vpc_id                  = "${aws_vpc.cloudhsm2_vpc.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = false
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "example-aws_cloudhsm_v2_cluster"
  }
}

resource "aws_cloudhsm_v2_cluster" "cloudhsm_v2_cluster" {
  hsm_type   = "hsm1.medium"
  subnet_ids = ["${aws_subnet.cloudhsm2_subnets.*.id}"]

  tags = {
    Name = "example-aws_cloudhsm_v2_cluster"
  }
}
```
## Argument Reference

The following arguments are supported:

* `source_backup_identifier` - (Optional) The id of Cloud HSM v2 cluster backup to be restored.
* `hsm_type` - (Required) The type of HSM module in the cluster. Currently, only hsm1.medium is supported.
* `subnet_ids` - (Required) The IDs of subnets in which cluster will operate.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

The following attributes are exported:

* `cluster_id` - The id of the CloudHSM cluster.
* `cluster_state` - The state of the cluster.
* `vpc_id` - The id of the VPC that the CloudHSM cluster resides in.
* `security_group_id` - The ID of the security group associated with the CloudHSM cluster.
* `cluster_certificates` - The list of cluster certificates.
  * `cluster_certificates.0.cluster_certificate` - The cluster certificate issued (signed) by the issuing certificate authority (CA) of the cluster's owner.
  * `cluster_certificates.0.cluster_csr` - The certificate signing request (CSR). Available only in UNINITIALIZED state.
  * `cluster_certificates.0.aws_hardware_certificate` - The HSM hardware certificate issued (signed) by AWS CloudHSM.
  * `cluster_certificates.0.hsm_certificate` - The HSM certificate issued (signed) by the HSM hardware.
  * `cluster_certificates.0.manufacturer_hardware_certificate` - The HSM hardware certificate issued (signed) by the hardware manufacturer.

[1]: https://docs.aws.amazon.com/cloudhsm/latest/userguide/introduction.html
[2]: https://docs.aws.amazon.com/cloudhsm/latest/APIReference/Welcome.html
