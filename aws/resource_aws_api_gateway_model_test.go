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

func TestAccAWSAPIGatewayModel_basic(t *testing.T) {
	var conf apigateway.Model
	rInt := acctest.RandString(10)
	rName := fmt.Sprintf("tf-acc-test-%s", rInt)
	modelName := fmt.Sprintf("tfacctest%s", rInt)
	resourceName := "aws_api_gateway_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayModelConfig(rName, modelName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayModelExists(resourceName, modelName, &conf),
					testAccCheckAWSAPIGatewayModelAttributes(&conf, modelName),
					resource.TestCheckResourceAttr(
						resourceName, "name", modelName),
					resource.TestCheckResourceAttr(
						resourceName, "description", "a test schema"),
					resource.TestCheckResourceAttr(
						resourceName, "content_type", "application/json"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayModelImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayModel_disappears(t *testing.T) {
	var conf apigateway.Model
	rInt := acctest.RandString(10)
	rName := fmt.Sprintf("tf-acc-test-%s", rInt)
	modelName := fmt.Sprintf("tfacctest%s", rInt)
	resourceName := "aws_api_gateway_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayModelConfig(rName, modelName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayModelExists(resourceName, modelName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayModel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayModelAttributes(conf *apigateway.Model, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *conf.Name != name {
			return fmt.Errorf("Wrong Name: %q", *conf.Name)
		}
		if *conf.Description != "a test schema" {
			return fmt.Errorf("Wrong Description: %q", *conf.Description)
		}
		if *conf.ContentType != "application/json" {
			return fmt.Errorf("Wrong ContentType: %q", *conf.ContentType)
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayModelExists(n, rName string, res *apigateway.Model) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Model ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		req := &apigateway.GetModelInput{
			ModelName: aws.String(rName),
			RestApiId: aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
		}
		describe, err := conn.GetModel(req)
		if err != nil {
			return err
		}
		if *describe.Id != rs.Primary.ID {
			return fmt.Errorf("APIGateway Model not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayModelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_model" {
			continue
		}

		req := &apigateway.GetModelsInput{
			RestApiId: aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
		}
		describe, err := conn.GetModels(req)

		if err == nil {
			if len(describe.Items) != 0 &&
				*describe.Items[0].Id == rs.Primary.ID {
				return fmt.Errorf("API Gateway Model still exists")
			}
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

func testAccAWSAPIGatewayModelImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["name"]), nil
	}
}

func testAccAWSAPIGatewayModelConfig(rName, modelName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}

resource "aws_api_gateway_model" "test" {
  rest_api_id  = aws_api_gateway_rest_api.test.id
  name         = "%s"
  description  = "a test schema"
  content_type = "application/json"
  schema       = <<EOF
{
  "type": "object"
}
EOF
}
`, rName, modelName)
}
