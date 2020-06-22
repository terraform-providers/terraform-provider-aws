package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSAPIGatewayV2Stage_basicWebSocket(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(fmt.Sprintf("/apis/.+/stages/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(fmt.Sprintf(".+/%s", rName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2StageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Stage_basicHttp(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_basicHttp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(fmt.Sprintf("/apis/.+/stages/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "execution_arn", ""),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2StageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Stage_defaultHttpStage(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_defaultHttpStage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+/stages/\$default`)),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "execution_arn", ""),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/", testAccGetRegion()))),
					resource.TestCheckResourceAttr(resourceName, "name", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2StageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Stage_disappears(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					testAccCheckAWSAPIGatewayV2StageDisappears(&apiId, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Stage_AccessLogSettings(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	cloudWatchResourceName := "aws_cloudwatch_log_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_accessLogSettings(rName, "$context.identity.sourceIp $context.requestId"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					testAccCheckAWSAPIGatewayV2StageAccessLogDestinationArn(resourceName, "access_log_settings.0.destination_arn", cloudWatchResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", "$context.identity.sourceIp $context.requestId"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(fmt.Sprintf("/apis/.+/stages/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(fmt.Sprintf(".+/%s", rName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2StageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2StageConfig_accessLogSettings(rName, "$context.requestId"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					testAccCheckAWSAPIGatewayV2StageAccessLogDestinationArn(resourceName, "access_log_settings.0.destination_arn", cloudWatchResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", "$context.requestId"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Stage_ClientCertificateIdAndDescription(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	certificateResourceName := "aws_api_gateway_client_certificate.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_clientCertificateIdAndDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(fmt.Sprintf("/apis/.+/stages/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "client_certificate_id", certificateResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test stage"),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(fmt.Sprintf(".+/%s", rName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2StageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2StageConfig_clientCertificateIdAndDescriptionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(fmt.Sprintf("/apis/.+/stages/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "true"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test stage updated"),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(fmt.Sprintf(".+/%s", rName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Stage_DefaultRouteSettings(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_defaultRouteSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(fmt.Sprintf("/apis/.+/stages/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "2222"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "8888"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(fmt.Sprintf(".+/%s", rName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2StageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2StageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Stage_Deployment(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	deploymentResourceName := "aws_apigatewayv2_deployment.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_deployment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(fmt.Sprintf("/apis/.+/stages/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_id", deploymentResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(fmt.Sprintf(".+/%s", rName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2StageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Stage_RouteSettings(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckAWSAPIGatewayV2StageDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_routeSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(fmt.Sprintf("/apis/.+/stages/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(fmt.Sprintf(".+/%s", rName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.4093656941.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.4093656941.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.4093656941.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.4093656941.route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.4093656941.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.4093656941.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.3867839051.data_trace_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.3867839051.detailed_metrics_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.3867839051.logging_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.3867839051.route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.3867839051.throttling_burst_limit", "2222"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.3867839051.throttling_rate_limit", "8888"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2StageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2StageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Stage_StageVariables(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_stageVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(fmt.Sprintf("/apis/.+/stages/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(fmt.Sprintf(".+/%s", rName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.Var1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.Var2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2StageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2StageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Stage_Tags(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(fmt.Sprintf("/apis/.+/stages/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(fmt.Sprintf(".+/%s", rName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2StageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2StageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayV2StageDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_stage" {
			continue
		}

		_, err := conn.GetStage(&apigatewayv2.GetStageInput{
			ApiId:     aws.String(rs.Primary.Attributes["api_id"]),
			StageName: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 stage %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGatewayV2StageDisappears(apiId *string, v *apigatewayv2.GetStageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.DeleteStage(&apigatewayv2.DeleteStageInput{
			ApiId:     apiId,
			StageName: v.StageName,
		})

		return err
	}
}

func testAccCheckAWSAPIGatewayV2StageExists(n string, vApiId *string, v *apigatewayv2.GetStageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 stage ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		apiId := aws.String(rs.Primary.Attributes["api_id"])
		resp, err := conn.GetStage(&apigatewayv2.GetStageInput{
			ApiId:     apiId,
			StageName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*vApiId = *apiId
		*v = *resp

		return nil
	}
}

func testAccAWSAPIGatewayV2StageImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccCheckAWSAPIGatewayV2StageAccessLogDestinationArn(resourceName, resourceKey, cloudWatchResourceName, cloudWatchResourceKey string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cwRs, ok := s.RootModule().Resources[cloudWatchResourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", cloudWatchResourceName)
		}
		cwArn, ok := cwRs.Primary.Attributes[cloudWatchResourceKey]
		if !ok {
			return fmt.Errorf("Attribute %q not found in resource %s", cloudWatchResourceKey, cloudWatchResourceName)
		}

		return resource.TestCheckResourceAttr(resourceName, resourceKey, strings.TrimSuffix(cwArn, ":*"))(s)
	}
}

func testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
`, rName)
}

func testAccAWSAPIGatewayV2StageConfig_apiHttp(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
}
`, rName)
}

func testAccAWSAPIGatewayV2StageConfig_basicWebSocket(rName string) string {
	return testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = "${aws_apigatewayv2_api.test.id}"
  name   = %[1]q
}
`, rName)
}

func testAccAWSAPIGatewayV2StageConfig_basicHttp(rName string) string {
	return testAccAWSAPIGatewayV2StageConfig_apiHttp(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = "${aws_apigatewayv2_api.test.id}"
  name   = %[1]q
}
`, rName)
}

func testAccAWSAPIGatewayV2StageConfig_defaultHttpStage(rName string) string {
	return testAccAWSAPIGatewayV2StageConfig_apiHttp(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = "${aws_apigatewayv2_api.test.id}"
  name   = "$default"
}
`)
}

func testAccAWSAPIGatewayV2StageConfig_accessLogSettings(rName, format string) string {
	return testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName) + fmt.Sprintf(`
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
  role = "${aws_iam_role.test.id}"

policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
      "logs:PutLogEvents",
      "logs:GetLogEvents",
      "logs:FilterLogEvents"
    ],
    "Resource": "*"
  }]
}
EOF
}

resource "aws_api_gateway_account" "test" {
  cloudwatch_role_arn = "${aws_iam_role.test.arn}"
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_apigatewayv2_stage" "test" {
  api_id = "${aws_apigatewayv2_api.test.id}"
  name   = %[1]q

  access_log_settings {
    destination_arn = "${aws_cloudwatch_log_group.test.arn}"
    format          = %[2]q
  }

  depends_on = ["aws_api_gateway_account.test"]
}
`, rName, format)
}

func testAccAWSAPIGatewayV2StageConfig_clientCertificateIdAndDescription(rName string) string {
	return testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName) + fmt.Sprintf(`
resource "aws_api_gateway_client_certificate" "test" {
  description = %[1]q
}

resource "aws_apigatewayv2_stage" "test" {
  api_id = "${aws_apigatewayv2_api.test.id}"
  name   = %[1]q

  client_certificate_id = "${aws_api_gateway_client_certificate.test.id}"
  description           = "Test stage"
}
`, rName)
}

func testAccAWSAPIGatewayV2StageConfig_clientCertificateIdAndDescriptionUpdated(rName string) string {
	return testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName) + fmt.Sprintf(`
resource "aws_api_gateway_client_certificate" "test" {
  description = %[1]q
}

resource "aws_apigatewayv2_stage" "test" {
  api_id = "${aws_apigatewayv2_api.test.id}"
  name   = %[1]q

  description = "Test stage updated"
  auto_deploy = true
}
`, rName)
}

func testAccAWSAPIGatewayV2StageConfig_defaultRouteSettings(rName string) string {
	return testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = "${aws_apigatewayv2_api.test.id}"
  name   = %[1]q

  default_route_settings {
    data_trace_enabled       = true
    detailed_metrics_enabled = true
    logging_level            = "ERROR"
    throttling_burst_limit   = 2222
    throttling_rate_limit    = 8888
  }
}
`, rName)
}

func testAccAWSAPIGatewayV2StageConfig_deployment(rName string) string {
	return testAccAWSAPIGatewayV2DeploymentConfig_basic(rName, rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = "${aws_apigatewayv2_api.test.id}"
  name   = %[1]q

  deployment_id = "${aws_apigatewayv2_deployment.test.id}"
}
`, rName)
}

func testAccAWSAPIGatewayV2StageConfig_routeSettings(rName string) string {
	return testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = "${aws_apigatewayv2_api.test.id}"
  name   = %[1]q

  route_settings {
    route_key = "$default"
  }

  route_settings {
    route_key = "$connect"

    data_trace_enabled       = true
    detailed_metrics_enabled = true
    logging_level            = "ERROR"
    throttling_burst_limit   = 2222
    throttling_rate_limit    = 8888
  }
}
`, rName)
}

func testAccAWSAPIGatewayV2StageConfig_stageVariables(rName string) string {
	return testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = "${aws_apigatewayv2_api.test.id}"
  name   = %[1]q

  stage_variables = {
    Var1 = "Value1"
    Var2 = "Value2"
  }
}
`, rName)
}

func testAccAWSAPIGatewayV2StageConfig_tags(rName string) string {
	return testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = "${aws_apigatewayv2_api.test.id}"
  name   = %[1]q

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
}
`, rName)
}
