package aws

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsEBSEncryptionByDefault_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEBSEncryptionByDefaultConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceAwsEBSEncryptionByDefault("data.aws_ebs_encryption_by_default.current"),
				),
			},
		},
	})
}

func testAccCheckDataSourceAwsEBSEncryptionByDefault(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		actual, err := conn.GetEbsEncryptionByDefault(&ec2.GetEbsEncryptionByDefaultInput{})
		if err != nil {
			return fmt.Errorf("Error reading default EBS encryption toggle: %q", err)
		}

		attr, _ := strconv.ParseBool(rs.Primary.Attributes["enabled"])

		if attr != *actual.EbsEncryptionByDefault {
			return fmt.Errorf("EBS encryption by default is not in expected state (%t)", *actual.EbsEncryptionByDefault)
		}

		return nil
	}
}

const testAccDataSourceAwsEBSEncryptionByDefaultConfig = `
data "aws_ebs_encryption_by_default" "current" {}
`
