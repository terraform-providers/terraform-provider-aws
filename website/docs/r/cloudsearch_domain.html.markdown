---
subcategory: "CloudSearch"
layout: "aws"
page_title: "AWS: aws_cloudsearch_domain"
description: |-
  Provides an CloudSearch domain resource. 
---

# Resource: aws_cloudsearch_domain

Provides an CloudSearch domain resource.

## Example Usage

```hcl
resource "aws_cloudsearch_domain" "my_domain" {
  domain_name   = "test-domain"
  instance_type = "search.m3.medium"

  indexes {
      name            = "headline"
      type            = "text"
      search          = true
      return          = true
      sort            = true
      highlight       = false
      analysis_scheme = "_en_default_"
    }
  indexes {
      name   = "price"
      type   = "double"
      search = true
      facet  = true
      return = true
      sort   = true
    }
  access_policy = data.aws_iam_policy_document.cloudsearch-access-policy.json
}

data "aws_iam_policy_document" "cloudsearch-access-policy" {
  statement {
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions = [
      "cloudsearch:search",
      "cloudsearch:suggest"
    ]
  }
}
```


## Using aws_cloudsearch_domain with API gateway.

When you are using a cloudsearch domain with an aws_api_gateway_integration you need to set uri of the AWS cloudsearch service, to a specially formatted uri, that includes the first part of the search endpoint that is returned from this provider.  The below will help.
```
data "aws_region" "current" {}
resource "aws_api_gateway_integration" "sample" {
  ...
  type        = "AWS"
  uri         = "arn:aws:apigateway:${data.aws_region.current.name}:${replace(aws_cloudsearch_domain.my_domain.searchservice, ".${data.aws_region.current.name}.cloudsearch.amazonaws.com", "")}.cloudsearch:path/2013-01-01/search"
  ...
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) Domain name
* `instance_type` - (Optional) The type of instance to start
* `replication_count` - (Optional) The amount of replicas.
* `partition_count` - (Optional) The amount of partitions on each instance. Currently only supported by `search.m3.2xlarge`
* `indexes` - (Required) See [Indexes](#indexes) below for details.
* `access_policy` - (Required) The iam access policy.
* `wait_for_endpoints` = (Optional) - Default true, wait for the search service end point.  If you set this to false, the search and doc endpoints won't be available to use as an attribute during the first run.

### Indexes

Each of the `indexes` entities represents an index field of the domain.

* `name` - (Required) Represents the property field name.
* `type` - (Required) Represents the property type. It can be one of `int,int-array,double,double-array,literal,literal-array,text,text-array,date,date-array,lation` [AWS's Docs](http://docs.aws.amazon.com/cloudsearch/latest/developerguide/configuring-index-fields.html)
* `search` - (Required) You can set whether this index should be searchable or not. (index of type text ist always searchable)
* `facet` - (Optional) You can get facet information by enabling this.
* `return` - (Required) You can enable returning the value of all searchable fields.
* `sort` - (Optional) You can enable the property to be sortable.
* `highlight` - (Optional) You can highlight information.
* `analysis_scheme` - (Optional) Only needed with type `text`. [AWS's Docs - supported languages](http://docs.aws.amazon.com/cloudsearch/latest/developerguide/text-processing.html)
* `default_value` - (Optional) The default value.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `docservice` - The doc service end point - see wait_for_endpoints parameter
* `searchservice` - The search service end point - see wait_for_endpoints parameter
* `domain_id` - The domain id
