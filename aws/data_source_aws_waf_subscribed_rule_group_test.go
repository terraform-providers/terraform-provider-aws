package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsWafSubscribedRuleGroup_Basic(t *testing.T) {
	if os.Getenv("AWS_WAF_SUBSCRIBED_RULE_GROUP_NAME") == "" {
		t.Skip("Environment variable AWS_WAF_SUBSCRIBED_RULE_GROUP_NAME is not set")
	}

	ruleGroupName := os.Getenv("AWS_WAF_SUBSCRIBED_RULE_GROUP_NAME")

	if os.Getenv("AWS_WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME") == "" {
		t.Skip("Environment variable AWS_WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME is not set")
	}

	metricName := os.Getenv("AWS_WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME")

	datasourceName := "data.aws_waf_subscribed_rule_group.rulegroup"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafSubscribedRuleGroupConfig_NonExistent,
				ExpectError: regexp.MustCompile(`WAF Subscribed Rule Group not found`),
			},
			{
				Config: testAccDataSourceAwsWafSubscribedRuleGroupConfig_Name(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", metricName),
				),
			},
			{
				Config: testAccDataSourceAwsWafSubscribedRuleGroupConfig_MetricName(metricName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", metricName),
				),
			},
			{
				Config: testAccDataSourceAwsWafSubscribedRuleGroupConfig_NameAndMetricName(ruleGroupName, metricName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", metricName),
				),
			},
			{
				Config:      testAccDataSourceAwsWafSubscribedRuleGroupConfig_NameAndMismatchingMetricName(ruleGroupName),
				ExpectError: regexp.MustCompile(`WAF Subscribed Rule Group not found`),
			},
		},
	})
}

func testAccDataSourceAwsWafSubscribedRuleGroupConfig_Name(name string) string {
	return fmt.Sprintf(`
data "aws_waf_subscribed_rule_group" "rulegroup" {
  name = %[1]q
}
`, name)
}

func testAccDataSourceAwsWafSubscribedRuleGroupConfig_MetricName(metricName string) string {
	return fmt.Sprintf(`
data "aws_waf_subscribed_rule_group" "rulegroup" {
  metric_name = %[1]q
}
`, metricName)
}

func testAccDataSourceAwsWafSubscribedRuleGroupConfig_NameAndMetricName(name string, metricName string) string {
	return fmt.Sprintf(`
data "aws_waf_subscribed_rule_group" "rulegroup" {
  name = %[1]q
  metric_name = %[2]q
}
`, name, metricName)
}

func testAccDataSourceAwsWafSubscribedRuleGroupConfig_NameAndMismatchingMetricName(name string) string {
	return fmt.Sprintf(`
data "aws_waf_subscribed_rule_group" "rulegroup" {
  name = %[1]q
  metric_name = "tf-acc-test-does-not-exist"
}
`, name)
}

const testAccDataSourceAwsWafSubscribedRuleGroupConfig_NonExistent = `
data "aws_waf_subscribed_rule_group" "rulegroup" {
  name = "tf-acc-test-does-not-exist"
}
`
