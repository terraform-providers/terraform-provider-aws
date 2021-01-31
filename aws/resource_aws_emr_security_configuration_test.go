package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEmrSecurityConfiguration_basic(t *testing.T) {
	resourceName := "aws_emr_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEmrSecurityConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmrSecurityConfigurationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmrSecurityConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "creation_date"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckEmrSecurityConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).emrconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emr_security_configuration" {
			continue
		}

		// Try to find the Security Configuration
		resp, err := conn.DescribeSecurityConfiguration(&emr.DescribeSecurityConfigurationInput{
			Name: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, "InvalidRequestException", "does not exist") {
			return nil
		}

		if err != nil {
			return err
		}

		if resp != nil && aws.StringValue(resp.Name) == rs.Primary.ID {
			return fmt.Errorf("Error: EMR Security Configuration still exists: %s", aws.StringValue(resp.Name))
		}

		return nil
	}

	return nil
}

func testAccCheckEmrSecurityConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EMR Security Configuration ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).emrconn
		resp, err := conn.DescribeSecurityConfiguration(&emr.DescribeSecurityConfigurationInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if resp.Name == nil {
			return fmt.Errorf("EMR Security Configuration had nil name which shouldn't happen")
		}

		if *resp.Name != rs.Primary.ID {
			return fmt.Errorf("EMR Security Configuration name mismatch, got (%s), expected (%s)", *resp.Name, rs.Primary.ID)
		}

		return nil
	}
}

const testAccEmrSecurityConfigurationConfig = `
data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_emr_security_configuration" "test" {
  configuration = <<EOF
{
  "EncryptionConfiguration": {
    "AtRestEncryptionConfiguration": {
      "S3EncryptionConfiguration": {
        "EncryptionMode": "SSE-S3"
      },
      "LocalDiskEncryptionConfiguration": {
        "EncryptionKeyProviderType": "AwsKms",
        "AwsKmsKey": "arn:${data.aws_partition.current.partition}:kms:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:alias/tf_emr_test_key"
      }
    },
    "EnableInTransitEncryption": false,
    "EnableAtRestEncryption": true
  }
}
EOF
}
`
