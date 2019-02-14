---
layout: "aws"
page_title: "AWS: aws_db_event_categories"
sidebar_current: "docs-aws-datasource-db-event-categories"
description: |-
    Provides a list of DB Event Categories which can be used to pass values into DB Event Subscription.
---

# Data Source: aws_db_event_categories

## Example Usage

List the event categories of all the RDS resources. 

```hcl
data "aws_db_event_categories" "example" {}

output "example" {
  value = "${data.aws_db_event_categories.example.event_categories}"
}
```

List the event categories specific to the RDS resource `db-snapshot`.

```hcl
data "aws_db_event_categories" "example" {
  source_type = "db-snapshot"
}

output "example" {
  value = "${data.aws_db_event_categories.example.event_categories}"
}
```

## Argument Reference

The following arguments are supported:

* `source_type` - (Optional) The type of source that will be generating the events. Valid options are db-instance, db-security-group, db-parameter-group, db-snapshot, db-cluster or db-cluster-snapshot.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `event_categories` - A list of the event categories.
