package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSEFSAccessPoint_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_efs_access_point.test"
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSEFSAccessPointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
					resource.TestCheckResourceAttrPair(dataSourceName, "posix_user", resourceName, "posix_user"),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_directory", resourceName, "root_directory"),
				),
			},
		},
	})
}

func testAccDataSourceAWSEFSAccessPointConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = "%s"
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
}

data "aws_efs_access_point" "test" {
  access_point_id = aws_efs_access_point.test.id
}
`, rName)
}
