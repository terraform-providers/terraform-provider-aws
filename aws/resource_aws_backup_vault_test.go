package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsBackupVault_basic(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultExists("aws_backup_vault.test", &vault),
				),
			},
		},
	})
}

func TestAccAwsBackupVault_withKmsKey(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultWithKmsKey(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultExists("aws_backup_vault.test", &vault),
					resource.TestCheckResourceAttrPair("aws_backup_vault.test", "kms_key_arn", "aws_kms_key.test", "arn"),
				),
			},
		},
	})
}

func TestAccAwsBackupVault_withTags(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultExists("aws_backup_vault.test", &vault),
					resource.TestCheckResourceAttr("aws_backup_vault.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_backup_vault.test", "tags.up", "down"),
					resource.TestCheckResourceAttr("aws_backup_vault.test", "tags.left", "right"),
				),
			},
		},
	})
}

func testAccCheckAwsBackupVaultDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).backupconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_vault" {
			continue
		}

		input := &backup.DescribeBackupVaultInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeBackupVault(input)

		if err == nil {
			if *resp.BackupVaultName == rs.Primary.ID {
				return fmt.Errorf("Vault '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsBackupVaultExists(name string, vault *backup.DescribeBackupVaultOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).backupconn
		params := &backup.DescribeBackupVaultInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeBackupVault(params)
		if err != nil {
			return err
		}

		*vault = *resp

		return nil
	}
}

func testAccBackupVaultConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
	name = "tf_acc_test_backup_vault_%d"
}
`, randInt)
}

func testAccBackupVaultWithKmsKey(randInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
	description             = "Test KMS Key for AWS Backup Vault"
	deletion_window_in_days = 10
}

resource "aws_backup_vault" "test" {
	name = "tf_acc_test_backup_vault_%d"
	kms_key_arn = "${aws_kms_key.test.arn}"
}
`, randInt)
}

func testAccBackupVaultWithTags(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
	name = "tf_acc_test_backup_vault_%d"
	
	tags = {
		up		= "down"
		left	= "right"
	}
}
`, randInt)
}
