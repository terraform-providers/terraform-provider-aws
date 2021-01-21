package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAPIGatewayV2Route_basic(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2RouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Route_disappears(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayV2Route(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Route_Authorizer(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	authorizerResourceName := "aws_apigatewayv2_authorizer.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_authorizer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeCustom),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_id", authorizerResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2RouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_authorizerUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeAwsIam),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Route_JwtAuthorization(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	authorizerResourceName := "aws_apigatewayv2_authorizer.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_jwtAuthorization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeJwt),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_id", authorizerResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "GET /test"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2RouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_jwtAuthorizationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeJwt),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_id", authorizerResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "GET /test"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Route_Model(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	modelResourceName := "aws_apigatewayv2_model.test"
	// Model name must be alphanumeric.
	rName := strings.ReplaceAll(acctest.RandomWithPrefix("tf-acc-test"), "-", "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_model(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", "action"),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "request_models.test", modelResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2RouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Route_SimpleAttributes(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_simpleAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "true"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", "GET"),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", "$default"),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_simpleAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "true"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", "GET"),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", "$default"),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2RouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Route_Target(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	integrationResourceName := "aws_apigatewayv2_integration.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_target(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					testAccCheckAWSAPIGatewayV2RouteTarget(resourceName, integrationResourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2RouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Route_UpdateRouteKey(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_routeKey(rName, "GET /path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "GET /path"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2RouteConfig_routeKey(rName, "POST /new/path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "POST /new/path"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2RouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayV2RouteDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_route" {
			continue
		}

		_, err := conn.GetRoute(&apigatewayv2.GetRouteInput{
			ApiId:   aws.String(rs.Primary.Attributes["api_id"]),
			RouteId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 route %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGatewayV2RouteExists(n string, vApiId *string, v *apigatewayv2.GetRouteOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 route ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		apiId := aws.String(rs.Primary.Attributes["api_id"])
		resp, err := conn.GetRoute(&apigatewayv2.GetRouteInput{
			ApiId:   apiId,
			RouteId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*vApiId = *apiId
		*v = *resp

		return nil
	}
}

func testAccAWSAPIGatewayV2RouteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccCheckAWSAPIGatewayV2RouteTarget(resourceName, integrationResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[integrationResourceName]
		if !ok {
			return fmt.Errorf("Not Found: %s", integrationResourceName)
		}

		return resource.TestCheckResourceAttr(resourceName, "target", fmt.Sprintf("integrations/%s", rs.Primary.ID))(s)
	}
}

func testAccAWSAPIGatewayV2RouteConfig_apiWebSocket(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
`, rName)
}

func testAccAWSAPIGatewayV2RouteConfig_apiHttp(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
}
`, rName)
}

func testAccAWSAPIGatewayV2RouteConfig_basic(rName string) string {
	return testAccAWSAPIGatewayV2RouteConfig_apiWebSocket(rName) + `
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$default"
}
`
}

func testAccAWSAPIGatewayV2RouteConfig_authorizer(rName string) string {
	return testAccAWSAPIGatewayV2AuthorizerConfig_basic(rName) + `
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$connect"

  authorization_type = "CUSTOM"
  authorizer_id      = aws_apigatewayv2_authorizer.test.id
}
`
}

func testAccAWSAPIGatewayV2RouteConfig_authorizerUpdated(rName string) string {
	return testAccAWSAPIGatewayV2AuthorizerConfig_basic(rName) + `
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$connect"

  authorization_type = "AWS_IAM"
}
`
}

func testAccAWSAPIGatewayV2RouteConfig_jwtAuthorization(rName string) string {
	return testAccAWSAPIGatewayV2AuthorizerConfig_jwt(rName) + `
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "GET /test"

  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.test.id

  authorization_scopes = ["user.id", "user.email"]
}
`
}

func testAccAWSAPIGatewayV2RouteConfig_jwtAuthorizationUpdated(rName string) string {
	return testAccAWSAPIGatewayV2AuthorizerConfig_jwt(rName) + `
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "GET /test"

  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.test.id

  authorization_scopes = ["user.email"]
}
`
}

func testAccAWSAPIGatewayV2RouteConfig_model(rName string) string {
	schema := `
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "ExampleModel",
  "type": "object",
  "properties": {
    "id": {
      "type": "string"
    }
  }
}
`

	return testAccAWSAPIGatewayV2ModelConfig_basic(rName, schema) + `
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$default"

  model_selection_expression = "action"

  request_models = {
    "test" = aws_apigatewayv2_model.test.name
  }
}
`
}

func testAccAWSAPIGatewayV2RouteConfig_routeKey(rName, routeKey string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2RouteConfig_apiHttp(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = %[1]q
}
`, routeKey))
}

// Simple attributes - No authorization, models or targets.
func testAccAWSAPIGatewayV2RouteConfig_simpleAttributes(rName string) string {
	return testAccAWSAPIGatewayV2RouteConfig_apiWebSocket(rName) + `
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$default"

  api_key_required                    = true
  operation_name                      = "GET"
  route_response_selection_expression = "$default"
}
`
}

func testAccAWSAPIGatewayV2RouteConfig_target(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_basic(rName) + `
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$default"

  target = "integrations/${aws_apigatewayv2_integration.test.id}"
}
`
}
