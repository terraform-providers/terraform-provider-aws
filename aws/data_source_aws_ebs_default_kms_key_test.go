package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsEBSDefaultKmsKey_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEBSDefaultKmsKeyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceAwsEBSDefaultKmsKey("data.aws_ebs_default_kms_key.current"),
				),
			},
		},
	})
}

func testAccCheckDataSourceAwsEBSDefaultKmsKey(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		actual, err := conn.GetEbsDefaultKmsKeyId(&ec2.GetEbsDefaultKmsKeyIdInput{})
		if err != nil {
			return fmt.Errorf("Error reading EBS default KMS key: %q", err)
		}

		attr := rs.Primary.Attributes["key_arn"]

		if attr != *actual.KmsKeyId {
			return fmt.Errorf("EBS default KMS key is not the expected value (%v)", actual.KmsKeyId)
		}

		return nil
	}
}

const testAccDataSourceAwsEBSDefaultKmsKeyConfig = `
data "aws_ebs_default_kms_key" "current" {}
`
