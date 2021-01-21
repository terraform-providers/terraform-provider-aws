package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsTransferServer_basic(t *testing.T) {
	resourceName := "aws_transfer_server.test"
	datasourceName := "data.aws_transfer_server.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsTransferServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "endpoint", resourceName, "endpoint"),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_provider_type", resourceName, "identity_provider_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "logging_role", resourceName, "logging_role"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsTransferServer_service_managed(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_transfer_server.test"
	datasourceName := "data.aws_transfer_server.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsTransferServerConfig_service_managed(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "endpoint", resourceName, "endpoint"),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_provider_type", resourceName, "identity_provider_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "logging_role", resourceName, "logging_role"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsTransferServer_apigateway(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_transfer_server.test"
	datasourceName := "data.aws_transfer_server.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsTransferServerConfig_apigateway(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "endpoint", resourceName, "endpoint"),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_provider_type", resourceName, "identity_provider_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "invocation_role", resourceName, "invocation_role"),
					resource.TestCheckResourceAttrPair(datasourceName, "logging_role", resourceName, "logging_role"),
					resource.TestCheckResourceAttrPair(datasourceName, "url", resourceName, "url"),
				),
			},
		},
	})
}

const testAccDataSourceAwsTransferServerConfig_basic = `
resource "aws_transfer_server" "test" {}

data "aws_transfer_server" "test" {
  server_id = aws_transfer_server.test.id
}
`

func testAccDataSourceAwsTransferServerConfig_service_managed(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "tf-test-transfer-server-iam-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-test-transfer-server-iam-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoCloudWatchLogs",
      "Effect": "Allow",
      "Action": [
        "logs:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
  logging_role           = aws_iam_role.test.arn
}

data "aws_transfer_server" "test" {
  server_id = aws_transfer_server.test.id
}
`, rName)
}

func testAccDataSourceAwsTransferServerConfig_apigateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_name        = "test"
  description       = "%[1]s"
  stage_description = "%[1]s"

  variables = {
    "a" = "2"
  }
}

resource "aws_iam_role" "test" {
  name = "tf-test-transfer-server-iam-role-for-apigateway-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-test-transfer-server-iam-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoCloudWatchLogs",
      "Effect": "Allow",
      "Action": [
        "logs:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_region" "current" {}

resource "aws_transfer_server" "test" {
  identity_provider_type = "API_GATEWAY"
  url                    = "https://${aws_api_gateway_rest_api.test.id}.execute-api.${data.aws_region.current.name}.amazonaws.com${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.test.arn
  logging_role           = aws_iam_role.test.arn
}

data "aws_transfer_server" "test" {
  server_id = aws_transfer_server.test.id
}
`, rName)
}
