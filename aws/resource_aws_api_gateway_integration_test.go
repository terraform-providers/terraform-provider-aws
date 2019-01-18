package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGatewayIntegration_basic(t *testing.T) {
	var conf apigateway.Integration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayIntegrationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayIntegrationExists("aws_api_gateway_integration.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "type", "HTTP"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "integration_http_method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "uri", "https://www.google.de"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "credentials", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/json", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "timeout_milliseconds", "29000"),
				),
			},

			{
				Config: testAccAWSAPIGatewayIntegrationConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayIntegrationExists("aws_api_gateway_integration.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "type", "HTTP"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "integration_http_method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "uri", "https://www.google.de"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "credentials", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Authorization", "'updated'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-FooBar", "'Baz'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/json", "{'foobar': 'bar}"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.text/html", "<html>Foo</html>"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "timeout_milliseconds", "2000"),
				),
			},

			{
				Config: testAccAWSAPIGatewayIntegrationConfigUpdateURI,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayIntegrationExists("aws_api_gateway_integration.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "type", "HTTP"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "integration_http_method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "uri", "https://www.google.de/updated"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "credentials", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/json", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "timeout_milliseconds", "2000"),
				),
			},

			{
				Config: testAccAWSAPIGatewayIntegrationConfigUpdateNoTemplates,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayIntegrationExists("aws_api_gateway_integration.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "type", "HTTP"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "integration_http_method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "uri", "https://www.google.de"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "credentials", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.%", "0"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.%", "0"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "timeout_milliseconds", "2000"),
				),
			},

			{
				Config: testAccAWSAPIGatewayIntegrationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayIntegrationExists("aws_api_gateway_integration.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "type", "HTTP"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "integration_http_method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "uri", "https://www.google.de"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "credentials", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/json", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "timeout_milliseconds", "29000"),
				),
			},
			{
				ResourceName:      "aws_api_gateway_integration.test",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayIntegrationImportStateIdFunc("aws_api_gateway_integration.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayIntegration_contentHandling(t *testing.T) {
	var conf apigateway.Integration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayIntegrationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayIntegrationExists("aws_api_gateway_integration.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "type", "HTTP"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "integration_http_method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "uri", "https://www.google.de"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "credentials", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/json", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
				),
			},

			{
				Config: testAccAWSAPIGatewayIntegrationConfigUpdateContentHandling,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayIntegrationExists("aws_api_gateway_integration.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "type", "HTTP"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "integration_http_method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "uri", "https://www.google.de"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "content_handling", "CONVERT_TO_BINARY"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "credentials", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/json", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
				),
			},

			{
				Config: testAccAWSAPIGatewayIntegrationConfigRemoveContentHandling,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayIntegrationExists("aws_api_gateway_integration.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "type", "HTTP"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "integration_http_method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "uri", "https://www.google.de"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "content_handling", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "credentials", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/json", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
				),
			},
			{
				ResourceName:      "aws_api_gateway_integration.test",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayIntegrationImportStateIdFunc("aws_api_gateway_integration.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayIntegration_cache_key_parameters(t *testing.T) {
	var conf apigateway.Integration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayIntegrationConfigCacheKeyParameters,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayIntegrationExists("aws_api_gateway_integration.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "type", "HTTP"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "integration_http_method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "uri", "https://www.google.de"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "credentials", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.%", "3"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_parameters.integration.request.path.param", "method.request.path.param"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "cache_key_parameters.#", "1"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "cache_key_parameters.550492954", "method.request.path.param"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "cache_namespace", "foobar"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.%", "2"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/json", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
				),
			},
			{
				ResourceName:      "aws_api_gateway_integration.test",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayIntegrationImportStateIdFunc("aws_api_gateway_integration.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayIntegration_integrationType(t *testing.T) {
	var conf apigateway.Integration

	rName := fmt.Sprintf("tf-acctest-apigw-int-%s", acctest.RandString(7))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayIntegrationConfig_IntegrationTypeInternet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayIntegrationExists("aws_api_gateway_integration.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "connection_id", ""),
				),
			},
			{
				Config: testAccAWSAPIGatewayIntegrationConfig_IntegrationTypeVpcLink(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayIntegrationExists("aws_api_gateway_integration.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "connection_type", "VPC_LINK"),
					resource.TestMatchResourceAttr("aws_api_gateway_integration.test", "connection_id", regexp.MustCompile("^[0-9a-z]+$")),
				),
			},
			{
				Config: testAccAWSAPIGatewayIntegrationConfig_IntegrationTypeInternet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayIntegrationExists("aws_api_gateway_integration.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr("aws_api_gateway_integration.test", "connection_id", ""),
				),
			},
			{
				ResourceName:      "aws_api_gateway_integration.test",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayIntegrationImportStateIdFunc("aws_api_gateway_integration.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayIntegrationExists(n string, res *apigateway.Integration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Method ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigateway

		req := &apigateway.GetIntegrationInput{
			HttpMethod: aws.String("GET"),
			ResourceId: aws.String(s.RootModule().Resources["aws_api_gateway_resource.test"].Primary.ID),
			RestApiId:  aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
		}
		describe, err := conn.GetIntegration(req)
		if err != nil {
			return err
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayIntegrationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_integration" {
			continue
		}

		req := &apigateway.GetIntegrationInput{
			HttpMethod: aws.String("GET"),
			ResourceId: aws.String(s.RootModule().Resources["aws_api_gateway_resource.test"].Primary.ID),
			RestApiId:  aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
		}
		_, err := conn.GetIntegration(req)

		if err == nil {
			return fmt.Errorf("API Gateway Method still exists")
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

func testAccAWSAPIGatewayIntegrationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["resource_id"], rs.Primary.Attributes["http_method"]), nil
	}
}

const testAccAWSAPIGatewayIntegrationConfig = `
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  parent_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
  path_part = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"

  request_templates = {
    "application/json" = ""
    "application/xml" = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'static'"
    "integration.request.header.X-Foo" = "'Bar'"
  }

  type = "HTTP"
  uri = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior = "WHEN_NO_MATCH"
	content_handling = "CONVERT_TO_TEXT"
}
`

const testAccAWSAPIGatewayIntegrationConfigUpdate = `
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  parent_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
  path_part = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"

  request_templates = {
    "application/json" = "{'foobar': 'bar}"
    "text/html" = "<html>Foo</html>"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'updated'"
    "integration.request.header.X-FooBar" = "'Baz'"
  }

  type = "HTTP"
  uri = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior = "WHEN_NO_MATCH"
	content_handling = "CONVERT_TO_TEXT"
	timeout_milliseconds = 2000
}
`

const testAccAWSAPIGatewayIntegrationConfigUpdateURI = `
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  parent_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
  path_part = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"

  request_templates = {
    "application/json" = ""
    "application/xml" = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'static'"
    "integration.request.header.X-Foo" = "'Bar'"
  }

  type = "HTTP"
  uri = "https://www.google.de/updated"
  integration_http_method = "GET"
  passthrough_behavior = "WHEN_NO_MATCH"
	content_handling = "CONVERT_TO_TEXT"
	timeout_milliseconds = 2000
}
`

const testAccAWSAPIGatewayIntegrationConfigUpdateContentHandling = `
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  parent_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
  path_part = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"

  request_templates = {
    "application/json" = ""
    "application/xml" = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'static'"
    "integration.request.header.X-Foo" = "'Bar'"
  }

  type = "HTTP"
  uri = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior = "WHEN_NO_MATCH"
	content_handling = "CONVERT_TO_BINARY"
	timeout_milliseconds = 2000
}
`

const testAccAWSAPIGatewayIntegrationConfigRemoveContentHandling = `
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  parent_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
  path_part = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"

  request_templates = {
    "application/json" = ""
    "application/xml" = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'static'"
    "integration.request.header.X-Foo" = "'Bar'"
  }

  type = "HTTP"
  uri = "https://www.google.de"
  integration_http_method = "GET"
	passthrough_behavior = "WHEN_NO_MATCH"
	timeout_milliseconds = 2000
}
`

const testAccAWSAPIGatewayIntegrationConfigUpdateNoTemplates = `
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  parent_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
  path_part = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"

  type = "HTTP"
  uri = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior = "WHEN_NO_MATCH"
	content_handling = "CONVERT_TO_TEXT"
	timeout_milliseconds = 2000
}
`

const testAccAWSAPIGatewayIntegrationConfigCacheKeyParameters = `
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  parent_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
  path_part = "{param}"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.path.param" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"

  request_templates = {
    "application/json" = ""
    "application/xml" = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'static'"
    "integration.request.header.X-Foo" = "'Bar'"
    "integration.request.path.param" = "method.request.path.param"
  }

  cache_key_parameters = ["method.request.path.param"]
  cache_namespace = "foobar"

  type = "HTTP"
  uri = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior = "WHEN_NO_MATCH"
	content_handling = "CONVERT_TO_TEXT"
	timeout_milliseconds = 2000
}
`

func testAccAWSAPIGatewayIntegrationConfig_IntegrationTypeBase(rName string) string {
	return fmt.Sprintf(`
variable "name" {
  default = "%s"
}

data "aws_availability_zones" "test" {}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "${var.name}"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.10.0.0/24"
  availability_zone = "${data.aws_availability_zones.test.names[0]}"
}

resource "aws_api_gateway_rest_api" "test" {
  name = "${var.name}"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  parent_id   = "${aws_api_gateway_rest_api.test.root_resource_id}"
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = "${aws_api_gateway_rest_api.test.id}"
  resource_id   = "${aws_api_gateway_resource.test.id}"
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_lb" "test" {
  name               = "${var.name}"
  internal           = true
  load_balancer_type = "network"
  subnets            = ["${aws_subnet.test.id}"]
}

resource "aws_api_gateway_vpc_link" "test" {
  name        = "${var.name}"
  target_arns = ["${aws_lb.test.arn}"]
}
`, rName)
}

func testAccAWSAPIGatewayIntegrationConfig_IntegrationTypeVpcLink(rName string) string {
	return testAccAWSAPIGatewayIntegrationConfig_IntegrationTypeBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_integration" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  content_handling        = "CONVERT_TO_TEXT"

  connection_type = "VPC_LINK"
  connection_id   = "${aws_api_gateway_vpc_link.test.id}"
}
`)
}

func testAccAWSAPIGatewayIntegrationConfig_IntegrationTypeInternet(rName string) string {
	return testAccAWSAPIGatewayIntegrationConfig_IntegrationTypeBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_integration" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  content_handling        = "CONVERT_TO_TEXT"
}
`)
}
