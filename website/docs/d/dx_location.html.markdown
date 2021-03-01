---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_location"
description: |-
  Retrieve information about a specific AWS Direct Connect location in the current AWS Region.
---

# Data Source: aws_dx_location

Retrieve information about a specific AWS Direct Connect location in the current AWS Region.
These are the locations that can be specified when configuring [`aws_dx_connection`](/docs/providers/aws/r/dx_connection.html) or [`aws_dx_lag`](/docs/providers/aws/r/dx_lag.html) resources.

~> **Note:** This data source is different from the [`aws_dx_locations`](/docs/providers/aws/d/dx_locations.html) data source which retrieves information about all the AWS Direct Connect locations in the current AWS Region.

## Example Usage

```hcl
data "aws_dx_location" "example" {
  location_code = "CS32A-24FL"
}
```

## Argument Reference

* `location_code` - (Required) The code for the location to retrieve.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `available_port_speeds` - The available port speeds for the location.
* `available_providers` - The names of the service providers for the location.
* `location_name` - The name of the location. This includes the name of the colocation partner and the physical site of the building.
