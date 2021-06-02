---
subcategory: "Resource Groups Tagging API"
layout: "aws"
page_title: "AWS: aws_resourcegroupstaggingapi_resources"
description: |-
  Provides details about resource tagging.
---

# Data Source: aws_resourcegroupstaggingapi_resources

Provides details about resource tagging.

## Example Usage

### Get All Resource Tag Mappings

```terraform
data "aws_resourcegroupstaggingapi_resources" "test" {}
```

### Filter By Tag Key and Value

```terraform
data "aws_resourcegroupstaggingapi_resources" "test" {
  tag_filter {
    key    = "tag-key"
    values = ["tag-value-1", "tag-value-2"]
  }
}
```

### Filter By Resource Type

```terraform
data "aws_resourcegroupstaggingapi_resources" "test" {
  resource_type_filters = ["ec2:instance"]
}
```


## Argument Reference

The following arguments are supported:

* `exclude_compliant_resources` - (Optional) Specifies whether to exclude resources that are compliant with the tag policy. You can use this parameter only if the `include_compliance_details` argument is also set to `true`.
* `include_compliance_details` - (Optional) Specifies whether to include details regarding the compliance with the effective tag policy.
* `tag_filter` - (Optional) Specifies a list of Tag Filters (keys and values) to restrict the output to only those resources that have the specified tag and, if included, the specified value. See [Tag Filter](#tag-filter) below. Conflicts with `resource_arn_list`.
* `resource_type_filters` - (Optional) The constraints on the resources that you want returned. The format of each resource type is `service:resourceType`. For example, specifying a resource type of `ec2` returns all Amazon EC2 resources (which includes EC2 instances). Specifying a resource type of `ec2:instance` returns only EC2 instances.
* `resource_arn_list` - (Optional) Specifies a list of ARNs of resources for which you want to retrieve tag data. Conflicts with `filter`.

### Tag Filter

A `tag_filter` block supports the following arguments:

If you do specify `tag_filter`, the response returns only those resources that are currently associated with the specified tag.
If you don't specify a `tag_filter`, the response includes all resources that were ever associated with tags. Resources that currently don't have associated tags are shown with an empty tag set.

* `key` - (Required) One part of a key-value pair that makes up a tag.
* `values` - (Optional) The optional part of a key-value pair that make up a tag.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `resource_tag_mapping_list` - List of objects matching the search criteria.
    * `compliance_details` - List of objects with information that shows whether a resource is compliant with the effective tag policy, including details on any noncompliant tag keys.
        * `compliance_status` - Whether the resource is compliant.
        * `keys_with_noncompliant_values ` - Set of tag keys with non-compliant tag values.
        * `non_compliant_keys ` - Set of non-compliant tag keys.
    * `resource_arn` - ARN of the resource.
    * `tags` - Map of tags assigned to the resource.
