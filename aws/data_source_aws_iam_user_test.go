package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDataSourceIAMUser_basic(t *testing.T) {
	resourceName := "data.aws_iam_user.test"

	userName := fmt.Sprintf("test-datasource-user-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourceIAMUserConfig(userName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "user_id", "aws_iam_user.user", "unique_id"),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					resource.TestCheckResourceAttr(resourceName, "user_name", userName),
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_iam_user.user", "arn"),
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
  user_name = aws_iam_user.user.name
}
`, name)
}
