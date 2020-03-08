package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"os"
	"regexp"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSDataSourceQuickSightGroup_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	namespace := "default"
	description := "This is just a test group to test datasource aws_quicksight_group"
	region := os.Getenv("AWS_DEFAULT_REGION")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceQuickSightGroupConfig(rName, description, namespace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_quicksight_group.test", "group_id"),
					resource.TestCheckResourceAttr("data.aws_quicksight_group.test", "description", description),
					resource.TestCheckResourceAttr("data.aws_quicksight_group.test", "group_name", rName),
					resource.TestMatchResourceAttr("data.aws_quicksight_group.test", "arn", regexp.MustCompile("^arn:[^:]+:quicksight:"+region+":[0-9]{12}:group/"+namespace+"/"+rName)),
				),
			},
		},
	})
}

func testAccAWSDataSourceQuickSightGroupConfig(rName, description, namespace string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_group" "default" {
  aws_account_id = "${data.aws_caller_identity.current.account_id}"
  group_name     = %[1]q
  description    = %[2]q
}

data "aws_quicksight_group" "test" {
  group_name 	 = "${aws_quicksight_group.default.group_name}"
  aws_account_id = "${data.aws_caller_identity.current.account_id}"
  namespace		 = %[3]q
}

`, rName, description, namespace)
}
