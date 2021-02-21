package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAPIGatewayStage_basic(t *testing.T) {
	var conf apigateway.Stage
	rName := acctest.RandString(5)
	resourceName := "aws_api_gateway_stage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "prod"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_size", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-test"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "invoke_url"),
					resource.TestCheckResourceAttr(resourceName, "xray_tracing_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayStageImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayStageConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "prod"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-test"),
					resource.TestCheckResourceAttr(resourceName, "tags.ExtraName", "tf-test"),
					resource.TestCheckResourceAttr(resourceName, "xray_tracing_enabled", "false"),
				),
			},
			{
				Config: testAccAWSAPIGatewayStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "prod"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_size", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-test"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "invoke_url"),
					resource.TestCheckResourceAttr(resourceName, "xray_tracing_enabled", "true"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12756
func TestAccAWSAPIGatewayStage_disappears_ReferencingDeployment(t *testing.T) {
	var stage apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_stage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayStageConfigReferencingDeployment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &stage),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayStageImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayStage_disappears(t *testing.T) {
	var stage apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_stage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayStageConfigReferencingDeployment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &stage),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayStage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayStage_accessLogSettings(t *testing.T) {
	var conf apigateway.Stage
	rName := acctest.RandString(5)
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_api_gateway_stage.test"
	clf := `$context.identity.sourceIp $context.identity.caller $context.identity.user [$context.requestTime] "$context.httpMethod $context.resourcePath $context.protocol" $context.status $context.responseLength $context.requestId`
	json := `{ "requestId":"$context.requestId", "ip": "$context.identity.sourceIp", "caller":"$context.identity.caller", "user":"$context.identity.user", "requestTime":"$context.requestTime", "httpMethod":"$context.httpMethod", "resourcePath":"$context.resourcePath", "status":"$context.status", "protocol":"$context.protocol", "responseLength":"$context.responseLength" }`
	xml := `<request id="$context.requestId"> <ip>$context.identity.sourceIp</ip> <caller>$context.identity.caller</caller> <user>$context.identity.user</user> <requestTime>$context.requestTime</requestTime> <httpMethod>$context.httpMethod</httpMethod> <resourcePath>$context.resourcePath</resourcePath> <status>$context.status</status> <protocol>$context.protocol</protocol> <responseLength>$context.responseLength</responseLength> </request>`
	csv := `$context.identity.sourceIp,$context.identity.caller,$context.identity.user,$context.requestTime,$context.httpMethod,$context.resourcePath,$context.protocol,$context.status,$context.responseLength,$context.requestId`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayStageConfig_accessLogSettings(rName, clf),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudwatchLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", clf),
				),
			},

			{
				Config: testAccAWSAPIGatewayStageConfig_accessLogSettings(rName, json),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudwatchLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", json),
				),
			},
			{
				Config: testAccAWSAPIGatewayStageConfig_accessLogSettings(rName, xml),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudwatchLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", xml),
				),
			},
			{
				Config: testAccAWSAPIGatewayStageConfig_accessLogSettings(rName, csv),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudwatchLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", csv),
				),
			},
			{
				Config: testAccAWSAPIGatewayStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayStage_accessLogSettings_kinesis(t *testing.T) {
	var conf apigateway.Stage
	rName := acctest.RandString(5)
	resourceName := "aws_api_gateway_stage.test"
	clf := `$context.identity.sourceIp $context.identity.caller $context.identity.user [$context.requestTime] "$context.httpMethod $context.resourcePath $context.protocol" $context.status $context.responseLength $context.requestId`
	json := `{ "requestId":"$context.requestId", "ip": "$context.identity.sourceIp", "caller":"$context.identity.caller", "user":"$context.identity.user", "requestTime":"$context.requestTime", "httpMethod":"$context.httpMethod", "resourcePath":"$context.resourcePath", "status":"$context.status", "protocol":"$context.protocol", "responseLength":"$context.responseLength" }`
	xml := `<request id="$context.requestId"> <ip>$context.identity.sourceIp</ip> <caller>$context.identity.caller</caller> <user>$context.identity.user</user> <requestTime>$context.requestTime</requestTime> <httpMethod>$context.httpMethod</httpMethod> <resourcePath>$context.resourcePath</resourcePath> <status>$context.status</status> <protocol>$context.protocol</protocol> <responseLength>$context.responseLength</responseLength> </request>`
	csv := `$context.identity.sourceIp,$context.identity.caller,$context.identity.user,$context.requestTime,$context.httpMethod,$context.resourcePath,$context.protocol,$context.status,$context.responseLength,$context.requestId`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayStageConfig_accessLogSettingsKinesis(rName, clf),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					testAccMatchResourceAttrRegionalARN(resourceName, "access_log_settings.0.destination_arn", "firehose", regexp.MustCompile(`deliverystream/amazon-apigateway-.+`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", clf),
				),
			},

			{
				Config: testAccAWSAPIGatewayStageConfig_accessLogSettingsKinesis(rName, json),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					testAccMatchResourceAttrRegionalARN(resourceName, "access_log_settings.0.destination_arn", "firehose", regexp.MustCompile(`deliverystream/amazon-apigateway-.+`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", json),
				),
			},
			{
				Config: testAccAWSAPIGatewayStageConfig_accessLogSettingsKinesis(rName, xml),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					testAccMatchResourceAttrRegionalARN(resourceName, "access_log_settings.0.destination_arn", "firehose", regexp.MustCompile(`deliverystream/amazon-apigateway-.+`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", xml),
				),
			},
			{
				Config: testAccAWSAPIGatewayStageConfig_accessLogSettingsKinesis(rName, csv),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					testAccMatchResourceAttrRegionalARN(resourceName, "access_log_settings.0.destination_arn", "firehose", regexp.MustCompile(`deliverystream/amazon-apigateway-.+`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", csv),
				),
			},
			{
				Config: testAccAWSAPIGatewayStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayStageExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayStageExists(n string, res *apigateway.Stage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Stage ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		req := &apigateway.GetStageInput{
			RestApiId: aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
			StageName: aws.String(rs.Primary.Attributes["stage_name"]),
		}
		out, err := conn.GetStage(req)
		if err != nil {
			return err
		}

		*res = *out

		return nil
	}
}

func testAccCheckAWSAPIGatewayStageDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_stage" {
			continue
		}

		req := &apigateway.GetStageInput{
			RestApiId: aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
			StageName: aws.String(rs.Primary.Attributes["stage_name"]),
		}
		out, err := conn.GetStage(req)
		if err == nil {
			return fmt.Errorf("API Gateway Stage still exists: %s", out)
		}

		awsErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if awsErr.Code() != apigateway.ErrCodeNotFoundException {
			return err
		}

		return nil
	}

	return nil
}

func testAccAWSAPIGatewayStageImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["stage_name"]), nil
	}
}

func testAccAWSAPIGatewayStageConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "tf-acc-test-%s"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.co.uk"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "dev" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "dev"
  description = "This is a dev env"

  variables = {
    "a" = "2"
  }
}
`, rName)
}

func testAccAWSAPIGatewayStageConfigBaseDeploymentStageName(rName string, stageName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
  rest_api_id = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_method" "test" {
  authorization = "NONE"
  http_method   = "GET"
  resource_id   = aws_api_gateway_resource.test.id
  rest_api_id   = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_method_response" "error" {
  http_method = aws_api_gateway_method.test.http_method
  resource_id = aws_api_gateway_resource.test.id
  rest_api_id = aws_api_gateway_rest_api.test.id
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  http_method             = aws_api_gateway_method.test.http_method
  integration_http_method = "GET"
  resource_id             = aws_api_gateway_resource.test.id
  rest_api_id             = aws_api_gateway_rest_api.test.id
  type                    = "HTTP"
  uri                     = "https://www.google.co.uk"
}

resource "aws_api_gateway_integration_response" "test" {
  http_method = aws_api_gateway_integration.test.http_method
  resource_id = aws_api_gateway_resource.test.id
  rest_api_id = aws_api_gateway_rest_api.test.id
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = %[2]q
}
`, rName, stageName)
}

func testAccAWSAPIGatewayStageConfigReferencingDeployment(rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayStageConfigBaseDeploymentStageName(rName, ""),
		`
# Due to oddities with API Gateway, certain environments utilize
# a double deployment configuration. This can cause destroy errors.
resource "aws_api_gateway_stage" "test" {
  deployment_id = aws_api_gateway_deployment.test.id
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "test"

  lifecycle {
    ignore_changes = [deployment_id]
  }
}

resource "aws_api_gateway_deployment" "test2" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_stage.test.stage_name
}
`)
}

func testAccAWSAPIGatewayStageConfig_basic(rName string) string {
	return testAccAWSAPIGatewayStageConfig_base(rName) + `
resource "aws_api_gateway_stage" "test" {
  rest_api_id           = aws_api_gateway_rest_api.test.id
  stage_name            = "prod"
  deployment_id         = aws_api_gateway_deployment.dev.id
  cache_cluster_enabled = true
  cache_cluster_size    = "0.5"
  xray_tracing_enabled  = true
  variables = {
    one = "1"
    two = "2"
  }
  tags = {
    Name = "tf-test"
  }
}
`
}

func testAccAWSAPIGatewayStageConfig_updated(rName string) string {
	return testAccAWSAPIGatewayStageConfig_base(rName) + `
resource "aws_api_gateway_stage" "test" {
  rest_api_id           = aws_api_gateway_rest_api.test.id
  stage_name            = "prod"
  deployment_id         = aws_api_gateway_deployment.dev.id
  cache_cluster_enabled = false
  description           = "Hello world"
  xray_tracing_enabled  = false
  variables = {
    one   = "1"
    three = "3"
  }
  tags = {
    Name      = "tf-test"
    ExtraName = "tf-test"
  }
}
`
}

func testAccAWSAPIGatewayStageConfig_accessLogSettings(rName string, format string) string {
	return testAccAWSAPIGatewayStageConfig_base(rName) + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%s"
}

resource "aws_api_gateway_stage" "test" {
  rest_api_id           = aws_api_gateway_rest_api.test.id
  stage_name            = "prod"
  deployment_id         = aws_api_gateway_deployment.dev.id
  cache_cluster_enabled = true
  cache_cluster_size    = "0.5"
  variables = {
    one = "1"
    two = "2"
  }
  tags = {
    Name = "tf-test"
  }
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.test.arn
    format          = %q
  }
}
`, rName, format)
}

func testAccAWSAPIGatewayStageConfig_accessLogSettingsKinesis(rName string, format string) string {
	return testAccAWSAPIGatewayStageConfig_base(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "%[1]s"
  acl    = "private"
}

resource "aws_iam_role" "test" {
  name = "%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = "amazon-apigateway-%[1]s"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_api_gateway_stage" "test" {
  rest_api_id           = aws_api_gateway_rest_api.test.id
  stage_name            = "prod"
  deployment_id         = aws_api_gateway_deployment.dev.id
  cache_cluster_enabled = true
  cache_cluster_size    = "0.5"
  variables = {
    one = "1"
    two = "2"
  }
  tags = {
    Name = "tf-test"
  }
  access_log_settings {
    destination_arn = aws_kinesis_firehose_delivery_stream.test.arn
    format          = %q
  }
}
`, rName, format)
}
