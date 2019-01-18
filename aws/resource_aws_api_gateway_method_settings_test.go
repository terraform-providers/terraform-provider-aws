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

func TestAccAWSAPIGatewayMethodSettings_basic(t *testing.T) {
	var stage apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsLoggingLevel(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.logging_level", "INFO"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayMethodSettings_Settings_CacheDataEncrypted(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsCacheDataEncrypted(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_data_encrypted", "true"),
				),
			},
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsCacheDataEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_data_encrypted", "false"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayMethodSettings_Settings_CacheTtlInSeconds(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsCacheTtlInSeconds(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_ttl_in_seconds", "1"),
				),
			},
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsCacheTtlInSeconds(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_ttl_in_seconds", "2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayMethodSettings_Settings_CachingEnabled(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsCachingEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.caching_enabled", "true"),
				),
			},
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsCachingEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.caching_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayMethodSettings_Settings_DataTraceEnabled(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsDataTraceEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.data_trace_enabled", "true"),
				),
			},
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsDataTraceEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.data_trace_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayMethodSettings_Settings_LoggingLevel(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsLoggingLevel(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage1),
					testAccCheckAWSAPIGatewayMethodSettings_loggingLevel(&stage1, "test/GET", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.logging_level", "INFO"),
				),
			},
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsLoggingLevel(rName, "OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage2),
					testAccCheckAWSAPIGatewayMethodSettings_loggingLevel(&stage2, "test/GET", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.logging_level", "OFF"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayMethodSettings_Settings_MetricsEnabled(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsMetricsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage1),
					testAccCheckAWSAPIGatewayMethodSettings_metricsEnabled(&stage1, "test/GET", true),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.metrics_enabled", "true"),
				),
			},
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsMetricsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage2),
					testAccCheckAWSAPIGatewayMethodSettings_metricsEnabled(&stage2, "test/GET", false),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.metrics_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayMethodSettings_Settings_Multiple(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsMultiple(rName, "INFO", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage1),
					testAccCheckAWSAPIGatewayMethodSettings_metricsEnabled(&stage1, "test/GET", true),
					testAccCheckAWSAPIGatewayMethodSettings_loggingLevel(&stage1, "test/GET", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.metrics_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.logging_level", "INFO"),
				),
			},
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsMultiple(rName, "OFF", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage2),
					testAccCheckAWSAPIGatewayMethodSettings_metricsEnabled(&stage2, "test/GET", false),
					testAccCheckAWSAPIGatewayMethodSettings_loggingLevel(&stage2, "test/GET", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.logging_level", "OFF"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayMethodSettings_Settings_RequireAuthorizationForCacheControl(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsRequireAuthorizationForCacheControl(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.require_authorization_for_cache_control", "true"),
				),
			},
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsRequireAuthorizationForCacheControl(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.require_authorization_for_cache_control", "false"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayMethodSettings_Settings_ThrottlingBurstLimit(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsThrottlingBurstLimit(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_burst_limit", "1"),
				),
			},
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsThrottlingBurstLimit(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_burst_limit", "2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayMethodSettings_Settings_ThrottlingRateLimit(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsThrottlingRateLimit(rName, 1.1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_rate_limit", "1.1"),
				),
			},
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsThrottlingRateLimit(rName, 2.2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_rate_limit", "2.2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayMethodSettings_Settings_UnauthorizedCacheControlHeaderStrategy(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsUnauthorizedCacheControlHeaderStrategy(rName, "SUCCEED_WITH_RESPONSE_HEADER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.unauthorized_cache_control_header_strategy", "SUCCEED_WITH_RESPONSE_HEADER"),
				),
			},
			{
				Config: testAccAWSAPIGatewayMethodSettingsConfigSettingsUnauthorizedCacheControlHeaderStrategy(rName, "SUCCEED_WITHOUT_RESPONSE_HEADER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayMethodSettingsExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.unauthorized_cache_control_header_strategy", "SUCCEED_WITHOUT_RESPONSE_HEADER"),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayMethodSettings_metricsEnabled(conf *apigateway.Stage, path string, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		settings, ok := conf.MethodSettings[path]
		if !ok {
			return fmt.Errorf("Expected to find method settings for %q", path)
		}

		if expected && *settings.MetricsEnabled != expected {
			return fmt.Errorf("Expected metrics to be enabled, got %t", *settings.MetricsEnabled)
		}
		if !expected && *settings.MetricsEnabled != expected {
			return fmt.Errorf("Expected metrics to be disabled, got %t", *settings.MetricsEnabled)
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayMethodSettings_loggingLevel(conf *apigateway.Stage, path string, expectedLevel string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		settings, ok := conf.MethodSettings[path]
		if !ok {
			return fmt.Errorf("Expected to find method settings for %q", path)
		}

		if *settings.LoggingLevel != expectedLevel {
			return fmt.Errorf("Expected logging level to match %q, got %q", expectedLevel, *settings.LoggingLevel)
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayMethodSettingsExists(n string, res *apigateway.Stage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Stage ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigateway

		req := &apigateway.GetStageInput{
			StageName: aws.String(rs.Primary.Attributes["stage_name"]),
			RestApiId: aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		out, err := conn.GetStage(req)
		if err != nil {
			return err
		}

		*res = *out

		return nil
	}
}

func testAccCheckAWSAPIGatewayMethodSettingsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_method_settings" {
			continue
		}

		req := &apigateway.GetStageInput{
			StageName: aws.String(rs.Primary.Attributes["stage_name"]),
			RestApiId: aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		out, err := conn.GetStage(req)
		if err == nil {
			return fmt.Errorf("API Gateway Stage still exists: %s", out)
		}

		awsErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if awsErr.Code() != "NotFoundException" {
			return err
		}
	}

	return nil
}

func testAccAWSAPIGatewayMethodSettingsConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %q
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

  request_parameters = {
    "method.request.header.Content-Type" = false,
	  "method.request.querystring.page" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"
  type        = "MOCK"

  request_templates = {
    "application/xml" = <<EOF
{
   "body" : $input.json('$')
}
EOF
  }
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = ["aws_api_gateway_integration.test"]
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name = "dev"
}`, rName)
}

func testAccAWSAPIGatewayMethodSettingsConfigSettingsCacheDataEncrypted(rName string, cacheDataEncrypted bool) string {
	return testAccAWSAPIGatewayMethodSettingsConfigBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "${aws_api_gateway_deployment.test.stage_name}"

  settings {
    cache_data_encrypted = %t
  }
}
`, cacheDataEncrypted)
}

func testAccAWSAPIGatewayMethodSettingsConfigSettingsCacheTtlInSeconds(rName string, cacheTtlInSeconds int) string {
	return testAccAWSAPIGatewayMethodSettingsConfigBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "${aws_api_gateway_deployment.test.stage_name}"

  settings {
    cache_ttl_in_seconds = %d
  }
}
`, cacheTtlInSeconds)
}

func testAccAWSAPIGatewayMethodSettingsConfigSettingsCachingEnabled(rName string, cachingEnabled bool) string {
	return testAccAWSAPIGatewayMethodSettingsConfigBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "${aws_api_gateway_deployment.test.stage_name}"

  settings {
    caching_enabled = %t
  }
}
`, cachingEnabled)
}

func testAccAWSAPIGatewayMethodSettingsConfigSettingsDataTraceEnabled(rName string, dataTraceEnabled bool) string {
	return testAccAWSAPIGatewayMethodSettingsConfigBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "${aws_api_gateway_deployment.test.stage_name}"

  settings {
    data_trace_enabled = %t
  }
}
`, dataTraceEnabled)
}

func testAccAWSAPIGatewayMethodSettingsConfigSettingsLoggingLevel(rName, loggingLevel string) string {
	return testAccAWSAPIGatewayMethodSettingsConfigBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "${aws_api_gateway_deployment.test.stage_name}"

  settings {
    logging_level = %q
  }
}
`, loggingLevel)
}

func testAccAWSAPIGatewayMethodSettingsConfigSettingsMetricsEnabled(rName string, metricsEnabled bool) string {
	return testAccAWSAPIGatewayMethodSettingsConfigBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "${aws_api_gateway_deployment.test.stage_name}"

  settings {
    metrics_enabled = %t
  }
}
`, metricsEnabled)
}

func testAccAWSAPIGatewayMethodSettingsConfigSettingsMultiple(rName, loggingLevel string, metricsEnabled bool) string {
	return testAccAWSAPIGatewayMethodSettingsConfigBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "${aws_api_gateway_deployment.test.stage_name}"
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"

  settings {
    logging_level   = %q
    metrics_enabled = %t
  }
}
`, loggingLevel, metricsEnabled)
}

func testAccAWSAPIGatewayMethodSettingsConfigSettingsRequireAuthorizationForCacheControl(rName string, requireAuthorizationForCacheControl bool) string {
	return testAccAWSAPIGatewayMethodSettingsConfigBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "${aws_api_gateway_deployment.test.stage_name}"

  settings {
    require_authorization_for_cache_control = %t
  }
}
`, requireAuthorizationForCacheControl)
}

func testAccAWSAPIGatewayMethodSettingsConfigSettingsThrottlingBurstLimit(rName string, throttlingBurstLimit int) string {
	return testAccAWSAPIGatewayMethodSettingsConfigBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "${aws_api_gateway_deployment.test.stage_name}"

  settings {
    throttling_burst_limit = %d
  }
}
`, throttlingBurstLimit)
}

func testAccAWSAPIGatewayMethodSettingsConfigSettingsThrottlingRateLimit(rName string, throttlingRateLimit float32) string {
	return testAccAWSAPIGatewayMethodSettingsConfigBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "${aws_api_gateway_deployment.test.stage_name}"

  settings {
    throttling_rate_limit = %f
  }
}
`, throttlingRateLimit)
}

func testAccAWSAPIGatewayMethodSettingsConfigSettingsUnauthorizedCacheControlHeaderStrategy(rName, unauthorizedCacheControlHeaderStrategy string) string {
	return testAccAWSAPIGatewayMethodSettingsConfigBase(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "${aws_api_gateway_deployment.test.stage_name}"

  settings {
    unauthorized_cache_control_header_strategy = %q
  }
}
`, unauthorizedCacheControlHeaderStrategy)
}
