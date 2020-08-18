package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsWafRegionalSubscribedRuleGroup_Basic(t *testing.T) {
	if os.Getenv("AWS_WAF_SUBSCRIBED_RULE_GROUP_NAME") == "" {
		t.Skip("Environment variable AWS_WAF_SUBSCRIBED_RULE_GROUP_NAME is not set")
	}

	ruleGroupName := os.Getenv("AWS_WAF_SUBSCRIBED_RULE_GROUP_NAME")

	if os.Getenv("AWS_WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME") == "" {
		t.Skip("Environment variable AWS_WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME is not set")
	}

	metricName := os.Getenv("AWS_WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME")

	datasourceName := "data.aws_wafregional_subscribed_rule_group.rulegroup"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafRegionalSubscribedRuleGroupConfig_NonExistent,
				ExpectError: regexp.MustCompile(`WAF Subscribed Rule Group not found`),
			},
			{
				Config: testAccDataSourceAwsWafRegionalSubscribedRuleGroupConfig_Name(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", metricName),
				),
			},
			{
				Config: testAccDataSourceAwsWafRegionalSubscribedRuleGroupConfig_MetricName(metricName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", metricName),
				),
			},
			{
				Config: testAccDataSourceAwsWafRegionalSubscribedRuleGroupConfig_NameAndMetricName(ruleGroupName, metricName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", metricName),
				),
			},
			{
				Config:      testAccDataSourceAwsWafRegionalSubscribedRuleGroupConfig_NameAndMismatchingMetricName(ruleGroupName),
				ExpectError: regexp.MustCompile(`WAF Subscribed Rule Group not found`),
			},
		},
	})
}

func testAccDataSourceAwsWafRegionalSubscribedRuleGroupConfig_Name(name string) string {
	return fmt.Sprintf(`
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  name = %[1]q
}
`, name)
}

func testAccDataSourceAwsWafRegionalSubscribedRuleGroupConfig_MetricName(metricName string) string {
	return fmt.Sprintf(`
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  metric_name = %[1]q
}
`, metricName)
}

func testAccDataSourceAwsWafRegionalSubscribedRuleGroupConfig_NameAndMetricName(name string, metricName string) string {
	return fmt.Sprintf(`
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  name = %[1]q
  metric_name = %[2]q
}
`, name, metricName)
}

func testAccDataSourceAwsWafRegionalSubscribedRuleGroupConfig_NameAndMismatchingMetricName(name string) string {
	return fmt.Sprintf(`
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  name = %[1]q
  metric_name = "tf-acc-test-does-not-exist"
}
`, name)
}

const testAccDataSourceAwsWafRegionalSubscribedRuleGroupConfig_NonExistent = `
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  name = "tf-acc-test-does-not-exist"
}
`
