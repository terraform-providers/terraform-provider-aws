package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_api_gateway_rest_api", &resource.Sweeper{
		Name: "aws_api_gateway_rest_api",
		F:    testSweepAPIGatewayRestApis,
	})
}

func testSweepAPIGatewayRestApis(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).apigatewayconn

	err = conn.GetRestApisPages(&apigateway.GetRestApisInput{}, func(page *apigateway.GetRestApisOutput, lastPage bool) bool {
		for _, item := range page.Items {
			input := &apigateway.DeleteRestApiInput{
				RestApiId: item.Id,
			}
			log.Printf("[INFO] Deleting API Gateway REST API: %s", input)
			// TooManyRequestsException: Too Many Requests can take over a minute to resolve itself
			err := resource.Retry(2*time.Minute, func() *resource.RetryError {
				_, err := conn.DeleteRestApi(input)
				if err != nil {
					if isAWSErr(err, apigateway.ErrCodeTooManyRequestsException, "") {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete API Gateway REST API %s: %s", *item.Name, err)
				continue
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway REST API sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving API Gateway REST APIs: %s", err)
	}

	return nil
}

func TestAccAWSAPIGatewayRestApi_basic(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/+.`)),
					testAccCheckAWSAPIGatewayRestAPINameAttribute(&conf, rName),
					testAccCheckAWSAPIGatewayRestAPIMinimumCompressionSizeAttribute(&conf, 0),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "binary_media_types"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			{
				Config: testAccAWSAPIGatewayRestAPIUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayRestAPINameAttribute(&conf, rName),
					testAccCheckAWSAPIGatewayRestAPIDescriptionAttribute(&conf, "test"),
					testAccCheckAWSAPIGatewayRestAPIMinimumCompressionSizeAttribute(&conf, 10485760),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "10485760"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet-stream"),
				),
			},

			{
				Config: testAccAWSAPIGatewayRestAPIDisableCompressionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayRestAPIMinimumCompressionSizeAttributeIsNil(&conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "-1"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_tags(t *testing.T) {
	var conf apigateway.RestApi
	resourceName := "aws_api_gateway_rest_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/+.`)),
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
				Config: testAccAWSAPIGatewayRestAPIConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/+.`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},

			{
				Config: testAccAWSAPIGatewayRestAPIConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/+.`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_disappears(t *testing.T) {
	var restApi apigateway.RestApi
	resourceName := "aws_api_gateway_rest_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &restApi),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayRestApi(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_EndpointConfiguration(t *testing.T) {
	var restApi apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfig_EndpointConfiguration(rName, "REGIONAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "REGIONAL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// For backwards compatibility, test removing endpoint_configuration, which should do nothing
			{
				Config: testAccAWSAPIGatewayRestAPIConfig_Name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "REGIONAL"),
				),
			},
			// Test updating endpoint type
			{
				PreConfig: func() {
					// Ensure region supports EDGE endpoint
					// This can eventually be moved to a PreCheck function
					// If the region does not support EDGE endpoint type, this test will either show
					// SKIP (if REGIONAL passed) or FAIL (if REGIONAL failed)
					conn := testAccProvider.Meta().(*AWSClient).apigatewayconn
					output, err := conn.CreateRestApi(&apigateway.CreateRestApiInput{
						Name: aws.String(acctest.RandomWithPrefix("tf-acc-test-edge-endpoint-precheck")),
						EndpointConfiguration: &apigateway.EndpointConfiguration{
							Types: []*string{aws.String("EDGE")},
						},
					})
					if err != nil {
						if isAWSErr(err, apigateway.ErrCodeBadRequestException, "Endpoint Configuration type EDGE is not supported in this region") {
							t.Skip("Region does not support EDGE endpoint type")
						}
						t.Fatal(err)
					}

					// Be kind and rewind. :)
					_, err = conn.DeleteRestApi(&apigateway.DeleteRestApiInput{
						RestApiId: output.Id,
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccAWSAPIGatewayRestAPIConfig_EndpointConfiguration(rName, "EDGE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "EDGE"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_EndpointConfiguration_Private(t *testing.T) {
	var restApi apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Ensure region supports PRIVATE endpoint
					// This can eventually be moved to a PreCheck function
					conn := testAccProvider.Meta().(*AWSClient).apigatewayconn
					output, err := conn.CreateRestApi(&apigateway.CreateRestApiInput{
						Name: aws.String(acctest.RandomWithPrefix("tf-acc-test-private-endpoint-precheck")),
						EndpointConfiguration: &apigateway.EndpointConfiguration{
							Types: []*string{aws.String("PRIVATE")},
						},
					})
					if err != nil {
						if isAWSErr(err, apigateway.ErrCodeBadRequestException, "Endpoint Configuration type PRIVATE is not supported in this region") {
							t.Skip("Region does not support PRIVATE endpoint type")
						}
						t.Fatal(err)
					}

					// Be kind and rewind. :)
					_, err = conn.DeleteRestApi(&apigateway.DeleteRestApiInput{
						RestApiId: output.Id,
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccAWSAPIGatewayRestAPIConfig_EndpointConfiguration(rName, "PRIVATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
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

func TestAccAWSAPIGatewayRestApi_EndpointConfiguration_VPCEndpoint(t *testing.T) {
	var restApi apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Ensure region supports PRIVATE endpoint
					// This can eventually be moved to a PreCheck function
					conn := testAccProvider.Meta().(*AWSClient).apigatewayconn
					output, err := conn.CreateRestApi(&apigateway.CreateRestApiInput{
						Name: aws.String(acctest.RandomWithPrefix("tf-acc-test-private-endpoint-precheck")),
						EndpointConfiguration: &apigateway.EndpointConfiguration{
							Types: []*string{aws.String("PRIVATE")},
						},
					})
					if err != nil {
						if isAWSErr(err, apigateway.ErrCodeBadRequestException, "Endpoint Configuration type PRIVATE is not supported in this region") {
							t.Skip("Region does not support PRIVATE endpoint type")
						}
						t.Fatal(err)
					}

					// Be kind and rewind. :)
					_, err = conn.DeleteRestApi(&apigateway.DeleteRestApiInput{
						RestApiId: output.Id,
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccAWSAPIGatewayRestAPIConfig_VPCEndpointConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfig_VPCEndpointConfiguration2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "2"),
				),
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfig_VPCEndpointConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_api_key_source(t *testing.T) {
	expectedAPIKeySource := "HEADER"
	expectedUpdateAPIKeySource := "AUTHORIZER"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigWithAPIKeySource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key_source", expectedAPIKeySource),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigWithUpdateAPIKeySource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key_source", expectedUpdateAPIKeySource),
				),
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key_source", expectedAPIKeySource),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_policy(t *testing.T) {
	resourceName := "aws_api_gateway_rest_api.test"
	expectedPolicyText := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":"*"},"Action":"execute-api:Invoke","Resource":"*","Condition":{"IpAddress":{"aws:SourceIp":"123.123.123.123/32"}}}]}`
	expectedUpdatePolicyText := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Principal":{"AWS":"*"},"Action":"execute-api:Invoke","Resource":"*"}]}`
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigWithPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "policy", expectedPolicyText),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigUpdatePolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "policy", expectedUpdatePolicyText),
				),
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_openapi(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigOpenAPI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayRestAPINameAttribute(&conf, rName),
					testAccCheckAWSAPIGatewayRestAPIRoutes(&conf, []string{"/", "/test"}),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "binary_media_types"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAWSAPIGatewayRestAPIUpdateConfigOpenAPI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayRestAPINameAttribute(&conf, rName),
					testAccCheckAWSAPIGatewayRestAPIRoutes(&conf, []string{"/", "/update"}),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayRestAPINameAttribute(conf *apigateway.RestApi, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *conf.Name != name {
			return fmt.Errorf("Wrong Name: %q instead of %s", *conf.Name, name)
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayRestAPIDescriptionAttribute(conf *apigateway.RestApi, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *conf.Description != description {
			return fmt.Errorf("Wrong Description: %q", *conf.Description)
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayRestAPIMinimumCompressionSizeAttribute(conf *apigateway.RestApi, minimumCompressionSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.MinimumCompressionSize == nil {
			return fmt.Errorf("MinimumCompressionSize should not be nil")
		}
		if *conf.MinimumCompressionSize != minimumCompressionSize {
			return fmt.Errorf("Wrong MinimumCompressionSize: %d", *conf.MinimumCompressionSize)
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayRestAPIMinimumCompressionSizeAttributeIsNil(conf *apigateway.RestApi) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.MinimumCompressionSize != nil {
			return fmt.Errorf("MinimumCompressionSize should be nil: %d", *conf.MinimumCompressionSize)
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayRestAPIRoutes(conf *apigateway.RestApi, routes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		resp, err := conn.GetResources(&apigateway.GetResourcesInput{
			RestApiId: conf.Id,
		})
		if err != nil {
			return err
		}

		actualRoutePaths := map[string]bool{}
		for _, resource := range resp.Items {
			actualRoutePaths[*resource.Path] = true
		}

		for _, route := range routes {
			if _, ok := actualRoutePaths[route]; !ok {
				return fmt.Errorf("Expected path %v but did not find it in %v", route, actualRoutePaths)
			}
			delete(actualRoutePaths, route)
		}

		if len(actualRoutePaths) > 0 {
			return fmt.Errorf("Found unexpected paths %v", actualRoutePaths)
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayRestAPIExists(n string, res *apigateway.RestApi) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		req := &apigateway.GetRestApiInput{
			RestApiId: aws.String(rs.Primary.ID),
		}
		describe, err := conn.GetRestApi(req)
		if err != nil {
			return err
		}

		if *describe.Id != rs.Primary.ID {
			return fmt.Errorf("APIGateway not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayRestAPIDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_rest_api" {
			continue
		}

		req := &apigateway.GetRestApisInput{}
		describe, err := conn.GetRestApis(req)

		if err == nil {
			if len(describe.Items) != 0 &&
				*describe.Items[0].Id == rs.Primary.ID {
				return fmt.Errorf("API Gateway still exists")
			}
		}

		return err
	}

	return nil
}

// testAccAPIGatewayTypeEDGEPreCheck checks if endpoint config type EDGE can be used in a test and skips test if not (i.e., not in standard partition).
func testAccAPIGatewayTypeEDGEPreCheck(t *testing.T) {
	if testAccGetPartition() != endpoints.AwsPartitionID {
		t.Skipf("skipping test; Endpoint Configuration type EDGE is not supported in this partition (%s)", testAccGetPartition())
	}
}

func testAccAWSAPIGatewayRestAPIConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name                     = "%s"
  minimum_compression_size = 0
}
`, rName)
}

func testAccAWSAPIGatewayRestAPIConfig_EndpointConfiguration(rName, endpointType string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"

  endpoint_configuration {
    types = ["%s"]
  }
}
`, rName, endpointType)
}

func testAccAWSAPIGatewayRestAPIConfig_Name(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, rName)
}

func testAccAWSAPIGatewayRestAPIConfig_VPCEndpointConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "11.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

data "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = "default"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = aws_vpc.test.cidr_block
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false

  subnet_ids = [
    aws_subnet.test.id,
  ]

  security_group_ids = [
    data.aws_security_group.test.id,
  ]
}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  endpoint_configuration {
    types            = ["PRIVATE"]
    vpc_endpoint_ids = [aws_vpc_endpoint.test.id]
  }
}
`, rName)
}

func testAccAWSAPIGatewayRestAPIConfig_VPCEndpointConfiguration2(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "11.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

data "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = "default"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = aws_vpc.test.cidr_block
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false

  subnet_ids = [
    aws_subnet.test.id,
  ]

  security_group_ids = [
    data.aws_security_group.test.id,
  ]
}

resource "aws_vpc_endpoint" "test2" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false

  subnet_ids = [
    aws_subnet.test.id,
  ]

  security_group_ids = [
    data.aws_security_group.test.id,
  ]
}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  endpoint_configuration {
    types            = ["PRIVATE"]
    vpc_endpoint_ids = [aws_vpc_endpoint.test.id, aws_vpc_endpoint.test2.id]
  }
}
`, rName)
}

func testAccAWSAPIGatewayRestAPIConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"

  tags = {
    %q = %q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSAPIGatewayRestAPIConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSAPIGatewayRestAPIConfigWithAPIKeySource(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name           = "%s"
  api_key_source = "HEADER"
}
`, rName)
}

func testAccAWSAPIGatewayRestAPIConfigWithUpdateAPIKeySource(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name           = "%s"
  api_key_source = "AUTHORIZER"
}
`, rName)
}

func testAccAWSAPIGatewayRestAPIConfigWithPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name                     = "%s"
  minimum_compression_size = 0
  policy                   = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "execute-api:Invoke",
      "Resource": "*",
      "Condition": {
        "IpAddress": {
          "aws:SourceIp": "123.123.123.123/32"
        }
      }
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSAPIGatewayRestAPIConfigUpdatePolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name                     = "%s"
  minimum_compression_size = 0
  policy                   = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Deny",
            "Principal": {
                "AWS": "*"
            },
            "Action": "execute-api:Invoke",
            "Resource": "*"
        }
    ]
}
EOF
}
`, rName)
}

func testAccAWSAPIGatewayRestAPIUpdateConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name                     = "%s"
  description              = "test"
  binary_media_types       = ["application/octet-stream"]
  minimum_compression_size = 10485760
}
`, rName)
}

func testAccAWSAPIGatewayRestAPIDisableCompressionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name                     = "%s"
  description              = "test"
  binary_media_types       = ["application/octet-stream"]
  minimum_compression_size = -1
}
`, rName)
}

func testAccAWSAPIGatewayRestAPIConfigOpenAPI(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
  body = <<EOF
{
  "swagger": "2.0",
  "info": {
    "title": "%s",
    "version": "2017-04-20T04:08:08Z"
  },
  "schemes": [
    "https"
  ],
  "paths": {
    "/test": {
      "get": {
        "responses": {
          "200": {
            "description": "200 response"
          }
        },
        "x-amazon-apigateway-integration": {
          "type": "HTTP",
          "uri": "https://www.google.de",
          "httpMethod": "GET",
          "responses": {
            "default": {
              "statusCode": 200
            }
          }
        }
      }
    }
  }
}
EOF
}
`, rName, rName)
}

func testAccAWSAPIGatewayRestAPIUpdateConfigOpenAPI(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
  body = <<EOF
{
  "swagger": "2.0",
  "info": {
    "title": "%s",
    "version": "2017-04-20T04:08:08Z"
  },
  "schemes": [
    "https"
  ],
  "paths": {
    "/update": {
      "get": {
        "responses": {
          "200": {
            "description": "200 response"
          }
        },
        "x-amazon-apigateway-integration": {
          "type": "HTTP",
          "uri": "https://www.google.de",
          "httpMethod": "GET",
          "responses": {
            "default": {
              "statusCode": 200
            }
          }
        }
      }
    }
  }
}
EOF
}
`, rName, rName)
}
