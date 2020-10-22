package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAPIGatewayRequestValidator_basic(t *testing.T) {
	var conf apigateway.UpdateRequestValidatorOutput
	rName := fmt.Sprintf("tf-test-acc-%s", acctest.RandString(8))
	resourceName := "aws_api_gateway_request_validator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRequestValidatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRequestValidatorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRequestValidatorExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayRequestValidatorName(&conf, "tf-acc-test-request-validator"),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-acc-test-request-validator"),
					testAccCheckAWSAPIGatewayRequestValidatorValidateRequestBody(&conf, false),
					resource.TestCheckResourceAttr(resourceName, "validate_request_body", "false"),
					testAccCheckAWSAPIGatewayRequestValidatorValidateRequestParameters(&conf, false),
					resource.TestCheckResourceAttr(resourceName, "validate_request_parameters", "false"),
				),
			},
			{
				Config: testAccAWSAPIGatewayRequestValidatorUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRequestValidatorExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayRequestValidatorName(&conf, "tf-acc-test-request-validator_modified"),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-acc-test-request-validator_modified"),
					testAccCheckAWSAPIGatewayRequestValidatorValidateRequestBody(&conf, true),
					resource.TestCheckResourceAttr(resourceName, "validate_request_body", "true"),
					testAccCheckAWSAPIGatewayRequestValidatorValidateRequestParameters(&conf, true),
					resource.TestCheckResourceAttr(resourceName, "validate_request_parameters", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayRequestValidatorImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayRequestValidator_disappears(t *testing.T) {
	var conf apigateway.UpdateRequestValidatorOutput
	rName := fmt.Sprintf("tf-test-acc-%s", acctest.RandString(8))
	resourceName := "aws_api_gateway_request_validator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRequestValidatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRequestValidatorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRequestValidatorExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayRequestValidator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayRequestValidatorName(conf *apigateway.UpdateRequestValidatorOutput, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.Name == nil {
			return fmt.Errorf("Empty Name, expected: %q", expectedName)
		}
		if *conf.Name != expectedName {
			return fmt.Errorf("Name didn't match. Expected: %q, Given: %q", expectedName, *conf.Name)
		}
		return nil
	}
}

func testAccCheckAWSAPIGatewayRequestValidatorValidateRequestBody(conf *apigateway.UpdateRequestValidatorOutput, expectedValue bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.ValidateRequestBody == nil {
			return fmt.Errorf("Empty ValidateRequestBody, expected: %t", expectedValue)
		}
		if *conf.ValidateRequestBody != expectedValue {
			return fmt.Errorf("ValidateRequestBody didn't match. Expected: %t, Given: %t", expectedValue, *conf.ValidateRequestBody)
		}
		return nil
	}
}

func testAccCheckAWSAPIGatewayRequestValidatorValidateRequestParameters(conf *apigateway.UpdateRequestValidatorOutput, expectedValue bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.ValidateRequestParameters == nil {
			return fmt.Errorf("Empty ValidateRequestParameters, expected: %t", expectedValue)
		}
		if *conf.ValidateRequestParameters != expectedValue {
			return fmt.Errorf("ValidateRequestParameters didn't match. Expected: %t, Given: %t", expectedValue, *conf.ValidateRequestParameters)
		}
		return nil
	}
}

func testAccCheckAWSAPIGatewayRequestValidatorExists(n string, res *apigateway.UpdateRequestValidatorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Request Validator ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		req := &apigateway.GetRequestValidatorInput{
			RequestValidatorId: aws.String(rs.Primary.ID),
			RestApiId:          aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		describe, err := conn.GetRequestValidator(req)
		if err != nil {
			return err
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayRequestValidatorDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_request_validator" {
			continue
		}

		req := &apigateway.GetRequestValidatorInput{
			RequestValidatorId: aws.String(rs.Primary.ID),
			RestApiId:          aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		_, err := conn.GetRequestValidator(req)

		if err == nil {
			return fmt.Errorf("API Request Validator still exists")
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

func testAccAWSAPIGatewayRequestValidatorImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.ID), nil
	}
}

func testAccAWSAPIGatewayRequestValidatorConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, rName)
}

func testAccAWSAPIGatewayRequestValidatorConfig(rName string) string {
	return fmt.Sprintf(testAccAWSAPIGatewayRequestValidatorConfig_base(rName) + `
resource "aws_api_gateway_request_validator" "test" {
  name        = "tf-acc-test-request-validator"
  rest_api_id = aws_api_gateway_rest_api.test.id
}
`)
}

func testAccAWSAPIGatewayRequestValidatorUpdatedConfig(rName string) string {
	return fmt.Sprintf(testAccAWSAPIGatewayRequestValidatorConfig_base(rName) + `
resource "aws_api_gateway_request_validator" "test" {
  name                        = "tf-acc-test-request-validator_modified"
  rest_api_id                 = aws_api_gateway_rest_api.test.id
  validate_request_body       = true
  validate_request_parameters = true
}
`)
}
