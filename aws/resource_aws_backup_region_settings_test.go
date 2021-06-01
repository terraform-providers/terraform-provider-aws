package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsBackupRegionSettings_basic(t *testing.T) {
	var settings backup.DescribeRegionSettingsOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_backup_region_settings.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(fsx.EndpointsID, t)
			testAccPreCheckAWSBackup(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupRegionSettingsConfig1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupRegionSettingsExists(&settings),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.%", "8"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.DynamoDB", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Aurora", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EBS", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EC2", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EFS", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.FSx", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.RDS", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Storage Gateway", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBackupRegionSettingsConfig2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupRegionSettingsExists(&settings),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.%", "8"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.DynamoDB", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Aurora", "false"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EBS", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EC2", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EFS", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.FSx", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.RDS", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Storage Gateway", "true"),
				),
			},
			{
				Config: testAccBackupRegionSettingsConfig1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupRegionSettingsExists(&settings),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.%", "8"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.DynamoDB", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Aurora", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EBS", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EC2", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.EFS", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.FSx", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.RDS", "true"),
					resource.TestCheckResourceAttr(resourceName, "resource_type_opt_in_preference.Storage Gateway", "true"),
				),
			},
		},
	})
}

func testAccCheckAwsBackupRegionSettingsExists(settings *backup.DescribeRegionSettingsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*AWSClient).backupconn
		resp, err := conn.DescribeRegionSettings(&backup.DescribeRegionSettingsInput{})
		if err != nil {
			return err
		}

		*settings = *resp

		return nil
	}
}

func testAccBackupRegionSettingsConfig1(rName string) string {
	return `
resource "aws_backup_region_settings" "test" {
  resource_type_opt_in_preference = {
    "DynamoDB"        = true
    "Aurora"          = true
    "EBS"             = true
    "EC2"             = true
    "EFS"             = true
    "FSx"             = true
    "RDS"             = true
    "Storage Gateway" = true
  }
}
`
}

func testAccBackupRegionSettingsConfig2(rName string) string {
	return `
resource "aws_backup_region_settings" "test" {
  resource_type_opt_in_preference = {
    "DynamoDB"        = true
    "Aurora"          = false
    "EBS"             = true
    "EC2"             = true
    "EFS"             = true
    "FSx"             = true
    "RDS"             = true
    "Storage Gateway" = true
  }
}
`
}
