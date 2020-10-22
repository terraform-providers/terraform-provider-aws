package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsWafRegionalRule_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafregional_rule.wafrule"
	datasourceName := "data.aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafRegionalRuleConfig_NonExistent,
				ExpectError: regexp.MustCompile(`WAF Rule not found`),
			},
			{
				Config: testAccDataSourceAwsWafRegionalRuleConfig_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafRegionalRuleConfig_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = "WafruleTest"
}

data "aws_wafregional_rule" "wafrule" {
  name = aws_wafregional_rule.wafrule.name
}
`, name)
}

const testAccDataSourceAwsWafRegionalRuleConfig_NonExistent = `
data "aws_wafregional_rule" "wafrule" {
  name = "tf-acc-test-does-not-exist"
}
`
