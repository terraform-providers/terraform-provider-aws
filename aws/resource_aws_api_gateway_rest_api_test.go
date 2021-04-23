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
				Config: testAccAWSAPIGatewayRestAPIConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/+.`)),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "body"),
					testAccCheckResourceAttrRfc3339(resourceName, "created_date"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "-1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "root_resource_id", regexp.MustCompile(`[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
				Config: testAccAWSAPIGatewayRestAPIConfigName(rName),
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

func TestAccAWSAPIGatewayRestApi_ApiKeySource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigApiKeySource(rName, "AUTHORIZER"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "AUTHORIZER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigApiKeySource(rName, "HEADER"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
				),
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_ApiKeySource_OverrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigApiKeySourceOverrideBody(rName, "AUTHORIZER", "HEADER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "AUTHORIZER"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated API key source still overrides
			{
				Config: testAccAWSAPIGatewayRestAPIConfigApiKeySourceOverrideBody(rName, "HEADER", "HEADER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
				),
			},
			// Verify updated body API key source is still overridden
			{
				Config: testAccAWSAPIGatewayRestAPIConfigApiKeySourceOverrideBody(rName, "HEADER", "AUTHORIZER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_ApiKeySource_SetByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigApiKeySourceSetByBody(rName, "AUTHORIZER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "AUTHORIZER"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_BinaryMediaTypes(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigBinaryMediaTypes1(rName, "application/octet-stream"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet-stream"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigBinaryMediaTypes1(rName, "application/octet"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_BinaryMediaTypes_OverrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigBinaryMediaTypes1OverrideBody(rName, "application/octet-stream", "image/jpeg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet-stream"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated minimum compression size still overrides
			{
				Config: testAccAWSAPIGatewayRestAPIConfigBinaryMediaTypes1OverrideBody(rName, "application/octet", "image/jpeg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet"),
				),
			},
			// Verify updated body minimum compression size is still overridden
			{
				Config: testAccAWSAPIGatewayRestAPIConfigBinaryMediaTypes1OverrideBody(rName, "application/octet", "image/png"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_BinaryMediaTypes_SetByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigBinaryMediaTypes1SetByBody(rName, "application/octet-stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet-stream"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_Body(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			// The body is expected to only set a title (name) and one route
			{
				Config: testAccAWSAPIGatewayRestAPIConfigBody(rName, "/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
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
				Config: testAccAWSAPIGatewayRestAPIConfigBody(rName, "/update"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayRestAPIRoutes(&conf, []string{"/", "/update"}),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_Description(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDescription(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDescription(rName, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_Description_OverrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDescriptionOverrideBody(rName, "tfdescription1", "oasdescription1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "tfdescription1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated description still overrides
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDescriptionOverrideBody(rName, "tfdescription2", "oasdescription1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "tfdescription2"),
				),
			},
			// Verify updated body description is still overridden
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDescriptionOverrideBody(rName, "tfdescription2", "oasdescription2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "tfdescription2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_Description_SetByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDescriptionSetByBody(rName, "oasdescription1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "oasdescription1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_DisableExecuteApiEndpoint(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDisableExecuteApiEndpoint(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", `false`),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDisableExecuteApiEndpoint(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", `true`),
				),
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDisableExecuteApiEndpoint(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", `false`),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_DisableExecuteApiEndpoint_OverrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDisableExecuteApiEndpointOverrideBody(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify override can be unset (only for body set to false)
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDisableExecuteApiEndpointOverrideBody(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
				),
			},
			// Verify override can be reset
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDisableExecuteApiEndpointOverrideBody(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_DisableExecuteApiEndpoint_SetByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigDisableExecuteApiEndpointSetByBody(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_EndpointConfiguration_VpcEndpointIds(t *testing.T) {
	var restApi apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"
	vpcEndpointResourceName1 := "aws_vpc_endpoint.test"
	vpcEndpointResourceName2 := "aws_vpc_endpoint.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigEndpointConfigurationVpcEndpointIds1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigEndpointConfigurationVpcEndpointIds2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName2, "id"),
				),
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigEndpointConfigurationVpcEndpointIds1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, "id"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_EndpointConfiguration_VpcEndpointIds_OverrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"
	vpcEndpointResourceName1 := "aws_vpc_endpoint.test.0"
	vpcEndpointResourceName2 := "aws_vpc_endpoint.test.1"
	vpcEndpointResourceName3 := "aws_vpc_endpoint.test.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigEndpointConfigurationVpcEndpointIdsOverrideBody(rName, vpcEndpointResourceName1, vpcEndpointResourceName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated configuration value still overrides
			{
				Config: testAccAWSAPIGatewayRestAPIConfigEndpointConfigurationVpcEndpointIdsOverrideBody(rName, vpcEndpointResourceName3, vpcEndpointResourceName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName3, "id"),
				),
			},
			// Verify updated body value is still overridden
			{
				Config: testAccAWSAPIGatewayRestAPIConfigEndpointConfigurationVpcEndpointIdsOverrideBody(rName, vpcEndpointResourceName3, vpcEndpointResourceName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName3, "id"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_EndpointConfiguration_VpcEndpointIds_SetByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigEndpointConfigurationVpcEndpointIdsSetByBody(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_MinimumCompressionSize(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigMinimumCompressionSize(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigMinimumCompressionSize(rName, -1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "-1"),
				),
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigMinimumCompressionSize(rName, 5242880),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "5242880"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_MinimumCompressionSize_OverrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigMinimumCompressionSizeOverrideBody(rName, 1, 5242880),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated minimum compression size still overrides
			{
				Config: testAccAWSAPIGatewayRestAPIConfigMinimumCompressionSizeOverrideBody(rName, 2, 5242880),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "2"),
				),
			},
			// Verify updated body minimum compression size is still overridden
			{
				Config: testAccAWSAPIGatewayRestAPIConfigMinimumCompressionSizeOverrideBody(rName, 2, 1048576),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_MinimumCompressionSize_SetByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigMinimumCompressionSizeSetByBody(rName, 1048576),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "1048576"),
				),
				// TODO: The attribute type must be changed to NullableTypeInt so it can be Computed properly.
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_Name_OverrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigNameOverrideBody(rName, "title1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated name still overrides
			{
				Config: testAccAWSAPIGatewayRestAPIConfigNameOverrideBody(rName2, "title1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
			// Verify updated title still overrides
			{
				Config: testAccAWSAPIGatewayRestAPIConfigNameOverrideBody(rName2, "title2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_Parameters(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigParameters1(rName, "basepath", "prepend"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayRestAPIRoutes(&conf, []string{"/", "/foo", "/foo/bar", "/foo/bar/baz", "/foo/bar/baz/test"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "parameters"},
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigParameters1(rName, "basepath", "ignore"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayRestAPIRoutes(&conf, []string{"/", "/test"}),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_Policy(t *testing.T) {
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
		},
	})
}

func TestAccAWSAPIGatewayRestApi_Policy_OverrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigPolicyOverrideBody(rName, "/test", "Allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayRestAPIRoutes(&conf, []string{"/", "/test"}),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"Allow"`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated body still has override policy
			{
				Config: testAccAWSAPIGatewayRestAPIConfigPolicyOverrideBody(rName, "/test2", "Allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayRestAPIRoutes(&conf, []string{"/", "/test2"}),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"Allow"`)),
				),
			},
			// Verify updated policy still overrides body
			{
				Config: testAccAWSAPIGatewayRestAPIConfigPolicyOverrideBody(rName, "/test2", "Deny"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayRestAPIRoutes(&conf, []string{"/", "/test2"}),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"Deny"`)),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_Policy_SetByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigPolicySetByBody(rName, "Allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists(resourceName, &conf),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"Allow"`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
		},
	})
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

func testAccAWSAPIGatewayRestAPIConfigDisableExecuteApiEndpoint(rName string, disableExecuteApiEndpoint bool) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  disable_execute_api_endpoint = %[2]t
  name                         = %[1]q
}
`, rName, disableExecuteApiEndpoint)
}

func testAccAWSAPIGatewayRestAPIConfigDisableExecuteApiEndpointOverrideBody(rName string, configDisableExecuteApiEndpoint bool, bodyDisableExecuteApiEndpoint bool) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  disable_execute_api_endpoint = %[2]t
  name                         = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
    x-amazon-apigateway-endpoint-configuration = {
      disableExecuteApiEndpoint = %[3]t
    }
  })
}
`, rName, configDisableExecuteApiEndpoint, bodyDisableExecuteApiEndpoint)
}

func testAccAWSAPIGatewayRestAPIConfigDisableExecuteApiEndpointSetByBody(rName string, bodyDisableExecuteApiEndpoint bool) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
    x-amazon-apigateway-endpoint-configuration = {
      disableExecuteApiEndpoint = %[2]t
    }
  })
}
`, rName, bodyDisableExecuteApiEndpoint)
}

func testAccAWSAPIGatewayRestAPIConfig_Name(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, rName)
}

func testAccAWSAPIGatewayRestAPIConfigEndpointConfigurationVpcEndpointIds1(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  endpoint_configuration {
    types            = ["PRIVATE"]
    vpc_endpoint_ids = [aws_vpc_endpoint.test.id]
  }
}
`, rName))
}

func testAccAWSAPIGatewayRestAPIConfigEndpointConfigurationVpcEndpointIds2(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

resource "aws_vpc_endpoint" "test2" {
  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  endpoint_configuration {
    types            = ["PRIVATE"]
    vpc_endpoint_ids = [aws_vpc_endpoint.test.id, aws_vpc_endpoint.test2.id]
  }
}
`, rName))
}

func testAccAWSAPIGatewayRestAPIConfigEndpointConfigurationVpcEndpointIdsOverrideBody(rName string, configVpcEndpointResourceName string, bodyVpcEndpointResourceName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  count = 3

  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  endpoint_configuration {
    types            = ["PRIVATE"]
    vpc_endpoint_ids = [%[2]s]
  }

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
    x-amazon-apigateway-endpoint-configuration = {
      vpcEndpointIds = [%[3]s]
    }
  })
}
`, rName, configVpcEndpointResourceName+".id", bodyVpcEndpointResourceName+".id"))
}

func testAccAWSAPIGatewayRestAPIConfigEndpointConfigurationVpcEndpointIdsSetByBody(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  endpoint_configuration {
    types = ["PRIVATE"]
  }

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
    x-amazon-apigateway-endpoint-configuration = {
      vpcEndpointIds = [aws_vpc_endpoint.test.id]
    }
  })
}
`, rName))
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

func testAccAWSAPIGatewayRestAPIConfigWithPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name   = %[1]q
  policy = <<EOF
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
  name   = %[1]q
  policy = <<EOF
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

func testAccAWSAPIGatewayRestAPIConfigApiKeySource(rName string, apiKeySource string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  api_key_source = %[2]q
  name           = %[1]q
}
`, rName, apiKeySource)
}

func testAccAWSAPIGatewayRestAPIConfigApiKeySourceOverrideBody(rName string, apiKeySource string, bodyApiKeySource string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  api_key_source = %[2]q
  name           = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
    x-amazon-apigateway-api-key-source = %[3]q
  })
}
`, rName, apiKeySource, bodyApiKeySource)
}

func testAccAWSAPIGatewayRestAPIConfigApiKeySourceSetByBody(rName string, bodyApiKeySource string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
    x-amazon-apigateway-api-key-source = %[2]q
  })
}
`, rName, bodyApiKeySource)
}

func testAccAWSAPIGatewayRestAPIConfigBinaryMediaTypes1(rName string, binaryMediaTypes1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  binary_media_types = [%[2]q]
  name               = %[1]q
}
`, rName, binaryMediaTypes1)
}

func testAccAWSAPIGatewayRestAPIConfigBinaryMediaTypes1OverrideBody(rName string, binaryMediaTypes1 string, bodyBinaryMediaTypes1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  binary_media_types = [%[2]q]
  name               = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
    x-amazon-apigateway-binary-media-types = [%[3]q]
  })
}
`, rName, binaryMediaTypes1, bodyBinaryMediaTypes1)
}

func testAccAWSAPIGatewayRestAPIConfigBinaryMediaTypes1SetByBody(rName string, bodyBinaryMediaTypes1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
    x-amazon-apigateway-binary-media-types = [%[2]q]
  })
}
`, rName, bodyBinaryMediaTypes1)
}

func testAccAWSAPIGatewayRestAPIConfigBody(rName string, basePath string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      %[2]q = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
  })
}
`, rName, basePath)
}

func testAccAWSAPIGatewayRestAPIConfigDescription(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  description = %[2]q
  name        = %[1]q
}
`, rName, description)
}

func testAccAWSAPIGatewayRestAPIConfigDescriptionOverrideBody(rName string, description string, bodyDescription string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  description = %[2]q
  name        = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      description = %[3]q
      title       = "test"
      version     = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
  })
}
`, rName, description, bodyDescription)
}

func testAccAWSAPIGatewayRestAPIConfigDescriptionSetByBody(rName string, bodyDescription string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      description = %[2]q
      title       = "test"
      version     = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
  })
}
`, rName, bodyDescription)
}

func testAccAWSAPIGatewayRestAPIConfigMinimumCompressionSize(rName string, minimumCompressionSize int) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  minimum_compression_size = %[2]d
  name                     = %[1]q
}
`, rName, minimumCompressionSize)
}

func testAccAWSAPIGatewayRestAPIConfigMinimumCompressionSizeOverrideBody(rName string, minimumCompressionSize int, bodyMinimumCompressionSize int) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  minimum_compression_size = %[2]d
  name                     = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
    x-amazon-apigateway-minimum-compression-size = %[3]d
  })
}
`, rName, minimumCompressionSize, bodyMinimumCompressionSize)
}

func testAccAWSAPIGatewayRestAPIConfigMinimumCompressionSizeSetByBody(rName string, bodyMinimumCompressionSize int) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
    x-amazon-apigateway-minimum-compression-size = %[2]d
  })
}
`, rName, bodyMinimumCompressionSize)
}

func testAccAWSAPIGatewayRestAPIConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAWSAPIGatewayRestAPIConfigNameOverrideBody(rName string, title string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = %[2]q
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
  })
}
`, rName, title)
}

func testAccAWSAPIGatewayRestAPIConfigParameters1(rName string, parameterKey1 string, parameterValue1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes  = ["https"]
    basePath = "/foo/bar/baz"
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
  })

  parameters = {
    %[2]s = %[3]q
  }
}
`, rName, parameterKey1, parameterValue1)
}

func testAccAWSAPIGatewayRestAPIConfigPolicyOverrideBody(rName string, bodyPath string, policyEffect string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      %[2]q = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
  })

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "execute-api:Invoke"
        Condition = {
          IpAddress = {
            "aws:SourceIp" = "123.123.123.123/32"
          }
        }
        Effect = %[3]q
        Principal = {
          AWS = "*"
        }
        Resource = "*"
      }
    ]
  })
}
`, rName, bodyPath, policyEffect)
}

func testAccAWSAPIGatewayRestAPIConfigPolicySetByBody(rName string, bodyPolicyEffect string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://aws.amazon.com/"
          }
        }
      }
    }
    x-amazon-apigateway-policy = {
      Version = "2012-10-17"
      Statement = [
        {
          Action = "execute-api:Invoke"
          Condition = {
            IpAddress = {
              "aws:SourceIp" = "123.123.123.123/32"
            }
          }
          Effect = %[2]q
          Principal = {
            AWS = "*"
          }
          Resource = "*"
        }
      ]
    }
  })
}
`, rName, bodyPolicyEffect)
}
