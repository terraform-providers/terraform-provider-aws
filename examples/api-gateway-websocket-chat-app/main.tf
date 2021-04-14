#
# Terraform configuration for simple-websockets-chat-app that has the DynamoDB table and
# Lambda functions needed to demonstrate the Websocket protocol on API Gateway.
#

terraform {
  required_version = ">= 0.12"
}

#
# Providers.
#

provider "aws" {
  region = var.aws_region
}

provider "archive" {}

#
# Data sources for current AWS account ID, partition and region.
#

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

#
# DynamoDB table resources.
#

resource "aws_dynamodb_table" "ConnectionsTable" {
  name           = "simplechat_connections"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "connectionId"

  attribute {
    name = "connectionId"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }
}

# See https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-policy-template-list.html#dynamo-db-crud-policy.
resource "aws_iam_policy" "DynamoDBCrudPolicy" {
  name = "DynamoDBCrudPolicy"

  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "dynamodb:GetItem",
      "dynamodb:DeleteItem",
      "dynamodb:PutItem",
      "dynamodb:Scan",
      "dynamodb:Query",
      "dynamodb:UpdateItem",
      "dynamodb:BatchWriteItem",
      "dynamodb:BatchGetItem",
      "dynamodb:DescribeTable"
    ],
    "Resource": [
      "${aws_dynamodb_table.ConnectionsTable.arn}",
      "${aws_dynamodb_table.ConnectionsTable.arn}/index/*"
    ]
  }]
}
EOT
}

#
# WebSocket API resources.
#

resource "aws_apigatewayv2_api" "SimpleChatWebSocket" {
  name                       = "SimpleChatWebSocket"
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.message"
}

resource "aws_apigatewayv2_deployment" "Deployment" {
  api_id = aws_apigatewayv2_api.SimpleChatWebSocket.id

  depends_on = [
    aws_apigatewayv2_route.ConnectRoute,
    aws_apigatewayv2_route.DisconnectRoute,
    aws_apigatewayv2_route.SendRoute,
  ]
}

resource "aws_apigatewayv2_stage" "Stage" {
  api_id        = aws_apigatewayv2_api.SimpleChatWebSocket.id
  name          = "Prod"
  description   = "Prod Stage"
  deployment_id = aws_apigatewayv2_deployment.Deployment.id
}

###########
# OnConnect
###########
resource "aws_apigatewayv2_integration" "ConnectIntegrat" {
  api_id             = aws_apigatewayv2_api.SimpleChatWebSocket.id
  integration_type   = "AWS_PROXY"
  description        = "Connect Integration"
  integration_uri    = aws_lambda_function.OnConnectFunction.invoke_arn
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "ConnectRoute" {
  api_id         = aws_apigatewayv2_api.SimpleChatWebSocket.id
  route_key      = "$connect"
  operation_name = "ConnectRoute"
  target         = "integrations/${aws_apigatewayv2_integration.ConnectIntegrat.id}"
}

##############
# OnDisconnect
##############
resource "aws_apigatewayv2_integration" "DisconnectInteg" {
  api_id             = aws_apigatewayv2_api.SimpleChatWebSocket.id
  integration_type   = "AWS_PROXY"
  description        = "Disconnect Integration"
  integration_uri    = aws_lambda_function.OnDisconnectFunction.invoke_arn
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "DisconnectRoute" {
  api_id         = aws_apigatewayv2_api.SimpleChatWebSocket.id
  route_key      = "$disconnect"
  operation_name = "DisconnectRoute"
  target         = "integrations/${aws_apigatewayv2_integration.DisconnectInteg.id}"
}

#############
# SendMessage
#############
resource "aws_apigatewayv2_integration" "SendInteg" {
  api_id             = aws_apigatewayv2_api.SimpleChatWebSocket.id
  integration_type   = "AWS_PROXY"
  description        = "Send Integration"
  integration_uri    = aws_lambda_function.SendMessageFunction.invoke_arn
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "SendRoute" {
  api_id         = aws_apigatewayv2_api.SimpleChatWebSocket.id
  route_key      = "sendmessage"
  operation_name = "SendRoute"
  target         = "integrations/${aws_apigatewayv2_integration.SendInteg.id}"
}

#
# Lambda function resources.
#

###########
# OnConnect
###########
data "archive_file" "OnConnectZip" {
  type        = "zip"
  source_file = "${path.module}/onconnect/app.js"
  output_path = "${path.module}/onconnect/app.zip"
}

resource "aws_lambda_function" "OnConnectFunction" {
  filename      = data.archive_file.OnConnectZip.output_path
  function_name = "OnConnectFunction"
  role          = aws_iam_role.OnConnectRole.arn
  handler       = "app.handler"
  runtime       = "nodejs12.x"
  memory_size   = 256

  environment {
    variables = {
      TABLE_NAME = aws_dynamodb_table.ConnectionsTable.name
    }
  }
}

resource "aws_lambda_permission" "OnConnectPermission" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.OnConnectFunction.function_name
  principal     = "apigateway.amazonaws.com"
}

resource "aws_iam_role" "OnConnectRole" {
  name = "OnConnectRole"

  assume_role_policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOT
}

resource "aws_iam_role_policy_attachment" "OnConnectRoleDynamoDBCrudPolicyAttachment" {
  role       = aws_iam_role.OnConnectRole.name
  policy_arn = aws_iam_policy.DynamoDBCrudPolicy.arn
}

resource "aws_iam_policy" "OnConnectCloudWatchLogsPolicy" {
  name = "OnConnectCloudWatchLogsPolicy"

  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "logs:CreateLogGroup",
      "Resource": "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${aws_lambda_function.OnConnectFunction.function_name}:*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "OnConnectRoleOnConnectCloudWatchLogsPolicyAttachment" {
  role       = aws_iam_role.OnConnectRole.name
  policy_arn = aws_iam_policy.OnConnectCloudWatchLogsPolicy.arn
}

##############
# OnDisconnect
##############
data "archive_file" "OnDisconnectZip" {
  type        = "zip"
  source_file = "${path.module}/ondisconnect/app.js"
  output_path = "${path.module}/ondisconnect/app.zip"
}

resource "aws_lambda_function" "OnDisconnectFunction" {
  filename      = data.archive_file.OnDisconnectZip.output_path
  function_name = "OnDisconnectFunction"
  role          = aws_iam_role.OnDisconnectRole.arn
  handler       = "app.handler"
  runtime       = "nodejs12.x"
  memory_size   = 256

  environment {
    variables = {
      TABLE_NAME = aws_dynamodb_table.ConnectionsTable.name
    }
  }
}

resource "aws_lambda_permission" "OnDisconnectPermission" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.OnDisconnectFunction.function_name
  principal     = "apigateway.amazonaws.com"
}

resource "aws_iam_role" "OnDisconnectRole" {
  name = "OnDisconnectRole"

  assume_role_policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOT
}

resource "aws_iam_role_policy_attachment" "OnDisconnectRoleDynamoDBCrudPolicyAttachment" {
  role       = aws_iam_role.OnDisconnectRole.name
  policy_arn = aws_iam_policy.DynamoDBCrudPolicy.arn
}

resource "aws_iam_policy" "OnDisconnectCloudWatchLogsPolicy" {
  name = "OnDisconnectCloudWatchLogsPolicy"

  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "logs:CreateLogGroup",
      "Resource": "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${aws_lambda_function.OnDisconnectFunction.function_name}:*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "OnDisconnectRoleOnDisconnectCloudWatchLogsPolicyAttachment" {
  role       = aws_iam_role.OnDisconnectRole.name
  policy_arn = aws_iam_policy.OnDisconnectCloudWatchLogsPolicy.arn
}

#############
# SendMessage
#############
data "archive_file" "SendMessageZip" {
  type        = "zip"
  source_file = "${path.module}/sendmessage/app.js"
  output_path = "${path.module}/sendmessage/app.zip"
}

resource "aws_lambda_function" "SendMessageFunction" {
  filename      = data.archive_file.SendMessageZip.output_path
  function_name = "SendMessageFunction"
  role          = aws_iam_role.SendMessageRole.arn
  handler       = "app.handler"
  runtime       = "nodejs12.x"
  memory_size   = 256

  environment {
    variables = {
      TABLE_NAME = aws_dynamodb_table.ConnectionsTable.name
    }
  }
}

resource "aws_lambda_permission" "SendMessagePermission" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.SendMessageFunction.function_name
  principal     = "apigateway.amazonaws.com"
}

resource "aws_iam_role" "SendMessageRole" {
  name = "SendMessageRole"

  assume_role_policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOT
}

resource "aws_iam_role_policy_attachment" "SendMessageRoleDynamoDBCrudPolicyAttachment" {
  role       = aws_iam_role.SendMessageRole.name
  policy_arn = aws_iam_policy.DynamoDBCrudPolicy.arn
}

resource "aws_iam_policy" "SendMessageCloudWatchLogsPolicy" {
  name = "SendMessageCloudWatchLogsPolicy"

  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "logs:CreateLogGroup",
      "Resource": "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${aws_lambda_function.SendMessageFunction.function_name}:*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "SendMessageRoleSendMessageCloudWatchLogsPolicyAttachment" {
  role       = aws_iam_role.SendMessageRole.name
  policy_arn = aws_iam_policy.SendMessageCloudWatchLogsPolicy.arn
}

resource "aws_iam_policy" "SendMessageManageConnectionsPolicy" {
  name = "SendMessageManageConnectionsPolicy"

  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": "execute-api:ManageConnections",
    "Resource": "${aws_apigatewayv2_api.SimpleChatWebSocket.execution_arn}/*"
  }]
}
EOT
}

resource "aws_iam_role_policy_attachment" "SendMessageRoleSendMessageManageConnectionsPolicyAttachment" {
  role       = aws_iam_role.SendMessageRole.name
  policy_arn = aws_iam_policy.SendMessageManageConnectionsPolicy.arn
}

#
# Outputs.
#

output "ConnectionsTableArn" {
  value = aws_dynamodb_table.ConnectionsTable.arn
}

output "OnConnectFunctionArn" {
  value = aws_lambda_function.OnConnectFunction.arn
}

output "OnDisconnectFunctionArn" {
  value = aws_lambda_function.OnDisconnectFunction.arn
}

output "SendMessageFunctionArn" {
  value = aws_lambda_function.SendMessageFunction.arn
}

output "WebSocketURI" {
  value = aws_apigatewayv2_stage.Stage.invoke_url
}
