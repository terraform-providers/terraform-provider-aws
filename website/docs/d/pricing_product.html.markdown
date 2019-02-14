---
layout: "aws"
page_title: "AWS: aws_pricing_product"
sidebar_current: "docs-aws-datasource-pricing-product"
description: |-
  Get information regarding the pricing of an Amazon product
---

# Data Source: aws_pricing_product

Use this data source to get the pricing information of all products in AWS.
This data source is only available in a us-east-1 or ap-south-1 provider.

## Example Usage

```hcl
data "aws_pricing_product" "example" {
  service_code = "AmazonEC2"

  filters {
    field = "instanceType"
    value = "c5.xlarge"
  }
    
  filters {
    field = "operatingSystem"
     value = "Linux"
  }

  filters {
    field = "location"
    value = "US East (N. Virginia)"
  }

  filters {
    field = "preInstalledSw"
    value = "NA"
  }
    
  filters {
    field = "licenseModel"
    value = "No License required"
  }
    
  filters {
    field = "tenancy"
    value = "Shared"
  }
}
```

```hcl
data "aws_pricing_product" "example" {
  service_code = "AmazonRedshift"

  filters {
   field = "instanceType"
   value = "ds1.xlarge"
  }
  
  filters {
    field = "location"
    value = "US East (N. Virginia)"
  }
}
```

## Argument Reference

 * `service_code` - (Required) The code of the service. Available service codes can be fetched using the DescribeServices pricing API call.
 * `filters` - (Required) A list of filters. Passed directly to the API (see GetProducts API reference). These filters must describe a single product, this resource will fail if more than one product is returned by the API.

### filters

 * `field` (Required) The product attribute name that you want to filter on.
 * `value` (Required) The product attribute value that you want to filter on.

## Attributes Reference

 * `result` - Set to the product returned from the API.

