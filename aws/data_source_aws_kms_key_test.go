package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsKmsKey_basic(t *testing.T) {
	resourceName := "aws_kms_key.test"
	datasourceName := "data.aws_kms_key.test"
	rName := fmt.Sprintf("tf-testacc-kms-key-%s", acctest.RandString(13))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsKmsKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsKmsKeyCheck(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "customer_master_key_spec", resourceName, "customer_master_key_spec"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "enabled", resourceName, "is_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "key_usage", resourceName, "key_usage"),
					resource.TestCheckResourceAttrSet(datasourceName, "aws_account_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(datasourceName, "key_manager"),
					resource.TestCheckResourceAttrSet(datasourceName, "key_state"),
					resource.TestCheckResourceAttrSet(datasourceName, "origin"),
				),
			},
		},
	})
}

func testAccDataSourceAwsKmsKeyCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

func testAccDataSourceAwsKmsKeyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

data "aws_kms_key" "test" {
  key_id = aws_kms_key.test.key_id
}
`, rName)
}
