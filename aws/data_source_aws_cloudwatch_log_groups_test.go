package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSCloudwatchLogGroupsDataSource_basic(t *testing.T) {
	rName1 := acctest.RandomWithPrefix("/abc/tf-acc-test1")
	rName2 := acctest.RandomWithPrefix("/abc/tf-acc-test2")

	resourceName := "data.aws_cloudwatch_log_groups.blah"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudwatchLogGroupsDataSourceConfig(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "arns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "arns.0"),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.0", rName1),
				),
			},
		},
	})
}

func testAccCheckAWSCloudwatchLogGroupsDataSourceConfig(rName1 string, rName2 string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_log_group "test1" {
  name = "%s"
}

resource aws_cloudwatch_log_group "test2" {
  name = "%s"
}

data aws_cloudwatch_log_groups "blah" {
	log_group_name_prefix = "/abc"
  depends_on = [aws_cloudwatch_log_group.test1,aws_cloudwatch_log_group.test2]
}
`, rName1, rName2)
}
