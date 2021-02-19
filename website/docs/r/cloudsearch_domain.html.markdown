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
  instance_type = "search.medium"

  index {
    name            = "headline"
    type            = "text"
    search          = true
    return          = true
    sort            = true
    highlight       = false
    analysis_scheme = "_en_default_"
  }

  index {
    name   = "price"
    type   = "double"
    search = true
    facet  = true
    return = true
    sort   = true
  }
  service_access_policies = data.aws_iam_policy_document.cloudsearch_access_policy.json
}

data "aws_iam_policy_document" "cloudsearch_access_policy" {
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
  uri         = "arn:aws:apigateway:${data.aws_region.current.name}:${replace(aws_cloudsearch_domain.my_domain.search_endpoint, ".${data.aws_region.current.name}.cloudsearch.amazonaws.com", "")}.cloudsearch:path/2013-01-01/search"
  ...
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the CloudSearch domain.
* `instance_type` - (Optional) The type of instance to start.
* `replication_count` - (Optional) The amount of replicas.
* `partition_count` - (Optional) The amount of partitions on each instance. Currently only supported by `search.2xlarge`.
* `index` - (Required) See [Indices](#indices) below for details.
* `service_access_policies` - (Required) The AWS IAM access policy.
* `wait_for_endpoints` - (Optional) - Default true, wait for the search service end point.  If you set this to false, the search and document endpoints won't be available to use as an attribute during the first run.

### Indices

Each of the `index` entities represents an index field of the domain.

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

* `document_endpoint` - The doc service end point - see wait_for_endpoints parameter
* `search_endpoint` - The search service end point - see wait_for_endpoints parameter
* `domain_id` - The domain id
