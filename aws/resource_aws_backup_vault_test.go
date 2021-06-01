package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsBackupVault_basic(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := acctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultExists(resourceName, &vault),
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

func TestAccAwsBackupVault_withKmsKey(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := acctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultWithKmsKey(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", "aws_kms_key.test", "arn"),
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

func TestAccAwsBackupVault_withTags(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := acctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.up", "down"),
					resource.TestCheckResourceAttr(resourceName, "tags.left", "right"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBackupVaultWithUpdateTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.up", "downdown"),
					resource.TestCheckResourceAttr(resourceName, "tags.left", "rightright"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccBackupVaultWithRemoveTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
		},
	})
}

func TestAccAwsBackupVault_disappears(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := acctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultExists(resourceName, &vault),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsBackupVault(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

func testAccPreCheckAWSBackup(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).backupconn

	input := &backup.ListBackupVaultsInput{}

	_, err := conn.ListBackupVaults(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
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
  name        = "tf_acc_test_backup_vault_%d"
  kms_key_arn = aws_kms_key.test.arn
}
`, randInt)
}

func testAccBackupVaultWithTags(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"

  tags = {
    up   = "down"
    left = "right"
  }
}
`, randInt)
}

func testAccBackupVaultWithUpdateTags(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"

  tags = {
    up   = "downdown"
    left = "rightright"
    foo  = "bar"
    fizz = "buzz"
  }
}
`, randInt)
}

func testAccBackupVaultWithRemoveTags(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"

  tags = {
    foo  = "bar"
    fizz = "buzz"
  }
}
`, randInt)
}
