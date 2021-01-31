package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAPIGatewayUsagePlan_basic(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := acctest.RandomWithPrefix("tf-acc-test")
	updatedName := acctest.RandomWithPrefix("tf-acc-test-2")
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSApiGatewayUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/usageplans/+.`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "api_stages.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSApiGatewayUsagePlanBasicConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayUsagePlan_tags(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSApiGatewayUsagePlanBasicTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSApiGatewayUsagePlanBasicTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSApiGatewayUsagePlanBasicTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayUsagePlan_description(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSApiGatewayUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSApiGatewayUsagePlanDescriptionConfig(rName, "This is a description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a description"),
				),
			},
			{
				Config: testAccAWSApiGatewayUsagePlanDescriptionConfig(rName, "This is a new description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "This is a new description"),
				),
			},
			{
				Config: testAccAWSApiGatewayUsagePlanDescriptionConfig(rName, "This is a description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a description"),
				),
			},
			{
				Config: testAccAWSApiGatewayUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayUsagePlan_productCode(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSApiGatewayUsagePlanProductCodeConfig(rName, "MYCODE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "product_code", "MYCODE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSApiGatewayUsagePlanProductCodeConfig(rName, "MYCODE2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "product_code", "MYCODE2"),
				),
			},
			{
				Config: testAccAWSApiGatewayUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "product_code", ""),
				),
			},
			{
				Config: testAccAWSApiGatewayUsagePlanProductCodeConfig(rName, "MYCODE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "product_code", "MYCODE"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayUsagePlan_throttling(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSApiGatewayUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "throttle_settings"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSApiGatewayUsagePlanThrottlingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.burst_limit", "2"),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.rate_limit", "5"),
				),
			},
			{
				Config: testAccAWSApiGatewayUsagePlanThrottlingModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.burst_limit", "3"),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.rate_limit", "6"),
				),
			},
			{
				Config: testAccAWSApiGatewayUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "throttle_settings"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/2057
func TestAccAWSAPIGatewayUsagePlan_throttlingInitialRateLimit(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSApiGatewayUsagePlanThrottlingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.rate_limit", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayUsagePlan_quota(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSApiGatewayUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "quota_settings"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSApiGatewayUsagePlanQuotaConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.limit", "100"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.offset", "6"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.period", "WEEK"),
				),
			},
			{
				Config: testAccAWSApiGatewayUsagePlanQuotaModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.limit", "200"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.offset", "20"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.period", "MONTH"),
				),
			},
			{
				Config: testAccAWSApiGatewayUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "quota_settings"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayUsagePlan_apiStages(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayUsagePlanDestroy,
		Steps: []resource.TestStep{
			// Create UsagePlan WITH Stages as the API calls are different
			// when creating or updating.
			{
				Config: testAccAWSApiGatewayUsagePlanApiStagesConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "test",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Handle api stages removal
			{
				Config: testAccAWSApiGatewayUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "api_stages"),
				),
			},
			// Handle api stages additions
			{
				Config: testAccAWSApiGatewayUsagePlanApiStagesConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "test",
					}),
				),
			},
			// Handle api stages updates
			{
				Config: testAccAWSApiGatewayUsagePlanApiStagesMultipleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "foo",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "test",
					}),
				),
			},
			{
				Config: testAccAWSApiGatewayUsagePlanApiStagesModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "foo",
					}),
				),
			},
			{
				Config: testAccAWSApiGatewayUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "api_stages"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayUsagePlan_apiStages_multiple(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSApiGatewayUsagePlanApiStagesMultipleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "foo",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						"stage": "test",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayUsagePlan_disappears(t *testing.T) {
	var conf apigateway.UsagePlan
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayUsagePlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSApiGatewayUsagePlanBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayUsagePlanExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayUsagePlan(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayUsagePlanExists(n string, res *apigateway.UsagePlan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Usage Plan ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		req := &apigateway.GetUsagePlanInput{
			UsagePlanId: aws.String(rs.Primary.ID),
		}
		up, err := conn.GetUsagePlan(req)
		if err != nil {
			return err
		}

		if aws.StringValue(up.Id) != rs.Primary.ID {
			return fmt.Errorf("APIGateway Usage Plan not found")
		}

		*res = *up

		return nil
	}
}

func testAccCheckAWSAPIGatewayUsagePlanDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_usage_plan" {
			continue
		}

		req := &apigateway.GetUsagePlanInput{
			UsagePlanId: aws.String(rs.Primary.ID),
		}
		describe, err := conn.GetUsagePlan(req)

		if err == nil {
			if describe.Id != nil && aws.StringValue(describe.Id) == rs.Primary.ID {
				return fmt.Errorf("API Gateway Usage Plan still exists")
			}
		}

		if !isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
			return err
		}

		return nil
	}

	return nil
}

func testAccAWSAPIGatewayUsagePlanConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
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
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "test"
  description = "This is a test"

  variables = {
    "a" = "2"
  }
}

resource "aws_api_gateway_deployment" "foo" {
  depends_on = [
    aws_api_gateway_deployment.test,
    aws_api_gateway_integration.test,
  ]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "foo"
  description = "This is a prod stage"
}
`, rName)
}

func testAccAWSApiGatewayUsagePlanBasicConfig(rName string) string {
	return testAccAWSAPIGatewayUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"
}
`, rName)
}

func testAccAWSApiGatewayUsagePlanBasicTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSAPIGatewayUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  tags = {
    %q = %q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSApiGatewayUsagePlanBasicTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSAPIGatewayUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSApiGatewayUsagePlanDescriptionConfig(rName, desc string) string {
	return testAccAWSAPIGatewayUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, desc)
}

func testAccAWSApiGatewayUsagePlanProductCodeConfig(rName, code string) string {
	return testAccAWSAPIGatewayUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name         = %[1]q
  product_code = %[2]q
}
`, rName, code)
}

func testAccAWSApiGatewayUsagePlanThrottlingConfig(rName string) string {
	return testAccAWSAPIGatewayUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  throttle_settings {
    burst_limit = 2
    rate_limit  = 5
  }
}
`, rName)
}

func testAccAWSApiGatewayUsagePlanThrottlingModifiedConfig(rName string) string {
	return testAccAWSAPIGatewayUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  throttle_settings {
    burst_limit = 3
    rate_limit  = 6
  }
}
`, rName)
}

func testAccAWSApiGatewayUsagePlanQuotaConfig(rName string) string {
	return testAccAWSAPIGatewayUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  quota_settings {
    limit  = 100
    offset = 6
    period = "WEEK"
  }
}
`, rName)
}

func testAccAWSApiGatewayUsagePlanQuotaModifiedConfig(rName string) string {
	return testAccAWSAPIGatewayUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  quota_settings {
    limit  = 200
    offset = 20
    period = "MONTH"
  }
}
`, rName)
}

func testAccAWSApiGatewayUsagePlanApiStagesConfig(rName string) string {
	return testAccAWSAPIGatewayUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.test.stage_name
  }
}
`, rName)
}

func testAccAWSApiGatewayUsagePlanApiStagesModifiedConfig(rName string) string {
	return testAccAWSAPIGatewayUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.foo.stage_name
  }
}
`, rName)
}

func testAccAWSApiGatewayUsagePlanApiStagesMultipleConfig(rName string) string {
	return testAccAWSAPIGatewayUsagePlanConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = "%s"

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.foo.stage_name
  }

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.test.stage_name
  }
}
`, rName)
}
