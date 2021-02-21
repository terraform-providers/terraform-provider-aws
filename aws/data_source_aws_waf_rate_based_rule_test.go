package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsWafRateBasedRule_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_rate_based_rule.wafrule"
	datasourceName := "data.aws_waf_rate_based_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(waf.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafRateBasedRuleConfig_NonExistent,
				ExpectError: regexp.MustCompile(`WAF Rate Based Rules not found`),
			},
			{
				Config: testAccDataSourceAwsWafRateBasedRuleConfig_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafRateBasedRuleConfig_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = "WafruleTest"
  rate_key    = "IP"
  rate_limit  = 2000
}

data "aws_waf_rate_based_rule" "wafrule" {
  name = aws_waf_rate_based_rule.wafrule.name
}
`, name)
}

const testAccDataSourceAwsWafRateBasedRuleConfig_NonExistent = `
data "aws_waf_rate_based_rule" "wafrule" {
  name = "tf-acc-test-does-not-exist"
}
`
