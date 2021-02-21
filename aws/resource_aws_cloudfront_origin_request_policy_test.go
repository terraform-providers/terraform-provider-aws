package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSCloudFrontOriginRequestPolicy_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_cloudfront_origin_request_policy.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontOriginRequestPolicyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "comment", "test comment"),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookie_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookies.0.items.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.header_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.headers.0.items.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_string_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_strings.0.items.0", "test"),
				),
			},
			{
				ResourceName:            "aws_cloudfront_origin_request_policy.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccAWSCloudFrontOriginRequestPolicy_update(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_cloudfront_origin_request_policy.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontOriginRequestPolicyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "comment", "test comment"),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookie_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookies.0.items.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.header_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.headers.0.items.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_string_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_strings.0.items.0", "test"),
				),
			},
			{
				Config: testAccAWSCloudFrontOriginRequestPolicyConfigUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "comment", "test comment updated"),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookies.0.items.0", "test2"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.header_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.headers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_strings.0.items.0", "test2"),
				),
			},
			{
				ResourceName:            "aws_cloudfront_origin_request_policy.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccAWSCloudFrontOriginRequestPolicy_noneBehavior(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_cloudfront_origin_request_policy.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontOriginRequestPolicyConfigNoneBehavior(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "comment", "test comment"),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookie_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookies.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.header_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.headers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_string_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_strings.#", "0"),
				),
			},
			{
				ResourceName:            "aws_cloudfront_origin_request_policy.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccAWSCloudFrontOriginRequestPolicyConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_request_policy" "example" {
  name    = "test-policy%[1]d"
  comment = "test comment"
  cookies_config {
    cookie_behavior = "whitelist"
    cookies {
      items = ["test"]
    }
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = ["test"]
    }
  }
  query_strings_config {
    query_string_behavior = "whitelist"
    query_strings {
      items = ["test"]
    }
  }
}
`, rInt)
}

func testAccAWSCloudFrontOriginRequestPolicyConfigUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_request_policy" "example" {
  name    = "test-policy-updated%[1]d"
  comment = "test comment updated"
  cookies_config {
    cookie_behavior = "whitelist"
    cookies {
      items = ["test2"]
    }
  }
  headers_config {
    header_behavior = "none"
  }
  query_strings_config {
    query_string_behavior = "whitelist"
    query_strings {
      items = ["test2"]
    }
  }
}
`, rInt)
}

func testAccAWSCloudFrontOriginRequestPolicyConfigNoneBehavior(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_request_policy" "example" {
  name    = "test-policy-updated%[1]d"
  comment = "test comment"
  cookies_config {
    cookie_behavior = "none"
  }
  headers_config {
    header_behavior = "none"
  }
  query_strings_config {
    query_string_behavior = "none"
  }
}
`, rInt)
}
