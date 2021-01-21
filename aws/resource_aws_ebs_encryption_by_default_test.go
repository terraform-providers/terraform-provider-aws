package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEBSEncryptionByDefault_basic(t *testing.T) {
	resourceName := "aws_ebs_encryption_by_default.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEncryptionByDefaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsEncryptionByDefaultConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsEncryptionByDefault(resourceName, false),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				Config: testAccAwsEbsEncryptionByDefaultConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsEncryptionByDefault(resourceName, true),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckAwsEncryptionByDefaultDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	response, err := conn.GetEbsEncryptionByDefault(&ec2.GetEbsEncryptionByDefaultInput{})
	if err != nil {
		return err
	}

	if aws.BoolValue(response.EbsEncryptionByDefault) != false {
		return fmt.Errorf("EBS encryption by default not disabled on resource removal")
	}

	return nil
}

func testAccCheckEbsEncryptionByDefault(n string, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		response, err := conn.GetEbsEncryptionByDefault(&ec2.GetEbsEncryptionByDefaultInput{})
		if err != nil {
			return err
		}

		if aws.BoolValue(response.EbsEncryptionByDefault) != enabled {
			return fmt.Errorf("EBS encryption by default is not in expected state (%t)", enabled)
		}

		return nil
	}
}

func testAccAwsEbsEncryptionByDefaultConfig(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ebs_encryption_by_default" "test" {
  enabled = %[1]t
}
`, enabled)
}
