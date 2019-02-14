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

* `endpoint_type` - (Optional) Endpoint type. Valid values: `iot:CredentialProvider`, `iot:Data`, `iot:Data-ATS`, `iot:Job`.

## Attributes Reference

* `endpoint_address` - The endpoint based on `endpoint_type`:
  * No `endpoint_type`: Either `iot:Data` or `iot:Data-ATS` [depending on region](https://aws.amazon.com/blogs/iot/aws-iot-core-ats-endpoints/)
  * `iot:CredentialsProvider`: `IDENTIFIER.credentials.iot.REGION.amazonaws.com`
  * `iot:Data`: `IDENTIFIER.iot.REGION.amazonaws.com`
  * `iot:Data-ATS`: `IDENTIFIER-ats.iot.REGION.amazonaws.com`
  * `iot:Job`: `IDENTIFIER.jobs.iot.REGION.amazonaws.com`
