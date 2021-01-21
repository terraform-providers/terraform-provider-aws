package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
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

func TestAccAWSAPIGatewayV2Stage_autoDeployHttp(t *testing.T) {
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
				Config: testAccAWSAPIGatewayV2StageConfig_autoDeployHttp(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2StageConfig_autoDeployHttp(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "true"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					// The stage's DeploymentId attribute will be set asynchronously as deployments are done automatically.
					// resource.TestCheckResourceAttrSet(resourceName, "deployment_id"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccAWSAPIGatewayV2StageImportStateIdFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deployment_id"},
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
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayV2Stage(), resourceName),
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
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAPIGatewayAccountCloudWatchRoleArn(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_accessLogSettings(rName, "$context.identity.sourceIp $context.requestId"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudWatchResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", "$context.identity.sourceIp $context.requestId"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
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
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudWatchResourceName, "arn"),
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
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test stage updated"),
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
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

func TestAccAWSAPIGatewayV2Stage_DefaultRouteSettingsWebSocket(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAPIGatewayAccountCloudWatchRoleArn(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_defaultRouteSettingsWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2StageConfig_defaultRouteSettingsWebSocketUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "1111"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "9999"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
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
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "INFO"), // No drift detection if not configured
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

func TestAccAWSAPIGatewayV2Stage_DefaultRouteSettingsHttp(t *testing.T) {
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
				Config: testAccAWSAPIGatewayV2StageConfig_defaultRouteSettingsHttp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "2222"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "8888"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2StageConfig_defaultRouteSettingsHttpUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "1111"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "9999"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
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
			{
				Config: testAccAWSAPIGatewayV2StageConfig_basicHttp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
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
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
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

func TestAccAWSAPIGatewayV2Stage_RouteSettingsWebSocket(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAPIGatewayAccountCloudWatchRoleArn(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2StageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2StageConfig_routeSettingsWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"route_key":                "$default",
						"throttling_burst_limit":   "0",
						"throttling_rate_limit":    "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "true",
						"detailed_metrics_enabled": "true",
						"logging_level":            "ERROR",
						"route_key":                "$connect",
						"throttling_burst_limit":   "2222",
						"throttling_rate_limit":    "8888",
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2StageConfig_routeSettingsWebSocketUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"route_key":                "$default",
						"throttling_burst_limit":   "1111",
						"throttling_rate_limit":    "9999",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"logging_level":            "INFO",
						"route_key":                "$connect",
						"throttling_burst_limit":   "0",
						"throttling_rate_limit":    "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"route_key":                "$disconnect",
						"throttling_burst_limit":   "0",
						"throttling_rate_limit":    "0",
					}),
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

func TestAccAWSAPIGatewayV2Stage_RouteSettingsHttp(t *testing.T) {
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
				Config: testAccAWSAPIGatewayV2StageConfig_routeSettingsHttp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "true",
						"route_key":                "$default",
						"throttling_burst_limit":   "2222",
						"throttling_rate_limit":    "8888",
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2StageConfig_routeSettingsHttpUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"route_key":                "$default",
						"throttling_burst_limit":   "1111",
						"throttling_rate_limit":    "9999",
					}),
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
				Config: testAccAWSAPIGatewayV2StageConfig_basicHttp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
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
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Stage_RouteSettingsHttp_WithRoute(t *testing.T) {
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
				Config: testAccAWSAPIGatewayV2StageConfig_routeSettingsHttpWithRoute(rName, "GET /first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"route_key":                "GET /first",
						"throttling_burst_limit":   "1",
						"throttling_rate_limit":    "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2StageConfig_routeSettingsHttpWithRoute(rName, "POST /second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2StageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"route_key":                "POST /second",
						"throttling_burst_limit":   "1",
						"throttling_rate_limit":    "0",
					}),
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
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageARN(resourceName, "arn", &apiId, &v),
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
					testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, "execution_arn", &apiId, &v),
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

func testAccCheckAWSAPIGatewayV2StageARN(resourceName, attributeName string, vApiId *string, v *apigatewayv2.GetStageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckResourceAttrRegionalARNNoAccount(resourceName, attributeName, "apigateway", fmt.Sprintf("/apis/%s/stages/%s", *vApiId, *v.StageName))(s)
	}
}

func testAccCheckAWSAPIGatewayV2StageExecutionARN(resourceName, attributeName string, vApiId *string, v *apigatewayv2.GetStageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckResourceAttrRegionalARN(resourceName, attributeName, "execute-api", fmt.Sprintf("%s/%s", *vApiId, *v.StageName))(s)
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
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_basicHttp(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiHttp(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_defaultHttpStage(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiHttp(rName),
		`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = "$default"
}
`)
}

func testAccAWSAPIGatewayV2StageConfig_autoDeployHttp(rName string, autoDeploy bool) string {
	return composeConfig(
		testAccAWSAPIGatewayV2IntegrationConfig_httpProxy(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "GET /test"
  target    = "integrations/${aws_apigatewayv2_integration.test.id}"
}

resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  auto_deploy = %[2]t
}
`, rName, autoDeploy))
}

func testAccAWSAPIGatewayV2StageConfig_accessLogSettings(rName, format string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.test.arn
    format          = %[2]q
  }
}
`, rName, format))
}

func testAccAWSAPIGatewayV2StageConfig_clientCertificateIdAndDescription(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_api_gateway_client_certificate" "test" {
  description = %[1]q
}

resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  client_certificate_id = aws_api_gateway_client_certificate.test.id
  description           = "Test stage"
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_clientCertificateIdAndDescriptionUpdated(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_api_gateway_client_certificate" "test" {
  description = %[1]q
}

resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  description = "Test stage updated"
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_defaultRouteSettingsWebSocket(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  default_route_settings {
    data_trace_enabled       = true
    detailed_metrics_enabled = true
    logging_level            = "ERROR"
    throttling_burst_limit   = 2222
    throttling_rate_limit    = 8888
  }
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_defaultRouteSettingsWebSocketUpdated(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  default_route_settings {
    data_trace_enabled       = false
    detailed_metrics_enabled = true
    logging_level            = "INFO"
    throttling_burst_limit   = 1111
    throttling_rate_limit    = 9999
  }
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_defaultRouteSettingsHttp(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiHttp(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  default_route_settings {
    detailed_metrics_enabled = true
    throttling_burst_limit   = 2222
    throttling_rate_limit    = 8888
  }
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_defaultRouteSettingsHttpUpdated(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiHttp(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  default_route_settings {
    throttling_burst_limit = 1111
    throttling_rate_limit  = 9999
  }
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_deployment(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2DeploymentConfig_basic(rName, rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  deployment_id = aws_apigatewayv2_deployment.test.id
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_routeSettingsWebSocket(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
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
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_routeSettingsWebSocketUpdated(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  route_settings {
    route_key = "$default"

    throttling_burst_limit = 1111
    throttling_rate_limit  = 9999
  }

  route_settings {
    route_key = "$connect"

    logging_level = "INFO"
  }

  route_settings {
    route_key = "$disconnect"
  }
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_routeSettingsHttp(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiHttp(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  route_settings {
    route_key = "$default"

    detailed_metrics_enabled = true
    throttling_burst_limit   = 2222
    throttling_rate_limit    = 8888
  }
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_routeSettingsHttpUpdated(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiHttp(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  route_settings {
    route_key = "$default"

    throttling_burst_limit = 1111
    throttling_rate_limit  = 9999
  }
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_routeSettingsHttpWithRoute(rName, routeKey string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiHttp(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  route_settings {
    route_key = aws_apigatewayv2_route.test.route_key

    throttling_burst_limit = 1
  }
}

resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = %[2]q
  target    = "integrations/${aws_apigatewayv2_integration.test.id}"
}

resource "aws_apigatewayv2_integration" "test" {
  api_id             = aws_apigatewayv2_api.test.id
  integration_type   = "HTTP_PROXY"
  integration_method = "GET"
  integration_uri    = "https://example.com/"
}
`, rName, routeKey))
}

func testAccAWSAPIGatewayV2StageConfig_stageVariables(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  stage_variables = {
    Var1 = "Value1"
    Var2 = "Value2"
  }
}
`, rName))
}

func testAccAWSAPIGatewayV2StageConfig_tags(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2StageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
}
`, rName))
}
