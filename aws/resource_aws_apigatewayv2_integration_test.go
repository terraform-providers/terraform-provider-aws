package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAPIGatewayV2Integration_basicWebSocket(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "MOCK"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "29000"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
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

func TestAccAWSAPIGatewayV2Integration_basicHttp(t *testing.T) {
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
				Config: testAccAWSAPIGatewayV2IntegrationConfig_httpProxy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_id", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "HTTP_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", ""),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "30000"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
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

func TestAccAWSAPIGatewayV2Integration_DataMappingHttp(t *testing.T) {
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
				Config: testAccAWSAPIGatewayV2IntegrationConfig_dataMappingHttp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_id", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_method", "ANY"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "HTTP_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", "http://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", ""),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.append:header.header1", "$context.requestId"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.remove:querystring.qs1", "''"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "response_parameters.*", map[string]string{
						"status_code":                    "500",
						"mappings.%":                     "2",
						"mappings.append:header.header1": "$context.requestId",
						"mappings.overwrite:statuscode":  "403",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "response_parameters.*", map[string]string{
						"status_code":                  "404",
						"mappings.%":                   "1",
						"mappings.append:header.error": "$stageVariables.environmentId",
					}),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "30000"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_dataMappingHttpUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_id", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_method", "ANY"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "HTTP_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", "http://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", ""),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.append:header.header1", "$context.accountId"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.overwrite:header.header2", "$stageVariables.environmentId"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "response_parameters.*", map[string]string{
						"status_code":                    "500",
						"mappings.%":                     "2",
						"mappings.append:header.header1": "$context.requestId",
						"mappings.overwrite:statuscode":  "403",
					}),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "30000"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
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
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", "http://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.stage", "'value1'"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", "$request.body.name"),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "28999"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
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
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", "http://www.example.org"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_TEMPLATES"),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.x-userid", "'value2'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.path.op", "'value3'"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", "#set($number=42)"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/xml", "#set($percent=$number/100)"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", "$request.body.id"),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "51"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
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

func TestAccAWSAPIGatewayV2Integration_LambdaWebSocket(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetIntegrationOutput
	resourceName := "aws_apigatewayv2_integration.test"
	lambdaResourceName := "aws_lambda_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2IntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_lambdaWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test Lambda"),
					resource.TestCheckResourceAttr(resourceName, "integration_method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", "${integration.response.body.errorMessage}"),
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "AWS"),
					resource.TestCheckResourceAttrPair(resourceName, "integration_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "29000"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
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

func TestAccAWSAPIGatewayV2Integration_LambdaHttp(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetIntegrationOutput
	resourceName := "aws_apigatewayv2_integration.test"
	lambdaResourceName := "aws_lambda_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2IntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_lambdaHttp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test Lambda"),
					resource.TestCheckResourceAttr(resourceName, "integration_method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "AWS_PROXY"),
					resource.TestCheckResourceAttrPair(resourceName, "integration_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", ""),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "2.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "30000"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
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

func TestAccAWSAPIGatewayV2Integration_VpcLinkWebSocket(t *testing.T) {
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
				Config: testAccAWSAPIGatewayV2IntegrationConfig_vpcLinkWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttrPair(resourceName, "connection_id", vpcLinkResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "VPC_LINK"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test VPC Link"),
					resource.TestCheckResourceAttr(resourceName, "integration_method", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "HTTP_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", "http://www.example.net"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "NEVER"),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "12345"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
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

func TestAccAWSAPIGatewayV2Integration_VpcLinkHttp(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetIntegrationOutput
	resourceName := "aws_apigatewayv2_integration.test"
	vpcLinkResourceName := "aws_apigatewayv2_vpc_link.test"
	lbListenerResourceName := "aws_lb_listener.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2IntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_vpcLinkHttp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttrPair(resourceName, "connection_id", vpcLinkResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "VPC_LINK"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test private integration"),
					resource.TestCheckResourceAttr(resourceName, "integration_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "HTTP_PROXY"),
					resource.TestCheckResourceAttrPair(resourceName, "integration_uri", lbListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", ""),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "29001"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.0.server_name_to_verify", "www.example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2IntegrationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_vpcLinkHttpUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttrPair(resourceName, "connection_id", vpcLinkResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "VPC_LINK"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test private integration updated"),
					resource.TestCheckResourceAttr(resourceName, "integration_method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "HTTP_PROXY"),
					resource.TestCheckResourceAttrPair(resourceName, "integration_uri", lbListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", ""),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "29001"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.0.server_name_to_verify", "www.example.org"),
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

func TestAccAWSAPIGatewayV2Integration_AwsServiceIntegration(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetIntegrationOutput
	resourceName := "aws_apigatewayv2_integration.test"
	iamRoleResourceName := "aws_iam_role.test"
	sqsQueue1ResourceName := "aws_sqs_queue.test.0"
	sqsQueue2ResourceName := "aws_sqs_queue.test.1"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2IntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_sqsIntegration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", ""),
					resource.TestCheckResourceAttrPair(resourceName, "credentials_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test SQS send"),
					resource.TestCheckResourceAttr(resourceName, "integration_method", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", "SQS-SendMessage"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "AWS_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", ""),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.MessageBody", "$request.body"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.MessageGroupId", "$request.body.authentication_key"),
					resource.TestCheckResourceAttrPair(resourceName, "request_parameters.QueueUrl", sqsQueue1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "30000"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2IntegrationConfig_sqsIntegration(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2IntegrationExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", ""),
					resource.TestCheckResourceAttrPair(resourceName, "credentials_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test SQS send"),
					resource.TestCheckResourceAttr(resourceName, "integration_method", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "integration_subtype", "SQS-SendMessage"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "AWS_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", ""),
					resource.TestCheckResourceAttr(resourceName, "payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.MessageBody", "$request.body"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.MessageGroupId", "$request.body.authentication_key"),
					resource.TestCheckResourceAttrPair(resourceName, "request_parameters.QueueUrl", sqsQueue2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "30000"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
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

func testAccAWSAPIGatewayV2IntegrationConfig_apiHttp(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
}
`, rName)
}

func testAccAWSAPIGatewayV2IntegrationConfig_lambdaBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["sts:AssumeRole"],
    "Principal": {"Service": "lambda.amazonaws.com"}
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ],
    "Resource": ["*"]
  }]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "index.handler"
  runtime       = "nodejs10.x"

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_lambda_permission" "test" {
  action        = "lambda:*"
  function_name = aws_lambda_function.test.arn
  principal     = "apigateway.amazonaws.com"
}
`, rName)
}

func testAccAWSAPIGatewayV2IntegrationConfig_vpcLinkHttpBase(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2IntegrationConfig_apiHttp(rName),
		testAccAWSAPIGatewayV2VpcLinkConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_lb" "test" {
  name = %[1]q

  internal           = true
  load_balancer_type = "network"
  subnets            = aws_subnet.test.*.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 80
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  health_check {
    port     = 80
    protocol = "TCP"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.arn
  port              = "80"
  protocol          = "TCP"

  default_action {
    target_group_arn = aws_lb_target_group.test.arn
    type             = "forward"
  }
}
`, rName))
}

func testAccAWSAPIGatewayV2IntegrationConfig_basic(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_apiWebSocket(rName) + `
resource "aws_apigatewayv2_integration" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "MOCK"
}
`
}

func testAccAWSAPIGatewayV2IntegrationConfig_dataMappingHttp(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_apiHttp(rName) + `
resource "aws_apigatewayv2_integration" "test" {
  api_id = aws_apigatewayv2_api.test.id

  integration_type   = "HTTP_PROXY"
  integration_method = "ANY"
  integration_uri    = "http://www.example.com"

  request_parameters = {
    "append:header.header1"  = "$context.requestId"
    "remove:querystring.qs1" = "''"
  }

  response_parameters {
    status_code = "500"

    mappings = {
      "append:header.header1" = "$context.requestId"
      "overwrite:statuscode"  = "403"
    }
  }

  response_parameters {
    status_code = "404"

    mappings = {
      "append:header.error" = "$stageVariables.environmentId"
    }
  }
}
`
}

func testAccAWSAPIGatewayV2IntegrationConfig_dataMappingHttpUpdated(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_apiHttp(rName) + `
resource "aws_apigatewayv2_integration" "test" {
  api_id = aws_apigatewayv2_api.test.id

  integration_type   = "HTTP_PROXY"
  integration_method = "ANY"
  integration_uri    = "http://www.example.com"

  request_parameters = {
    "append:header.header1"    = "$context.accountId"
    "overwrite:header.header2" = "$stageVariables.environmentId"
  }

  response_parameters {
    status_code = "500"

    mappings = {
      "append:header.header1" = "$context.requestId"
      "overwrite:statuscode"  = "403"
    }
  }
}
`
}

func testAccAWSAPIGatewayV2IntegrationConfig_integrationTypeHttp(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_apiWebSocket(rName) + `
resource "aws_apigatewayv2_integration" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "HTTP"

  connection_type               = "INTERNET"
  content_handling_strategy     = "CONVERT_TO_TEXT"
  description                   = "Test HTTP"
  integration_method            = "GET"
  integration_uri               = "http://www.example.com"
  passthrough_behavior          = "WHEN_NO_MATCH"
  template_selection_expression = "$request.body.name"
  timeout_milliseconds          = 28999

  request_parameters = {
    "integration.request.querystring.stage" = "'value1'"
  }

  request_templates = {
    "application/json" = ""
  }
}
`
}

func testAccAWSAPIGatewayV2IntegrationConfig_integrationTypeHttpUpdated(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_apiWebSocket(rName) + `
resource "aws_apigatewayv2_integration" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "HTTP"

  connection_type               = "INTERNET"
  content_handling_strategy     = "CONVERT_TO_BINARY"
  description                   = "Test HTTP updated"
  integration_method            = "POST"
  integration_uri               = "http://www.example.org"
  passthrough_behavior          = "WHEN_NO_TEMPLATES"
  template_selection_expression = "$request.body.id"
  timeout_milliseconds          = 51

  request_parameters = {
    "integration.request.header.x-userid" = "'value2'"
    "integration.request.path.op"         = "'value3'"
  }

  request_templates = {
    "application/json" = "#set($number=42)"
    "application/xml"  = "#set($percent=$number/100)"
  }
}
`
}

func testAccAWSAPIGatewayV2IntegrationConfig_lambdaWebSocket(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2IntegrationConfig_apiWebSocket(rName),
		testAccAWSAPIGatewayV2IntegrationConfig_lambdaBase(rName),
		`
resource "aws_apigatewayv2_integration" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "AWS"

  connection_type           = "INTERNET"
  content_handling_strategy = "CONVERT_TO_TEXT"
  description               = "Test Lambda"
  integration_uri           = aws_lambda_function.test.invoke_arn
  passthrough_behavior      = "WHEN_NO_MATCH"

  depends_on = [aws_lambda_permission.test]
}
`)
}

func testAccAWSAPIGatewayV2IntegrationConfig_lambdaHttp(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2IntegrationConfig_apiHttp(rName),
		testAccAWSAPIGatewayV2IntegrationConfig_lambdaBase(rName),
		`
resource "aws_apigatewayv2_integration" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "AWS_PROXY"

  connection_type = "INTERNET"
  description     = "Test Lambda"
  integration_uri = aws_lambda_function.test.invoke_arn

  payload_format_version = "2.0"

  depends_on = [aws_lambda_permission.test]
}
`)
}

func testAccAWSAPIGatewayV2IntegrationConfig_httpProxy(rName string) string {
	return composeConfig(testAccAWSAPIGatewayV2IntegrationConfig_apiHttp(rName), `
resource "aws_apigatewayv2_integration" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "HTTP_PROXY"

  integration_method = "GET"
  integration_uri    = "https://example.com"
}
`)
}

func testAccAWSAPIGatewayV2IntegrationConfig_vpcLinkHttp(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2IntegrationConfig_vpcLinkHttpBase(rName),
		`
resource "aws_apigatewayv2_integration" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "HTTP_PROXY"

  connection_type      = "VPC_LINK"
  connection_id        = aws_apigatewayv2_vpc_link.test.id
  description          = "Test private integration"
  integration_method   = "GET"
  integration_uri      = aws_lb_listener.test.arn
  timeout_milliseconds = 29001

  tls_config {
    server_name_to_verify = "www.example.com"
  }
}
`)
}

func testAccAWSAPIGatewayV2IntegrationConfig_vpcLinkHttpUpdated(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2IntegrationConfig_vpcLinkHttpBase(rName),
		`
resource "aws_apigatewayv2_integration" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "HTTP_PROXY"

  connection_type    = "VPC_LINK"
  connection_id      = aws_apigatewayv2_vpc_link.test.id
  description        = "Test private integration updated"
  integration_method = "POST"
  integration_uri    = aws_lb_listener.test.arn

  tls_config {
    server_name_to_verify = "www.example.org"
  }
}
`)
}

func testAccAWSAPIGatewayV2IntegrationConfig_vpcLinkWebSocket(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2IntegrationConfig_apiWebSocket(rName),
		fmt.Sprintf(`
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
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.10.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "network"
  subnets            = [aws_subnet.test.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_api_gateway_vpc_link" "test" {
  name        = %[1]q
  target_arns = [aws_lb.test.arn]
}

resource "aws_apigatewayv2_integration" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "HTTP_PROXY"

  connection_id             = aws_api_gateway_vpc_link.test.id
  connection_type           = "VPC_LINK"
  content_handling_strategy = "CONVERT_TO_TEXT"
  description               = "Test VPC Link"
  integration_method        = "PUT"
  integration_uri           = "http://www.example.net"
  passthrough_behavior      = "NEVER"
  timeout_milliseconds      = 12345
}
`, rName))
}

func testAccAWSAPIGatewayV2IntegrationConfig_sqsIntegration(rName string, queueIndex int) string {
	return composeConfig(
		testAccAWSAPIGatewayV2IntegrationConfig_apiHttp(rName),
		fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"Service": "apigateway.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["sqs:*"],
    "Resource": "*"
  }]
}
EOF
}

resource "aws_sqs_queue" "test" {
  count = 2

  name = "%[1]s-${count.index}"
}

resource "aws_apigatewayv2_integration" "test" {
  api_id              = aws_apigatewayv2_api.test.id
  credentials_arn     = aws_iam_role.test.arn
  description         = "Test SQS send"
  integration_type    = "AWS_PROXY"
  integration_subtype = "SQS-SendMessage"

  request_parameters = {
    "QueueUrl"       = aws_sqs_queue.test.%[2]d.id
    "MessageGroupId" = "$request.body.authentication_key"
    "MessageBody"    = "$request.body"
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, queueIndex))
}
