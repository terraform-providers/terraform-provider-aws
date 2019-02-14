package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGatewayGatewayResponse_basic(t *testing.T) {
	var conf apigateway.UpdateGatewayResponseOutput

	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayGatewayResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayGatewayResponseConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayGatewayResponseExists("aws_api_gateway_gateway_response.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_gateway_response.test", "status_code", "401"),
					resource.TestCheckResourceAttr("aws_api_gateway_gateway_response.test", "response_parameters.gatewayresponse.header.Authorization", "'Basic'"),
					resource.TestCheckResourceAttr("aws_api_gateway_gateway_response.test", "response_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
					resource.TestCheckNoResourceAttr("aws_api_gateway_gateway_response.test", "response_templates.application/json"),
				),
			},

			{
				Config: testAccAWSAPIGatewayGatewayResponseConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayGatewayResponseExists("aws_api_gateway_gateway_response.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_gateway_response.test", "status_code", "477"),
					resource.TestCheckResourceAttr("aws_api_gateway_gateway_response.test", "response_templates.application/json", "{'message':$context.error.messageString}"),
					resource.TestCheckNoResourceAttr("aws_api_gateway_gateway_response.test", "response_templates.application/xml"),
					resource.TestCheckNoResourceAttr("aws_api_gateway_gateway_response.test", "response_parameters.gatewayresponse.header.Authorization"),
				),
			},
			{
				ResourceName:      "aws_api_gateway_gateway_response.test",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayGatewayResponseImportStateIdFunc("aws_api_gateway_gateway_response.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayGatewayResponseExists(n string, res *apigateway.UpdateGatewayResponseOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Gateway Response ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigateway

		req := &apigateway.GetGatewayResponseInput{
			RestApiId:    aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.main"].Primary.ID),
			ResponseType: aws.String(rs.Primary.Attributes["response_type"]),
		}
		describe, err := conn.GetGatewayResponse(req)
		if err != nil {
			return err
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayGatewayResponseDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_gateway_response" {
			continue
		}

		req := &apigateway.GetGatewayResponseInput{
			RestApiId:    aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.main"].Primary.ID),
			ResponseType: aws.String(rs.Primary.Attributes["response_type"]),
		}
		_, err := conn.GetGatewayResponse(req)

		if err == nil {
			return fmt.Errorf("API Gateway Gateway Response still exists")
		}

		aws2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if aws2err.Code() != "NotFoundException" {
			return err
		}

		return nil
	}

	return nil
}

func testAccAWSAPIGatewayGatewayResponseImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["response_type"]), nil
	}
}

func testAccAWSAPIGatewayGatewayResponseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "main" {
  name = "%s"
}

resource "aws_api_gateway_gateway_response" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.main.id}"
  status_code = "401"
  response_type = "UNAUTHORIZED"

  response_templates = {
    "application/xml"  = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  response_parameters = {
    "gatewayresponse.header.Authorization" = "'Basic'"
  }
}
`, rName)
}

func testAccAWSAPIGatewayGatewayResponseConfigUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "main" {
  name = "%s"
}

resource "aws_api_gateway_gateway_response" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.main.id}"
  status_code = "477"
  response_type = "UNAUTHORIZED"

  response_templates = {
    "application/json" = "{'message':$context.error.messageString}"
  }
}
`, rName)
}
