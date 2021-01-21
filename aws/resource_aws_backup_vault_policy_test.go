package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_backup_vault_policy", &resource.Sweeper{
		Name: "aws_backup_vault_policy",
		F:    testSweepBackupVaultPolicies,
	})
}

func testSweepBackupVaultPolicies(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*AWSClient).backupconn
	var sweeperErrs *multierror.Error

	input := &backup.ListBackupVaultsInput{}

	for {
		output, err := conn.ListBackupVaults(input)
		if err != nil {
			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping Backup Vault Policies sweep for %s: %s", region, err)
				return nil
			}
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Backup Vault Policies: %w", err))
			return sweeperErrs.ErrorOrNil()
		}

		if len(output.BackupVaultList) == 0 {
			log.Print("[DEBUG] No Backup Vault Policies to sweep")
			return nil
		}

		for _, rule := range output.BackupVaultList {
			name := aws.StringValue(rule.BackupVaultName)

			log.Printf("[INFO] Deleting Backup Vault Policies %s", name)
			_, err := conn.DeleteBackupVaultAccessPolicy(&backup.DeleteBackupVaultAccessPolicyInput{
				BackupVaultName: aws.String(name),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Backup Vault Policies %s: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsBackupVaultPolicy_basic(t *testing.T) {
	var vault backup.GetBackupVaultAccessPolicyOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_backup_vault_policy.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultPolicyExists(resourceName, &vault),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("^{\"Version\":\"2012-10-17\".+"))),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBackupVaultPolicyConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultPolicyExists(resourceName, &vault),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("^{\"Version\":\"2012-10-17\".+")),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("backup:ListRecoveryPointsByBackupVault")),
				),
			},
		},
	})
}

func TestAccAwsBackupVaultPolicy_disappears(t *testing.T) {
	var vault backup.GetBackupVaultAccessPolicyOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_backup_vault_policy.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultPolicyExists(resourceName, &vault),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsBackupVaultPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsBackupVaultPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).backupconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_vault_policy" {
			continue
		}

		input := &backup.GetBackupVaultAccessPolicyInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetBackupVaultAccessPolicy(input)

		if err == nil {
			if aws.StringValue(resp.BackupVaultName) == rs.Primary.ID {
				return fmt.Errorf("Backup Plan Policies '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsBackupVaultPolicyExists(name string, vault *backup.GetBackupVaultAccessPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).backupconn
		params := &backup.GetBackupVaultAccessPolicyInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetBackupVaultAccessPolicy(params)
		if err != nil {
			return err
		}

		*vault = *resp

		return nil
	}
}

func testAccBackupVaultPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_vault_policy" "test" {
  backup_vault_name = aws_backup_vault.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "default",
  "Statement": [
    {
      "Sid": "default",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": [
		"backup:DescribeBackupVault",
		"backup:DeleteBackupVault",
		"backup:PutBackupVaultAccessPolicy",
		"backup:DeleteBackupVaultAccessPolicy",
		"backup:GetBackupVaultAccessPolicy",
		"backup:StartBackupJob",
		"backup:GetBackupVaultNotifications",
		"backup:PutBackupVaultNotifications"
      ],
      "Resource": "${aws_backup_vault.test.arn}"
    }
  ]
}
POLICY
}
`, rName)
}

func testAccBackupVaultPolicyConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_vault_policy" "test" {
  backup_vault_name = aws_backup_vault.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "default",
  "Statement": [
    {
      "Sid": "default",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": [
		"backup:DescribeBackupVault",
		"backup:DeleteBackupVault",
		"backup:PutBackupVaultAccessPolicy",
		"backup:DeleteBackupVaultAccessPolicy",
		"backup:GetBackupVaultAccessPolicy",
		"backup:StartBackupJob",
		"backup:GetBackupVaultNotifications",
		"backup:PutBackupVaultNotifications",
		"backup:ListRecoveryPointsByBackupVault"
      ],
      "Resource": "${aws_backup_vault.test.arn}"
    }
  ]
}
POLICY
}
`, rName)
}
