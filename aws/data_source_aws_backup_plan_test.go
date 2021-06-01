package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSBackupPlanDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_backup_plan.test"
	resourceName := "aws_backup_plan.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsBackupPlanDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`Error getting Backup Plan`),
			},
			{
				Config: testAccAwsBackupPlanDataSourceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "version", resourceName, "version"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

const testAccAwsBackupPlanDataSourceConfig_nonExistent = `
data "aws_backup_plan" "test" {
  plan_id = "tf-acc-test-does-not-exist"
}
`

func testAccAwsBackupPlanDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%[1]d"
}

resource "aws_backup_plan" "test" {
  name = "tf_acc_test_backup_plan_%[1]d"

  rule {
    rule_name         = "tf_acc_test_backup_rule_%[1]d"
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"
  }

  tags = {
    Name = "Value%[1]d"
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}

data "aws_backup_plan" "test" {
  plan_id = aws_backup_plan.test.id
}
`, rInt)
}
