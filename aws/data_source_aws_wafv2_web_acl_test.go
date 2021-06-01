package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsWafv2WebACL_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"
	datasourceName := "data.aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafv2WebACL_NonExistent(name),
				ExpectError: regexp.MustCompile(`WAFv2 WebACL not found`),
			},
			{
				Config: testAccDataSourceAwsWafv2WebACL_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					testAccMatchResourceAttrRegionalARN(datasourceName, "arn", "wafv2", regexp.MustCompile(fmt.Sprintf("regional/webacl/%v/.+$", name))),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "scope", resourceName, "scope"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafv2WebACL_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = "%s"
  scope = "REGIONAL"

  default_action {
    block {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-rule-metric-name"
    sampled_requests_enabled   = false
  }
}

data "aws_wafv2_web_acl" "test" {
  name  = aws_wafv2_web_acl.test.name
  scope = "REGIONAL"
}
`, name)
}

func testAccDataSourceAwsWafv2WebACL_NonExistent(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = "%s"
  scope = "REGIONAL"

  default_action {
    block {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-rule-metric-name"
    sampled_requests_enabled   = false
  }
}

data "aws_wafv2_web_acl" "test" {
  name  = "tf-acc-test-does-not-exist"
  scope = "REGIONAL"
}
`, name)
}
