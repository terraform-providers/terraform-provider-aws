---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_group"
description: |-
  Get information on a Amazon QuickSight group
---

# Data Source: aws_quicksight_group

This data source can be used to fetch information about a specific
QuickSight group. By using this data source, you can reference QuickSight group
properties without having to hard code ARNs or unique IDs as input.

## Example Usage

```hcl
data "aws_quicksight_group" "example" {
  group_name 	 = "an_example_group_name"
  aws_account_id = "aws_account_id"
  namespace		 = "namespace"
}
```

## Argument Reference

* `group_name` - (Required) The name of the group that you want to match.
* `aws_account_id` - (Required) The ID for the AWS account that the group is in. Currently, you use the ID for the AWS account that contains your Amazon QuickSight account.
* `namespace` - (Required) The namespace. Currently, you should set this to default.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) for the group.
* `description` - The group description.
* `group_id` - The principal ID of the group.
