package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSBackupVaultDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_backup_vault.test"
	resourceName := "aws_backup_vault.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsBackupVaultDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`Error getting Backup Vault`),
			},
			{
				Config: testAccAwsBackupVaultDataSourceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "kms_key_arn", resourceName, "kms_key_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "recovery_points", resourceName, "recovery_points"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

const testAccAwsBackupVaultDataSourceConfig_nonExistent = `
data "aws_backup_vault" "test" {
  name = "tf-acc-test-does-not-exist"
}
`

func testAccAwsBackupVaultDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"

  tags = {
    up   = "down"
    left = "right"
  }
}

data "aws_backup_vault" "test" {
  name = aws_backup_vault.test.name
}
`, rInt)
}
