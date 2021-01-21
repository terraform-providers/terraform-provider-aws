package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDataSourceIAMUser_basic(t *testing.T) {
	resourceName := "aws_iam_user.test"
	dataSourceName := "data.aws_iam_user.test"

	userName := fmt.Sprintf("test-datasource-user-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourceIAMUserConfig(userName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "unique_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions_boundary", ""),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMUser_tags(t *testing.T) {
	resourceName := "aws_iam_user.test"
	dataSourceName := "data.aws_iam_user.test"

	userName := fmt.Sprintf("test-datasource-user-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourceIAMUserConfig_tags(userName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
				),
			},
		},
	})
}

func testAccAwsDataSourceIAMUserConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = "%s"
  path = "/"
}

data "aws_iam_user" "test" {
  user_name = aws_iam_user.test.name
}
`, name)
}

func testAccAwsDataSourceIAMUserConfig_tags(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = "%s"
  path = "/"

  tags = {
    tag1 = "test-value1"
    tag2 = "test-value2"
  }
}

data "aws_iam_user" "test" {
  user_name = aws_iam_user.test.name
}
`, name)
}
