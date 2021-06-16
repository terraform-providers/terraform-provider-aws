package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSAPIGatewayV2ApiDataSource_Http(t *testing.T) {
	dataSourceName := "data.aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, apigatewayv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiDataSourceConfigHttp(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "api_endpoint", resourceName, "api_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "api_key_selection_expression", resourceName, "api_key_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.#", resourceName, "cors_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.0.allow_credentials", resourceName, "cors_configuration.0.allow_credentials"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.0.allow_headers.#", resourceName, "cors_configuration.0.allow_headers.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.0.allow_methods.#", resourceName, "cors_configuration.0.allow_methods.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.0.allow_origins.#", resourceName, "cors_configuration.0.allow_origins.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.0.expose_headers.#", resourceName, "cors_configuration.0.expose_headers.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.0.max_age", resourceName, "cors_configuration.0.max_age"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "disable_execute_api_endpoint", resourceName, "disable_execute_api_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_arn", resourceName, "execution_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "protocol_type", resourceName, "protocol_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "route_selection_expression", resourceName, "route_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Key1", resourceName, "tags.Key1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Key2", resourceName, "tags.Key2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2ApiDataSource_WebSocket(t *testing.T) {
	dataSourceName := "data.aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, apigatewayv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiDataSourceConfigWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "api_endpoint", resourceName, "api_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "api_key_selection_expression", resourceName, "api_key_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.#", resourceName, "cors_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "disable_execute_api_endpoint", resourceName, "disable_execute_api_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_arn", resourceName, "execution_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "protocol_type", resourceName, "protocol_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "route_selection_expression", resourceName, "route_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Key1", resourceName, "tags.Key1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Key2", resourceName, "tags.Key2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func testAccAWSAPIGatewayV2ApiDataSourceConfigHttp(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  description   = "test description"
  name          = %[1]q
  protocol_type = "HTTP"
  version       = "v1"

  cors_configuration {
    allow_headers = ["Authorization"]
    allow_methods = ["GET", "put"]
    allow_origins = ["https://www.example.com"]
  }

  tags = {
    Key1 = "Value1h"
    Key2 = "Value2h"
  }
}

data "aws_apigatewayv2_api" "test" {
  api_id = aws_apigatewayv2_api.test.id
}
`, rName)
}

func testAccAWSAPIGatewayV2ApiDataSourceConfigWebSocket(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  api_key_selection_expression = "$context.authorizer.usageIdentifierKey"
  description                  = "test description"
  name                         = %[1]q
  protocol_type                = "WEBSOCKET"
  route_selection_expression   = "$request.body.service"
  version                      = "v1"

  tags = {
    Key1 = "Value1ws"
    Key2 = "Value2ws"
  }
}

data "aws_apigatewayv2_api" "test" {
  api_id = aws_apigatewayv2_api.test.id
}
`, rName)
}
