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

func TestAccAWSAPIGatewayV2Authorizer_basic(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetAuthorizerOutput
	resourceName := "aws_apigatewayv2_authorizer.test"
	lambdaResourceName := "aws_lambda_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckAWSAPIGatewayV2AuthorizerDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2AuthorizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2AuthorizerExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.645907014", "route.request.header.Auth"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2AuthorizerImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Authorizer_disappears(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetAuthorizerOutput
	resourceName := "aws_apigatewayv2_authorizer.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2AuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2AuthorizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2AuthorizerExists(resourceName, &apiId, &v),
					testAccCheckAWSAPIGatewayV2AuthorizerDisappears(&apiId, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Authorizer_Credentials(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetAuthorizerOutput
	resourceName := "aws_apigatewayv2_authorizer.test"
	iamRoleResourceName := "aws_iam_role.test"
	lambdaResourceName := "aws_lambda_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckAWSAPIGatewayV2AuthorizerDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2AuthorizerConfig_credentials(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2AuthorizerExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.645907014", "route.request.header.Auth"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2AuthorizerImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2AuthorizerConfig_credentialsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2AuthorizerExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.645907014", "route.request.header.Auth"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.4138478046", "route.request.querystring.Name"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("%s-updated", rName)),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2AuthorizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2AuthorizerExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.645907014", "route.request.header.Auth"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Authorizer_JWT(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetAuthorizerOutput
	resourceName := "aws_apigatewayv2_authorizer.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckAWSAPIGatewayV2AuthorizerDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2AuthorizerConfig_jwt(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2AuthorizerExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "JWT"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.2786136151", "$request.header.Authorization"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.0.audience.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.0.audience.1785148924", "test"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2AuthorizerImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2AuthorizerConfig_jwtUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2AuthorizerExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "JWT"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.2786136151", "$request.header.Authorization"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.0.audience.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.0.audience.1785148924", "test"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.0.audience.2323796166", "testing"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayV2AuthorizerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_authorizer" {
			continue
		}

		_, err := conn.GetAuthorizer(&apigatewayv2.GetAuthorizerInput{
			ApiId:        aws.String(rs.Primary.Attributes["api_id"]),
			AuthorizerId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 authorizer %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGatewayV2AuthorizerDisappears(apiId *string, v *apigatewayv2.GetAuthorizerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.DeleteAuthorizer(&apigatewayv2.DeleteAuthorizerInput{
			ApiId:        apiId,
			AuthorizerId: v.AuthorizerId,
		})

		return err
	}
}

func testAccCheckAWSAPIGatewayV2AuthorizerExists(n string, vApiId *string, v *apigatewayv2.GetAuthorizerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 authorizer ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		apiId := aws.String(rs.Primary.Attributes["api_id"])
		resp, err := conn.GetAuthorizer(&apigatewayv2.GetAuthorizerInput{
			ApiId:        apiId,
			AuthorizerId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*vApiId = *apiId
		*v = *resp

		return nil
	}
}

func testAccAWSAPIGatewayV2AuthorizerImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccAWSAPIGatewayV2AuthorizerConfig_baseWebSocket(rName string) string {
	return baseAccAWSLambdaConfig(rName, rName, rName) + fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  handler       = "index.handler"
  runtime       = "nodejs10.x"
}

resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}

resource "aws_iam_role" "test" {
  name = "%[1]s_auth_invocation_role"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {"Service": "apigateway.amazonaws.com"},
    "Effect": "Allow"
  }]
}
EOF
}
`, rName)
}

func testAccAWSAPIGatewayV2AuthorizerConfig_baseHttp(rName string) string {
	return baseAccAWSLambdaConfig(rName, rName, rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAWSAPIGatewayV2AuthorizerConfig_basic(rName string) string {
	return testAccAWSAPIGatewayV2AuthorizerConfig_baseWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_authorizer" "test" {
  api_id           = "${aws_apigatewayv2_api.test.id}"
  authorizer_type  = "REQUEST"
  authorizer_uri   = "${aws_lambda_function.test.invoke_arn}"
  identity_sources = ["route.request.header.Auth"]
  name             = %[1]q
}
`, rName)
}

func testAccAWSAPIGatewayV2AuthorizerConfig_credentials(rName string) string {
	return testAccAWSAPIGatewayV2AuthorizerConfig_baseWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_authorizer" "test" {
  api_id           = "${aws_apigatewayv2_api.test.id}"
  authorizer_type  = "REQUEST"
  authorizer_uri   = "${aws_lambda_function.test.invoke_arn}"
  identity_sources = ["route.request.header.Auth"]
  name             = %[1]q

  authorizer_credentials_arn = "${aws_iam_role.test.arn}"
}
`, rName)
}

func testAccAWSAPIGatewayV2AuthorizerConfig_credentialsUpdated(rName string) string {
	return testAccAWSAPIGatewayV2AuthorizerConfig_baseWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_authorizer" "test" {
  api_id           = "${aws_apigatewayv2_api.test.id}"
  authorizer_type  = "REQUEST"
  authorizer_uri   = "${aws_lambda_function.test.invoke_arn}"
  identity_sources = ["route.request.header.Auth", "route.request.querystring.Name"]
  name             = "%[1]s-updated"

  authorizer_credentials_arn = "${aws_iam_role.test.arn}"
}
`, rName)
}

func testAccAWSAPIGatewayV2AuthorizerConfig_jwt(rName string) string {
	return testAccAWSAPIGatewayV2AuthorizerConfig_baseHttp(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_authorizer" "test" {
  api_id           = "${aws_apigatewayv2_api.test.id}"
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]
  name             = %[1]q

  jwt_configuration {
    audience = ["test"]
    issuer   = "https://${aws_cognito_user_pool.test.endpoint}"
  }
}
`, rName)
}

func testAccAWSAPIGatewayV2AuthorizerConfig_jwtUpdated(rName string) string {
	return testAccAWSAPIGatewayV2AuthorizerConfig_baseHttp(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_authorizer" "test" {
  api_id           = "${aws_apigatewayv2_api.test.id}"
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]
  name             = %[1]q

  jwt_configuration {
    audience = ["test", "testing"]
    issuer   = "https://${aws_cognito_user_pool.test.endpoint}"
  }
}
`, rName)
}
