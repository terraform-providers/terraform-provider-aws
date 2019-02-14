package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataSourceIAMUser_basic(t *testing.T) {
	userName := fmt.Sprintf("test-datasource-user-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourceIAMUserConfig(userName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_iam_user.test", "user_id"),
					resource.TestCheckResourceAttr("data.aws_iam_user.test", "path", "/"),
					resource.TestCheckResourceAttr("data.aws_iam_user.test", "permissions_boundary", ""),
					resource.TestCheckResourceAttr("data.aws_iam_user.test", "user_name", userName),
					resource.TestMatchResourceAttr("data.aws_iam_user.test", "arn", regexp.MustCompile("^arn:[^:]+:iam::[0-9]{12}:user/"+userName)),
				),
			},
		},
	})
}

func testAccAwsDataSourceIAMUserConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
	name = "%s"
	path = "/"
}

data "aws_iam_user" "test" {
	  user_name = "${aws_iam_user.user.name}"
}
`, name)
}
