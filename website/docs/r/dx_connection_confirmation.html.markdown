---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_connection_confirmation"
description: |-
  Provides a confirmation of the creation of the specified hosted connection on an interconnect.
---

# Resource: aws_dx_connection_confirmation

Provides a confirmation of the creation of the specified hosted connection on an interconnect.

## Example Usage

```hcl
resource "aws_dx_connection_confirmation" "confirmation" {
    connection_id = "dxcon-ffabc123"
}
```

## Argument Reference

The following arguments are supported:

* `connection_id` - (Required) The ID of the hosted connection.

### Removing `aws_dx_connection_confirmation` from your configuration

Removing an `aws_dx_connection_confirmation` resource from your configuration will remove it
from your statefile and management, **but will not destroy the Hosted Connection.**

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the connection.
