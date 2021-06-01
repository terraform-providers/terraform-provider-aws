# S3 Bucket Integration for API Gateway

This example demonstrates how to create an S3 Proxy using AWS API Gateway. It takes you through listing the buckets of the API caller, but provides an example of all of the resources needed to extend the API to manipulating bucket contents. It follows [this article](https://docs.aws.amazon.com/apigateway/latest/developerguide/integrating-api-with-aws-services-s3.html) on AWS.

## Running this Example

*Note: In order to see the API Gateway that this configuration creates you must navigate to the correct region in your AWS console. To ensure the API works as desired, the region you use to create the API should be different than the S3 buckets being queried. If not, you may encounter a 500 Internal Server Error response. This limitation does not apply to any deployed API.*

Only three variables are required to run this example. They can be provided by ```cp terraform.template.tfvars terraform.tfvars```, modifying ```terraform.tfvars``` with your variables, and running ```terraform apply```. Alternatively, the variables can be provided as flags by running:
```
terraform apply -var="aws_region=us-west-2"
```

