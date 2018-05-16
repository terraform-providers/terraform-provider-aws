---
layout: "aws"
page_title: "AWS: aws_iot_endpoint"
sidebar_current: "docs-aws-datasource-iot-endpoint"
description: |-
  Get the unique IoT endpoint
---

# Data Source: aws_iot_endpoint

Returns a unique endpoint specific to the AWS account making the call.

## Example Usage

```hcl
data "aws_iot_endpoint" "example" {}

resource "kubernetes_pod" "agent" {
  metadata {
    name = "my-device"
  }
  spec {
    container {
      image = "gcr.io/my-project/image-name"
      name  = "image-name"

      env = [
        {
          name  = "IOT_ENDPOINT"
          value = "${data.aws_iot_endpoint.example.endpoint_address}"
        },
      ]
    }
  }
}
```

## Argument Reference

N/A

## Attributes Reference

* `endpoint_address` - The endpoint. The format of the endpoint is as follows: `IDENTIFIER.iot.REGION.amazonaws.com`.
