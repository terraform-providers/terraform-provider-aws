package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSDbParameterGroupDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_db_parameter_group.test"
	resourceName := "aws_db_parameter_group.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsDbParameterGroupDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`Error getting DB Parameter Groups`),
			},
			{
				Config: testAccAwsDbParameterGroupDataSourceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "family", resourceName, "family"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

const testAccAwsDbParameterGroupDataSourceConfig_nonExistent = `
data "aws_db_parameter_group" "test" {
	name = "tf-acc-test-does-not-exist"
}
`

func testAccAwsDbParameterGroupDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
	name   = "tf-acc-test-%d"
	family = "postgres12"
  
	parameter {
	  name         = "client_encoding"
	  value        = "UTF8"
	  apply_method = "pending-reboot"
	}
}

data "aws_db_parameter_group" "test" {
  name = aws_db_parameter_group.test.name
}
`, rInt)
}
