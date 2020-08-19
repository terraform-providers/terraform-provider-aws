---
subcategory: "MediaStore"
layout: "aws"
page_title: "AWS: aws_media_store_container_lifecycle_policy"
description: |-
  Provides a MediaStore Container Lifecycle Policy.
---

# Resource: aws_media_store_container_lifecycle_policy

Provides a MediaStore Container Lifecycle Policy.

## Example Usage

```hcl
resource "aws_media_store_container" "example" {
  name = "example"
}

resource "aws_media_store_container_lifecycle_policy" "example" {
  container_name = "${aws_media_store_container.example.name}"

  policy = <<EOF
{
    "rules": [
         {
            "definition": {
                "path": [ 
                    {"prefix": "Football/"}, 
                    {"prefix": "Baseball/"}
                ],
                "days_since_create": [
                    {"numeric": [">" , 28]}
                ]
            },
            "action": "EXPIRE"
        },
        {
            "definition": {
                "path": [ { "prefix": "AwardsShow/" }  ],
                "days_since_create": [
                    {"numeric": [">=" , 15]}
                ]
            },
            "action": "EXPIRE"
        },
        {
            "definition": {
                "path": [ { "prefix": "" }  ],
                "days_since_create": [
                    {"numeric": [">" , 40]}
                ]
            },
            "action": "EXPIRE"
        },
        {
            "definition": {
                "path": [ { "wildcard": "Football/*.ts" }  ],
                "days_since_create": [
                    {"numeric": [">" , 20]}
                ]
            },
            "action": "EXPIRE"
        },
        {
            "definition": {
                "path": [ 
                    {"wildcard": "Football/index*.m3u8"}
                ],
                "seconds_since_create": [
                    {"numeric": [">" , 15]}
                ]
            },
            "action": "EXPIRE"
        }
    ]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `container_name` - (Required) The name of the container.
* `policy` - (Required) A JSON formatted Lifecycle Policy. For more information about Lifecycle Policy, see the [Object Lifecycle Policies in AWS Elemental MediaStore](https://docs.aws.amazon.com/mediastore/latest/ug/policies-object-lifecycle.html)

## Import

MediaStore Container Lifecycle Policy can be imported using the MediaStore Container Name, e.g.

```
$ terraform import aws_media_store_container_lifecycle_policy.example example
```
