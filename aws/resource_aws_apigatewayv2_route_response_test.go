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

func TestAccAWSAPIGatewayV2RouteResponse_basic(t *testing.T) {
	var apiId, routeId string
	var v apigatewayv2.GetRouteResponseOutput
	resourceName := "aws_apigatewayv2_route_response.test"
	routeResourceName := "aws_apigatewayv2_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2RouteResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2RouteResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteResponseExists(resourceName, &apiId, &routeId, &v),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "response_models.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "route_id", routeResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_response_key", "$default"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2RouteResponseImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2RouteResponse_disappears(t *testing.T) {
	var apiId, routeId string
	var v apigatewayv2.GetRouteResponseOutput
	resourceName := "aws_apigatewayv2_route_response.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2RouteResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2RouteResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteResponseExists(resourceName, &apiId, &routeId, &v),
					testAccCheckAWSAPIGatewayV2RouteResponseDisappears(&apiId, &routeId, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2RouteResponse_Model(t *testing.T) {
	var apiId, routeId string
	var v apigatewayv2.GetRouteResponseOutput
	resourceName := "aws_apigatewayv2_route_response.test"
	modelResourceName := "aws_apigatewayv2_model.test"
	routeResourceName := "aws_apigatewayv2_route.test"
	// Model name must be alphanumeric.
	rName := strings.ReplaceAll(acctest.RandomWithPrefix("tf-acc-test"), "-", "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2RouteResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2RouteResponseConfig_model(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2RouteResponseExists(resourceName, &apiId, &routeId, &v),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", "action"),
					resource.TestCheckResourceAttr(resourceName, "response_models.%", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "response_models.test", modelResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "route_id", routeResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_response_key", "$default"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2RouteResponseImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayV2RouteResponseDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_route_response" {
			continue
		}

		_, err := conn.GetRouteResponse(&apigatewayv2.GetRouteResponseInput{
			ApiId:           aws.String(rs.Primary.Attributes["api_id"]),
			RouteId:         aws.String(rs.Primary.Attributes["route_id"]),
			RouteResponseId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 route response %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGatewayV2RouteResponseDisappears(apiId, routeId *string, v *apigatewayv2.GetRouteResponseOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.DeleteRouteResponse(&apigatewayv2.DeleteRouteResponseInput{
			ApiId:           apiId,
			RouteId:         routeId,
			RouteResponseId: v.RouteResponseId,
		})

		return err
	}
}

func testAccCheckAWSAPIGatewayV2RouteResponseExists(n string, vApiId, vRouteId *string, v *apigatewayv2.GetRouteResponseOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 route response ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		apiId := aws.String(rs.Primary.Attributes["api_id"])
		routeId := aws.String(rs.Primary.Attributes["route_id"])
		resp, err := conn.GetRouteResponse(&apigatewayv2.GetRouteResponseInput{
			ApiId:           apiId,
			RouteId:         routeId,
			RouteResponseId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*vApiId = *apiId
		*vRouteId = *routeId
		*v = *resp

		return nil
	}
}

func testAccAWSAPIGatewayV2RouteResponseImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.Attributes["route_id"], rs.Primary.ID), nil
	}
}

func testAccAWSAPIGatewayV2RouteResponseConfig_basic(rName string) string {
	return testAccAWSAPIGatewayV2RouteConfig_basic(rName) + `
resource "aws_apigatewayv2_route_response" "test" {
  api_id             = aws_apigatewayv2_api.test.id
  route_id           = aws_apigatewayv2_route.test.id
  route_response_key = "$default"
}
`
}

func testAccAWSAPIGatewayV2RouteResponseConfig_model(rName string) string {
	return testAccAWSAPIGatewayV2RouteConfig_model(rName) + `
resource "aws_apigatewayv2_route_response" "test" {
  api_id             = aws_apigatewayv2_api.test.id
  route_id           = aws_apigatewayv2_route.test.id
  route_response_key = "$default"

  model_selection_expression = "action"

  response_models = {
    "test" = aws_apigatewayv2_model.test.name
  }
}
`
}
