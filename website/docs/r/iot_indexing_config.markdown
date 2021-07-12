---
layout: "aws"
page_title: "AWS: aws_iot_indexing_config"
sidebar_current: "docs-aws-resource-iot-indexing-config"
description: |-
    Manages AWS IoT fleet indexing configuration.
---

# Resource: aws_iot_indexing_config

Manages AWS IoT fleet indexing configuration.

## Example Usage

```hcl
resource "aws_iot_indexing_config" "foo" {
  thing_group_indexing_enabled = true
  thing_connectivity_indexing_enabled = true
  thing_indexing_mode = "REGISTRY_AND_SHADOW"
}
```

## Argument Reference

* `thing_group_indexing_enabled` - (Optional, Defaults to false) Whether to enable indexing of thing groups.
* `thing_connectivity_indexing_enabled` - (Optional, Defaults to false) Whether to enable indexing of things connectivity.
* `thing_indexing_mode` - (Optional, Defaults to "OFF") The thing indexing mode. Can be either "OFF", "REGISTRY" or "REGISTRY_AND_SHADOW".


## Import

The IoT indexing configuration can be imported using the word `iot-indexing-config`, e.g.

```
$ terraform import aws_iot_indexing_config.demo iot-indexing-config
```