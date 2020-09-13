package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSStepFunctionsActivityDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sfn_activity.test"
	dataName := "data.aws_sfn_activity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSStepFunctionsActivityDataSourceConfig_ActivityArn(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "creation_date", dataName, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataName, "name"),
				),
			},
			{
				Config: testAccCheckAWSStepFunctionsActivityDataSourceConfig_ActivityName(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "creation_date", dataName, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataName, "name"),
				),
			},
		},
	})
}

func testAccCheckAWSStepFunctionsActivityDataSourceConfig_ActivityArn(rName string) string {
	return fmt.Sprintf(`
resource aws_sfn_activity "test" {
  name = "%s"
}

data aws_sfn_activity "test" {
  arn = aws_sfn_activity.test.id
}
`, rName)
}

func testAccCheckAWSStepFunctionsActivityDataSourceConfig_ActivityName(rName string) string {
	return fmt.Sprintf(`
resource aws_sfn_activity "test" {
  name = "%s"
}

data aws_sfn_activity "test" {
  name = aws_sfn_activity.test.name
}
`, rName)
}
