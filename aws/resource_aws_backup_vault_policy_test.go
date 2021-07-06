package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/backup/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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
	input := &backup.ListBackupVaultsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*testSweepResource, 0)

	err = conn.ListBackupVaultsPages(input, func(page *backup.ListBackupVaultsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vault := range page.BackupVaultList {
			r := resourceAwsBackupVaultPolicy()
			d := r.Data(nil)
			d.SetId(aws.StringValue(vault.BackupVaultName))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Backup Vault Policies sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Backup Vaults for %s: %w", region, err))
	}

	if err := testSweepResourceOrchestrator(sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Backup Vault Policies for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsBackupVaultPolicy_basic(t *testing.T) {
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_backup_vault_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		ErrorCheck:   testAccErrorCheck(t, backup.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, backup.EndpointsID),
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

func TestAccAwsBackupVaultPolicy_disappears_vault(t *testing.T) {
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_backup_vault_policy.test"
	vaultResourceName := "aws_backup_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		ErrorCheck:   testAccErrorCheck(t, backup.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultPolicyExists(resourceName, &vault),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsBackupVault(), vaultResourceName),
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

		_, err := finder.BackupVaultAccessPolicyByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Backup Vault Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsBackupVaultPolicyExists(name string, vault *backup.GetBackupVaultAccessPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Backup Vault Policy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).backupconn

		output, err := finder.BackupVaultAccessPolicyByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*vault = *output

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
