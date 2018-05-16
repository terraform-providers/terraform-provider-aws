package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
	conn := client.(*AWSClient).apigateway

	// https://github.com/terraform-providers/terraform-provider-aws/issues/3808
	prefixes := []string{
		"test",
		"tf_acc_",
		"tf-acc-",
	}

	err = conn.GetRestApisPages(&apigateway.GetRestApisInput{}, func(page *apigateway.GetRestApisOutput, lastPage bool) bool {
		for _, item := range page.Items {
			skip := true
			for _, prefix := range prefixes {
				if strings.HasPrefix(*item.Name, prefix) {
					skip = false
					break
				}
			}
			if skip {
				log.Printf("[INFO] Skipping API Gateway REST API: %s", *item.Name)
				continue
			}

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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists("aws_api_gateway_rest_api.test", &conf),
					testAccCheckAWSAPIGatewayRestAPINameAttribute(&conf, "bar"),
					testAccCheckAWSAPIGatewayRestAPIMinimumCompressionSizeAttribute(&conf, 0),
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "name", "bar"),
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "description", ""),
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "minimum_compression_size", "0"),
					resource.TestCheckResourceAttrSet("aws_api_gateway_rest_api.test", "created_date"),
					resource.TestCheckResourceAttrSet("aws_api_gateway_rest_api.test", "execution_arn"),
					resource.TestCheckNoResourceAttr("aws_api_gateway_rest_api.test", "binary_media_types"),
				),
			},

			{
				Config: testAccAWSAPIGatewayRestAPIUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists("aws_api_gateway_rest_api.test", &conf),
					testAccCheckAWSAPIGatewayRestAPINameAttribute(&conf, "test"),
					testAccCheckAWSAPIGatewayRestAPIDescriptionAttribute(&conf, "test"),
					testAccCheckAWSAPIGatewayRestAPIMinimumCompressionSizeAttribute(&conf, 10485760),
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "name", "test"),
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "description", "test"),
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "minimum_compression_size", "10485760"),
					resource.TestCheckResourceAttrSet("aws_api_gateway_rest_api.test", "created_date"),
					resource.TestCheckResourceAttrSet("aws_api_gateway_rest_api.test", "execution_arn"),
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "binary_media_types.0", "application/octet-stream"),
				),
			},

			{
				Config: testAccAWSAPIGatewayRestAPIDisableCompressionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists("aws_api_gateway_rest_api.test", &conf),
					testAccCheckAWSAPIGatewayRestAPIMinimumCompressionSizeAttributeIsNil(&conf),
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "minimum_compression_size", "-1"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_policy(t *testing.T) {
	expectedPolicyText := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":"*"},"Action":"execute-api:Invoke","Resource":"*"}]}`
	expectedUpdatePolicyText := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Principal":{"AWS":"*"},"Action":"execute-api:Invoke","Resource":"*"}]}`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigWithPolicy,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "policy", expectedPolicyText),
				),
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfigUpdatePolicy,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "policy", expectedUpdatePolicyText),
				),
			},
			{
				Config: testAccAWSAPIGatewayRestAPIConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "policy", ""),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApi_openapi(t *testing.T) {
	var conf apigateway.RestApi

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestAPIConfigOpenAPI,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists("aws_api_gateway_rest_api.test", &conf),
					testAccCheckAWSAPIGatewayRestAPINameAttribute(&conf, "test"),
					testAccCheckAWSAPIGatewayRestAPIRoutes(&conf, []string{"/", "/test"}),
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "name", "test"),
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "description", ""),
					resource.TestCheckResourceAttrSet("aws_api_gateway_rest_api.test", "created_date"),
					resource.TestCheckResourceAttrSet("aws_api_gateway_rest_api.test", "execution_arn"),
					resource.TestCheckNoResourceAttr("aws_api_gateway_rest_api.test", "binary_media_types"),
				),
			},

			{
				Config: testAccAWSAPIGatewayRestAPIUpdateConfigOpenAPI,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestAPIExists("aws_api_gateway_rest_api.test", &conf),
					testAccCheckAWSAPIGatewayRestAPINameAttribute(&conf, "test"),
					testAccCheckAWSAPIGatewayRestAPIRoutes(&conf, []string{"/", "/update"}),
					resource.TestCheckResourceAttr("aws_api_gateway_rest_api.test", "name", "test"),
					resource.TestCheckResourceAttrSet("aws_api_gateway_rest_api.test", "created_date"),
					resource.TestCheckResourceAttrSet("aws_api_gateway_rest_api.test", "execution_arn"),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayRestAPINameAttribute(conf *apigateway.RestApi, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *conf.Name != name {
			return fmt.Errorf("Wrong Name: %q", *conf.Name)
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
		conn := testAccProvider.Meta().(*AWSClient).apigateway

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

		conn := testAccProvider.Meta().(*AWSClient).apigateway

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
	conn := testAccProvider.Meta().(*AWSClient).apigateway

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

const testAccAWSAPIGatewayRestAPIConfig = `
resource "aws_api_gateway_rest_api" "test" {
  name = "bar"
  minimum_compression_size = 0
}
`

const testAccAWSAPIGatewayRestAPIConfigWithPolicy = `
resource "aws_api_gateway_rest_api" "test" {
  name = "bar"
  minimum_compression_size = 0
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
            "Resource": "*"
        }
    ]
}
EOF
}
`

const testAccAWSAPIGatewayRestAPIConfigUpdatePolicy = `
resource "aws_api_gateway_rest_api" "test" {
  name = "bar"
  minimum_compression_size = 0
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
`

const testAccAWSAPIGatewayRestAPIUpdateConfig = `
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
  description = "test"
  binary_media_types = ["application/octet-stream"]
  minimum_compression_size = 10485760
}
`

const testAccAWSAPIGatewayRestAPIDisableCompressionConfig = `
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
  description = "test"
  binary_media_types = ["application/octet-stream"]
  minimum_compression_size = -1
}
`

const testAccAWSAPIGatewayRestAPIConfigOpenAPI = `
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
  body = <<EOF
{
  "swagger": "2.0",
  "info": {
    "title": "test",
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
`

const testAccAWSAPIGatewayRestAPIUpdateConfigOpenAPI = `
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
  body = <<EOF
{
  "swagger": "2.0",
  "info": {
    "title": "test",
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
`
