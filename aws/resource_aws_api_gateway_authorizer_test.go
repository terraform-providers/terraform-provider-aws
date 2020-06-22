package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSAPIGatewayAuthorizer_basic(t *testing.T) {
	var conf apigateway.Authorizer
	apiGatewayName := acctest.RandomWithPrefix("tf-acctest-apigw")
	authorizerName := acctest.RandomWithPrefix("tf-acctest-igw-authorizer")
	lambdaName := acctest.RandomWithPrefix("tf-acctest-igw-auth-lambda")

	resourceName := "aws_api_gateway_authorizer.acctest"
	lambdaResourceName := "aws_lambda_function.authorizer"
	roleResourceName := "aws_iam_role.invocation_role"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayAuthorizerConfig_lambda(apiGatewayName, authorizerName, lambdaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayAuthorizerExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_source", "method.request.header.Authorization"),
					resource.TestCheckResourceAttr(resourceName, "name", authorizerName),
					resource.TestCheckResourceAttr(resourceName, "type", "TOKEN"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", strconv.Itoa(defaultAuthorizerTTL)),
					resource.TestCheckResourceAttr(resourceName, "identity_validation_expression", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayAuthorizerConfig_lambdaUpdate(apiGatewayName, authorizerName, lambdaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayAuthorizerExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_source", "method.request.header.Authorization"),
					resource.TestCheckResourceAttr(resourceName, "name", authorizerName+"_modified"),
					resource.TestCheckResourceAttr(resourceName, "type", "TOKEN"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", strconv.Itoa(360)),
					resource.TestCheckResourceAttr(resourceName, "identity_validation_expression", ".*"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayAuthorizer_cognito(t *testing.T) {
	apiGatewayName := acctest.RandomWithPrefix("tf-acctest-apigw")
	authorizerName := acctest.RandomWithPrefix("tf-acctest-igw-authorizer")
	cognitoName := acctest.RandomWithPrefix("tf-acctest-cognito-user-pool")

	resourceName := "aws_api_gateway_authorizer.acctest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayAuthorizerConfig_cognito(apiGatewayName, authorizerName, cognitoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", authorizerName+"-cognito"),
					resource.TestCheckResourceAttr(resourceName, "type", "COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "provider_arns.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayAuthorizerConfig_cognitoUpdate(apiGatewayName, authorizerName, cognitoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", authorizerName+"-cognito-update"),
					resource.TestCheckResourceAttr(resourceName, "type", "COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "provider_arns.#", "3"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayAuthorizer_switchAuthType(t *testing.T) {
	apiGatewayName := acctest.RandomWithPrefix("tf-acctest-apigw")
	authorizerName := acctest.RandomWithPrefix("tf-acctest-igw-authorizer")
	lambdaName := acctest.RandomWithPrefix("tf-acctest-igw-auth-lambda")
	cognitoName := acctest.RandomWithPrefix("tf-acctest-cognito-user-pool")

	resourceName := "aws_api_gateway_authorizer.acctest"
	lambdaResourceName := "aws_lambda_function.authorizer"
	roleResourceName := "aws_iam_role.invocation_role"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayAuthorizerConfig_lambda(apiGatewayName, authorizerName, lambdaName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", authorizerName),
					resource.TestCheckResourceAttr(resourceName, "type", "TOKEN"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials", roleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayAuthorizerConfig_cognito(apiGatewayName, authorizerName, cognitoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", authorizerName+"-cognito"),
					resource.TestCheckResourceAttr(resourceName, "type", "COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "provider_arns.#", "2"),
				),
			},
			{
				Config: testAccAWSAPIGatewayAuthorizerConfig_lambdaUpdate(apiGatewayName, authorizerName, lambdaName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", authorizerName+"_modified"),
					resource.TestCheckResourceAttr(resourceName, "type", "TOKEN"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials", roleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayAuthorizer_switchAuthorizerTTL(t *testing.T) {
	var conf apigateway.Authorizer
	apiGatewayName := acctest.RandomWithPrefix("tf-acctest-apigw")
	authorizerName := acctest.RandomWithPrefix("tf-acctest-igw-authorizer")
	lambdaName := acctest.RandomWithPrefix("tf-acctest-igw-auth-lambda")
	resourceName := "aws_api_gateway_authorizer.acctest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayAuthorizerConfig_lambda(apiGatewayName, authorizerName, lambdaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayAuthorizerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", strconv.Itoa(defaultAuthorizerTTL)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayAuthorizerConfig_lambdaUpdate(apiGatewayName, authorizerName, lambdaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayAuthorizerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", strconv.Itoa(360)),
				),
			},
			{
				Config: testAccAWSAPIGatewayAuthorizerConfig_lambdaNoCache(apiGatewayName, authorizerName, lambdaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayAuthorizerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", strconv.Itoa(0)),
				),
			},
			{
				Config: testAccAWSAPIGatewayAuthorizerConfig_lambda(apiGatewayName, authorizerName, lambdaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayAuthorizerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", strconv.Itoa(defaultAuthorizerTTL)),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayAuthorizer_authTypeValidation(t *testing.T) {
	apiGatewayName := acctest.RandomWithPrefix("tf-acctest-apigw")
	authorizerName := acctest.RandomWithPrefix("tf-acctest-igw-authorizer")
	lambdaName := acctest.RandomWithPrefix("tf-acctest-igw-auth-lambda")
	cognitoName := acctest.RandomWithPrefix("tf-acctest-cognito-user-pool")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSAPIGatewayAuthorizerConfig_authTypeValidationDefaultToken(apiGatewayName, authorizerName, lambdaName),
				ExpectError: regexp.MustCompile(`authorizer_uri must be set non-empty when authorizer type is TOKEN`),
			},
			{
				Config:      testAccAWSAPIGatewayAuthorizerConfig_authTypeValidationRequest(apiGatewayName, authorizerName, lambdaName),
				ExpectError: regexp.MustCompile(`authorizer_uri must be set non-empty when authorizer type is REQUEST`),
			},
			{
				Config:      testAccAWSAPIGatewayAuthorizerConfig_authTypeValidationCognito(apiGatewayName, authorizerName, cognitoName),
				ExpectError: regexp.MustCompile(`provider_arns must be set non-empty when authorizer type is COGNITO_USER_POOLS`),
			},
		},
	})
}

func TestAccAWSAPIGatewayAuthorizer_disappears(t *testing.T) {
	var conf apigateway.Authorizer
	apiGatewayName := acctest.RandomWithPrefix("tf-acctest-apigw")
	authorizerName := acctest.RandomWithPrefix("tf-acctest-igw-authorizer")
	lambdaName := acctest.RandomWithPrefix("tf-acctest-igw-auth-lambda")
	resourceName := "aws_api_gateway_authorizer.acctest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayAuthorizerConfig_lambda(apiGatewayName, authorizerName, lambdaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayAuthorizerExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayAuthorizerDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayAuthorizerDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}
		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn
		authorizerId := rs.Primary.ID
		restApiId := rs.Primary.Attributes["rest_api_id"]

		input := &apigateway.DeleteAuthorizerInput{
			AuthorizerId: aws.String(authorizerId),
			RestApiId:    aws.String(restApiId),
		}
		_, err := conn.DeleteAuthorizer(input)

		return err
	}
}

func testAccCheckAWSAPIGatewayAuthorizerExists(n string, res *apigateway.Authorizer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Authorizer ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		req := &apigateway.GetAuthorizerInput{
			AuthorizerId: aws.String(rs.Primary.ID),
			RestApiId:    aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		describe, err := conn.GetAuthorizer(req)
		if err != nil {
			return err
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayAuthorizerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_authorizer" {
			continue
		}

		req := &apigateway.GetAuthorizerInput{
			AuthorizerId: aws.String(rs.Primary.ID),
			RestApiId:    aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		_, err := conn.GetAuthorizer(req)

		if err == nil {
			return fmt.Errorf("API Gateway Authorizer still exists")
		}

		aws2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if aws2err.Code() != apigateway.ErrCodeNotFoundException {
			return err
		}

		return nil
	}

	return nil
}

func testAccAWSAPIGatewayAuthorizerImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.ID), nil
	}
}

func testAccAWSAPIGatewayAuthorizerConfig_baseLambda(apiGatewayName, lambdaName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "acctest" {
  name = "%s"
}

resource "aws_iam_role" "invocation_role" {
  name = "%s_auth_invocation_role"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "invocation_policy" {
  name = "default"
  role = aws_iam_role.invocation_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "lambda:InvokeFunction",
      "Effect": "Allow",
      "Resource": "${aws_lambda_function.authorizer.arn}"
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "%s_authorizer_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "authorizer" {
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = "%s"
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs12.x"
}
`, apiGatewayName, apiGatewayName, apiGatewayName, lambdaName)
}

func testAccAWSAPIGatewayAuthorizerConfig_lambda(apiGatewayName, authorizerName, lambdaName string) string {
	return testAccAWSAPIGatewayAuthorizerConfig_baseLambda(apiGatewayName, lambdaName) + fmt.Sprintf(`
resource "aws_api_gateway_authorizer" "acctest" {
  name                   = "%s"
  rest_api_id            = aws_api_gateway_rest_api.acctest.id
  authorizer_uri         = aws_lambda_function.authorizer.invoke_arn
  authorizer_credentials = aws_iam_role.invocation_role.arn
}
`, authorizerName)
}

func testAccAWSAPIGatewayAuthorizerConfig_lambdaUpdate(apiGatewayName, authorizerName, lambdaName string) string {
	return testAccAWSAPIGatewayAuthorizerConfig_baseLambda(apiGatewayName, lambdaName) + fmt.Sprintf(`
resource "aws_api_gateway_authorizer" "acctest" {
  name                             = "%s_modified"
  rest_api_id                      = aws_api_gateway_rest_api.acctest.id
  authorizer_uri                   = aws_lambda_function.authorizer.invoke_arn
  authorizer_credentials           = aws_iam_role.invocation_role.arn
  authorizer_result_ttl_in_seconds = 360
  identity_validation_expression   = ".*"
}
`, authorizerName)
}

func testAccAWSAPIGatewayAuthorizerConfig_lambdaNoCache(apiGatewayName, authorizerName, lambdaName string) string {
	return testAccAWSAPIGatewayAuthorizerConfig_baseLambda(apiGatewayName, lambdaName) + fmt.Sprintf(`
resource "aws_api_gateway_authorizer" "acctest" {
  name                             = "%s_modified"
  rest_api_id                      = aws_api_gateway_rest_api.acctest.id
  authorizer_uri                   = aws_lambda_function.authorizer.invoke_arn
  authorizer_credentials           = aws_iam_role.invocation_role.arn
  authorizer_result_ttl_in_seconds = 0
  identity_validation_expression   = ".*"
}
`, authorizerName)
}

func testAccAWSAPIGatewayAuthorizerConfig_cognito(apiGatewayName, authorizerName, cognitoName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "acctest" {
  name = "%s"
}

resource "aws_cognito_user_pool" "acctest" {
  count = 2
  name  = "%s-${count.index}"
}

resource "aws_api_gateway_authorizer" "acctest" {
  name          = "%s-cognito"
  type          = "COGNITO_USER_POOLS"
  rest_api_id   = aws_api_gateway_rest_api.acctest.id
  provider_arns = aws_cognito_user_pool.acctest[*].arn
}
`, apiGatewayName, cognitoName, authorizerName)
}

func testAccAWSAPIGatewayAuthorizerConfig_cognitoUpdate(apiGatewayName, authorizerName, cognitoName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "acctest" {
  name = "%s"
}

resource "aws_cognito_user_pool" "acctest_update" {
  count = 3
  name  = "%s-${count.index}-update"
}

resource "aws_api_gateway_authorizer" "acctest" {
  name          = "%s-cognito-update"
  type          = "COGNITO_USER_POOLS"
  rest_api_id   = aws_api_gateway_rest_api.acctest.id
  provider_arns = aws_cognito_user_pool.acctest_update[*].arn
}
`, apiGatewayName, cognitoName, authorizerName)
}

func testAccAWSAPIGatewayAuthorizerConfig_authTypeValidationDefaultToken(apiGatewayName, authorizerName, lambdaName string) string {
	return testAccAWSAPIGatewayAuthorizerConfig_baseLambda(apiGatewayName, lambdaName) + fmt.Sprintf(`
resource "aws_api_gateway_authorizer" "acctest" {
  name                   = "%s"
  rest_api_id            = aws_api_gateway_rest_api.acctest.id
  authorizer_credentials = aws_iam_role.invocation_role.arn
}
`, authorizerName)
}

func testAccAWSAPIGatewayAuthorizerConfig_authTypeValidationRequest(apiGatewayName, authorizerName, lambdaName string) string {
	return testAccAWSAPIGatewayAuthorizerConfig_baseLambda(apiGatewayName, lambdaName) + fmt.Sprintf(`
resource "aws_api_gateway_authorizer" "acctest" {
  name                   = "%s"
  type                   = "REQUEST"
  rest_api_id            = aws_api_gateway_rest_api.acctest.id
  authorizer_credentials = aws_iam_role.invocation_role.arn
}
`, authorizerName)
}

func testAccAWSAPIGatewayAuthorizerConfig_authTypeValidationCognito(apiGatewayName, authorizerName, cognitoName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "acctest" {
  name = "%s"
}

resource "aws_cognito_user_pool" "acctest" {
  count = 2
  name  = "%s-${count.index}"
}

resource "aws_api_gateway_authorizer" "acctest" {
  name        = "%s-cognito"
  type        = "COGNITO_USER_POOLS"
  rest_api_id = aws_api_gateway_rest_api.acctest.id
}
`, apiGatewayName, cognitoName, authorizerName)
}
