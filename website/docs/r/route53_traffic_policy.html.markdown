---
subcategory: "Route53"
layout: "aws"
page_title: "AWS: aws_route53_traffic_policy"
description: |-
  Manages a Route53 Traffic Policy
---

# Resource: aws_route53_traffic_policy

Manages a Route53 Traffic Policy.

## Example Usage

```hcl
resource "aws_route53_traffic_policy" "example" {
  name     = "example"
  comment  = "example comment"
  document = <<EOF
{
  "AWSPolicyFormatVersion": "2015-10-01",
  "RecordType": "A",
  "Endpoints": {
    "endpoint-start-NkPh": {
      "Type": "value",
      "Value": "10.0.0.2"
    }
  },
  "StartEndpoint": "endpoint-start-NkPh"
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) This is the name of the traffic policy.
* `comment` - (Optional) A comment for the traffic policy.
* `document` - (Required) The policy document. This is a JSON formatted string. For more information about building Route53 traffic policy documents, see the [AWS Route53 Traffic Policy document format](https://docs.aws.amazon.com/Route53/latest/APIReference/api-policies-traffic-policy-document-format.html)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the traffic policy
* `latest_version` - The latest version number of the traffic policy. This value is automatically incremented by AWS after each update of this resource.


## Import

Route53 Traffic Policy can be imported using the `id` and `latest_version`, e.g.

```
$ terraform import aws_route53_traffic_policy.example 01a52019-d16f-422a-ae72-c306d2b6df7e/1
```
