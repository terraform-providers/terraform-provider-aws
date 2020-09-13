package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsWafRegionalRateBasedRule_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	datasourceName := "data.aws_wafregional_rate_based_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafRegionalRateBasedRuleConfig_NonExistent,
				ExpectError: regexp.MustCompile(`WAF Rate Based Rules not found`),
			},
			{
				Config: testAccDataSourceAwsWafRegionalRateBasedRuleConfig_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafRegionalRateBasedRuleConfig_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = "WafruleTest"
  rate_key    = "IP"
  rate_limit  = 2000
}

data "aws_wafregional_rate_based_rule" "wafrule" {
  name = aws_wafregional_rate_based_rule.wafrule.name
}
`, name)
}

const testAccDataSourceAwsWafRegionalRateBasedRuleConfig_NonExistent = `
data "aws_wafregional_rate_based_rule" "wafrule" {
  name = "tf-acc-test-does-not-exist"
}
`
