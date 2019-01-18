package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSCloudwatchLogGroupDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudwatchLogGroupDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cloudwatch_log_group.test", "name", rName),
					resource.TestCheckResourceAttrSet("data.aws_cloudwatch_log_group.test", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_cloudwatch_log_group.test", "creation_time"),
				),
			},
		},
	})
}

func testAccCheckAWSCloudwatchLogGroupDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  name = "%s"
}

data aws_cloudwatch_log_group "test" {
  name = "${aws_cloudwatch_log_group.test.name}"
}
`, rName)
}
