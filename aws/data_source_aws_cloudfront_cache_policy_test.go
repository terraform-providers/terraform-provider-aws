package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSCloudFrontDataSourceCachePolicy_basic(t *testing.T) {
	rInt := acctest.RandInt()
	dataSourceName := "data.aws_cloudfront_cache_policy.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontCachePolicyDataSourceNameConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "comment", "test comment"),
					resource.TestCheckResourceAttr(dataSourceName, "default_ttl", "50"),
					resource.TestCheckResourceAttr(dataSourceName, "min_ttl", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "max_ttl", "100"),
					resource.TestCheckResourceAttr(dataSourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior", "whitelist"),
					resource.TestCheckResourceAttr(dataSourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.0.items.0", "test"),
					resource.TestCheckResourceAttr(dataSourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior", "whitelist"),
					resource.TestCheckResourceAttr(dataSourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.0.items.0", "test"),
					resource.TestCheckResourceAttr(dataSourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior", "whitelist"),
					resource.TestCheckResourceAttr(dataSourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.0.items.0", "test"),
				),
			},
			{
				ResourceName:            "aws_cloudfront_cache_policy.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccAWSCloudFrontCachePolicyDataSourceNameConfig(rInt int) string {
	return fmt.Sprintf(`
data "aws_cloudfront_cache_policy" "example" {
  name = aws_cloudfront_cache_policy.example.name
}

resource "aws_cloudfront_cache_policy" "example" {
  name        = "test-policy%[1]d"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
  parameters_in_cache_key_and_forwarded_to_origin {
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
}
`, rInt)
}
