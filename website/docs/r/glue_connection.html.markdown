---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_connection"
description: |-
  Provides an Glue Connection resource.
---

# Resource: aws_glue_connection

Provides a Glue Connection resource.

## Example Usage

### Non-VPC Connection

```hcl
resource "aws_glue_connection" "example" {
  connection_properties = {
    JDBC_CONNECTION_URL = "jdbc:mysql://example.com/exampledatabase"
    PASSWORD            = "examplepassword"
    USERNAME            = "exampleusername"
  }

  name = "example"
}
```

### VPC Connection

For more information, see the [AWS Documentation](https://docs.aws.amazon.com/glue/latest/dg/populate-add-connection.html#connection-JDBC-VPC).

```hcl
resource "aws_glue_connection" "example" {
  connection_properties = {
    JDBC_CONNECTION_URL = "jdbc:mysql://${aws_rds_cluster.example.endpoint}/exampledatabase"
    PASSWORD            = "examplepassword"
    USERNAME            = "exampleusername"
  }

  name = "example"

  physical_connection_requirements {
    availability_zone      = aws_subnet.example.availability_zone
    security_group_id_list = [aws_security_group.example.id]
    subnet_id              = aws_subnet.example.id
  }
}
```

## Argument Reference

The following arguments are supported:

* `catalog_id` – (Optional) The ID of the Data Catalog in which to create the connection. If none is supplied, the AWS account ID is used by default.
* `connection_properties` – (Required) A map of key-value pairs used as parameters for this connection.
* `connection_type` – (Optional) The type of the connection. Supported are: `JDBC`, `MONGODB`, `KAFKA`, and `NETWORK`. Defaults to `JBDC`.
* `description` – (Optional) Description of the connection.
* `match_criteria` – (Optional) A list of criteria that can be used in selecting this connection.
* `name` – (Required) The name of the connection.
* `physical_connection_requirements` - (Optional) A map of physical connection requirements, such as VPC and SecurityGroup. Defined below.

### physical_connection_requirements

* `availability_zone` - (Optional) The availability zone of the connection. This field is redundant and implied by `subnet_id`, but is currently an api requirement.
* `security_group_id_list` - (Optional) The security group ID list used by the connection.
* `subnet_id` - (Optional) The subnet ID used by the connection.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Catalog ID and name of the connection
* `arn` - The ARN of the Glue Connection.

## Import

Glue Connections can be imported using the `CATALOG-ID` (AWS account ID if not custom) and `NAME`, e.g.

```
$ terraform import aws_glue_connection.MyConnection 123456789012:MyConnection
```
