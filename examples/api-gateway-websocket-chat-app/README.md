# API Gateway WebSocket Chat Application

This example demonstrates how to create a simple chat application using [API Gateway's WebSocket-based API](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api.html) features.
It's based on the [AWS Serverless Application Model](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/what-is-sam.html) [`simple-websockets-chat-app`](https://github.com/aws-samples/simple-websockets-chat-app).

## Running this Example

Either `cp terraform.template.tfvars terraform.tfvars` and modify that new file accordingly or provide variables via CLI:

```
terraform apply -var="aws_region=us-east-1"
```

The `WebSocketURI` output contains the URL of the API endpoint to be used when following the [instructions](https://github.com/aws-samples/simple-websockets-chat-app#testing-the-chat-api) for testing.
