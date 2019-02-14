---
layout: "aws"
page_title: "AWS: cloudhsm_v2_hsm"
sidebar_current: "docs-aws-resource-cloudhsm-v2-hsm"
description: |-
  Provides a CloudHSM v2 HSM module resource.
---

# aws_cloudhsm_v2_hsm

Creates an HSM module in Amazon CloudHSM v2 cluster.

## Example Usage

The following example below creates an HSM module in CloudHSM cluster.

```hcl
data "aws_cloudhsm_v2_cluster" "cluster" {
  cluster_id = "${var.cloudhsm_cluster_id}"
}

resource "aws_cloudhsm_v2_hsm" "cloudhsm_v2_hsm" {
  subnet_id  = "${data.aws_cloudhsm_v2_cluster.cluster.subnet_ids[0]}"
  cluster_id = "${data.aws_cloudhsm_v2_cluster.cluster.cluster_id}"
}
```
## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of Cloud HSM v2 cluster to which HSM will be added.
* `subnet_id` - (Optional) The ID of subnet in which HSM module will be located.
* `availability_zone` - (Optional) The IDs of AZ in which HSM module will be located. Do not use together with subnet_id.
* `ip_address` - (Optional) The IP address of HSM module. Must be within the CIDR of selected subnet.

## Attributes Reference

The following attributes are exported:

* `hsm_id` - The id of the HSM module.
* `hsm_state` - The state of the HSM module.
* `hsm_eni_id` - The id of the ENI interface allocated for HSM module.
