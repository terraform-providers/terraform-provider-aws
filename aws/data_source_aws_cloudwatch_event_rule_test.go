package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSCloudwatchEventRuleDataSource_Basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "data.aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudwatchEventRuleDataSourceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
		},
	})
}

func testAccCheckAWSCloudwatchEventRuleDataSourceConfig_Basic(rName string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_event_rule "test" {
  name = "%s"
  schedule_expression = "rate(1 hour)"
}

data aws_cloudwatch_event_rule "test" {
  name = aws_cloudwatch_event_rule.test.name
}
`, rName)
}
