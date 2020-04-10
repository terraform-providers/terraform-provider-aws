package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSAPIGatewayV2Integration_basic(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetIntegrationOutput
	resourceName := "aws_apigatewayv2_integration.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2IntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_id", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_method", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", "${integration.response.statuscode}"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "MOCK"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "29000"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2IntegrationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Integration_disappears(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetIntegrationOutput
	resourceName := "aws_apigatewayv2_integration.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2IntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					testAccCheckAWSAPIGatewayV2IntegrationDisappears(&apiId, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Integration_IntegrationTypeHttp(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetIntegrationOutput
	resourceName := "aws_apigatewayv2_integration.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2IntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_integrationTypeHttp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_id", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", "${integration.response.statuscode}"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", "http://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", "$request.body.name"),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "28999"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_integrationTypeHttpUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_id", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", "CONVERT_TO_BINARY"),
					resource.TestCheckResourceAttr(resourceName, "credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test HTTP updated"),
					resource.TestCheckResourceAttr(resourceName, "integration_method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", "${integration.response.statuscode}"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", "http://www.example.org"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_TEMPLATES"),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", "#set($number=42)"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/xml", "#set($percent=$number/100)"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", "$request.body.id"),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "51"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2IntegrationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Integration_Lambda(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetIntegrationOutput
	resourceName := "aws_apigatewayv2_integration.test"
	callerIdentityDatasourceName := "data.aws_caller_identity.current"
	lambdaResourceName := "aws_lambda_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2IntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_lambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttrPair(resourceName, "credentials_arn", callerIdentityDatasourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test Lambda"),
					resource.TestCheckResourceAttr(resourceName, "integration_method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", "${integration.response.body.errorMessage}"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "AWS"),
					resource.TestCheckResourceAttrPair(resourceName, "integration_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "29000"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2IntegrationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Integration_VpcLink(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetIntegrationOutput
	resourceName := "aws_apigatewayv2_integration.test"
	vpcLinkResourceName := "aws_api_gateway_vpc_link.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2IntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_vpcLink(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttrPair(resourceName, "connection_id", vpcLinkResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "VPC_LINK"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test VPC Link"),
					resource.TestCheckResourceAttr(resourceName, "integration_method", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "HTTP_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", "http://www.example.net"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "NEVER"),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "12345"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2IntegrationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayV2IntegrationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_integration" {
			continue
		}

		_, err := conn.GetIntegration(&apigatewayv2.GetIntegrationInput{
			ApiId:         aws.String(rs.Primary.Attributes["api_id"]),
			IntegrationId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 integration %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGatewayV2IntegrationDisappears(apiId *string, v *apigatewayv2.GetIntegrationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.DeleteIntegration(&apigatewayv2.DeleteIntegrationInput{
			ApiId:         apiId,
			IntegrationId: v.IntegrationId,
		})

		return err
	}
}

func testAccCheckAWSAPIGatewayV2IntegrationExists(n string, vApiId *string, v *apigatewayv2.GetIntegrationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 integration ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		apiId := aws.String(rs.Primary.Attributes["api_id"])
		resp, err := conn.GetIntegration(&apigatewayv2.GetIntegrationInput{
			ApiId:         apiId,
			IntegrationId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*vApiId = *apiId
		*v = *resp

		return nil
	}
}

func testAccAWSAPIGatewayV2IntegrationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccAWSAPIGatewayV2IntegrationConfig_apiWebSocket(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
`, rName)
}

func testAccAWSAPIGatewayV2IntegrationConfig_basic(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_apiWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_integration" "test" {
  api_id           = "${aws_apigatewayv2_api.test.id}"
  integration_type = "MOCK"
}
`)
}

func testAccAWSAPIGatewayV2IntegrationConfig_integrationTypeHttp(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_apiWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_integration" "test" {
  api_id           = "${aws_apigatewayv2_api.test.id}"
  integration_type = "HTTP"

  connection_type               = "INTERNET"
  content_handling_strategy     = "CONVERT_TO_TEXT"
  description                   = "Test HTTP"
  integration_method            = "GET"
  integration_uri               = "http://www.example.com"
  passthrough_behavior          = "WHEN_NO_MATCH"
  template_selection_expression = "$request.body.name"
  timeout_milliseconds          = 28999

  request_templates = {
    "application/json" = ""
  }
}
`)
}

func testAccAWSAPIGatewayV2IntegrationConfig_integrationTypeHttpUpdated(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_apiWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_integration" "test" {
  api_id           = "${aws_apigatewayv2_api.test.id}"
  integration_type = "HTTP"

  connection_type               = "INTERNET"
  content_handling_strategy     = "CONVERT_TO_BINARY"
  description                   = "Test HTTP updated"
  integration_method            = "POST"
  integration_uri               = "http://www.example.org"
  passthrough_behavior          = "WHEN_NO_TEMPLATES"
  template_selection_expression = "$request.body.id"
  timeout_milliseconds          = 51

  request_templates = {
    "application/json" = "#set($number=42)"
    "application/xml"  = "#set($percent=$number/100)"
  }
}
`)
}

func testAccAWSAPIGatewayV2IntegrationConfig_lambda(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_apiWebSocket(rName) + baseAccAWSLambdaConfig(rName, rName, rName) + fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  handler       = "index.handler"
  runtime       = "nodejs10.x"
}

resource "aws_apigatewayv2_integration" "test" {
  api_id           = "${aws_apigatewayv2_api.test.id}"
  integration_type = "AWS"

  connection_type               = "INTERNET"
  content_handling_strategy     = "CONVERT_TO_TEXT"
  credentials_arn               = "${data.aws_caller_identity.current.arn}"
  description                   = "Test Lambda"
  integration_method            = "POST"
  integration_uri               = "${aws_lambda_function.test.invoke_arn}"
  passthrough_behavior          = "WHEN_NO_MATCH"
}
`, rName)
}

func testAccAWSAPIGatewayV2IntegrationConfig_vpcLink(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_apiWebSocket(rName) + fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.10.0.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "network"
  subnets            = ["${aws_subnet.test.id}"]
}

resource "aws_api_gateway_vpc_link" "test" {
  name        = %[1]q
  target_arns = ["${aws_lb.test.arn}"]
}

resource "aws_apigatewayv2_integration" "test" {
  api_id           = "${aws_apigatewayv2_api.test.id}"
  integration_type = "HTTP_PROXY"

  connection_id                 = "${aws_api_gateway_vpc_link.test.id}"
  connection_type               = "VPC_LINK"
  content_handling_strategy     = "CONVERT_TO_TEXT"
  description                   = "Test VPC Link"
  integration_method            = "PUT"
  integration_uri               = "http://www.example.net"
  passthrough_behavior          = "NEVER"
  timeout_milliseconds          = 12345
}
`, rName)
}
