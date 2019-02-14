package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsKmsKey_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsKmsKeyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsKmsKeyCheck("data.aws_kms_key.arbitrary"),
					resource.TestMatchResourceAttr("data.aws_kms_key.arbitrary", "arn", regexp.MustCompile("^arn:[^:]+:kms:[^:]+:[^:]+:key/.+")),
					resource.TestCheckResourceAttrSet("data.aws_kms_key.arbitrary", "aws_account_id"),
					resource.TestCheckResourceAttrSet("data.aws_kms_key.arbitrary", "creation_date"),
					resource.TestCheckResourceAttr("data.aws_kms_key.arbitrary", "description", "Terraform acc test"),
					resource.TestCheckResourceAttr("data.aws_kms_key.arbitrary", "enabled", "true"),
					resource.TestCheckResourceAttrSet("data.aws_kms_key.arbitrary", "key_manager"),
					resource.TestCheckResourceAttrSet("data.aws_kms_key.arbitrary", "key_state"),
					resource.TestCheckResourceAttr("data.aws_kms_key.arbitrary", "key_usage", "ENCRYPT_DECRYPT"),
					resource.TestCheckResourceAttrSet("data.aws_kms_key.arbitrary", "origin"),
				),
			},
		},
	})
}

func testAccDataSourceAwsKmsKeyCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		kmsKeyRs, ok := s.RootModule().Resources["aws_kms_key.arbitrary"]
		if !ok {
			return fmt.Errorf("can't find aws_kms_key.arbitrary in state")
		}

		attr := rs.Primary.Attributes

		checkProperties := []string{"arn", "key_usage", "description"}

		for _, p := range checkProperties {
			if attr[p] != kmsKeyRs.Primary.Attributes[p] {
				return fmt.Errorf(
					"%s is %s; want %s",
					p,
					attr[p],
					kmsKeyRs.Primary.Attributes[p],
				)
			}
		}

		return nil
	}
}

const testAccDataSourceAwsKmsKeyConfig = `
resource "aws_kms_key" "arbitrary" {
    description = "Terraform acc test"
    deletion_window_in_days = 7
}

data "aws_kms_key" "arbitrary" {
  key_id = "${aws_kms_key.arbitrary.key_id}"
}`
