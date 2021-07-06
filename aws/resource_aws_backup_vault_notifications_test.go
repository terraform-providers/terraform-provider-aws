package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_backup_vault_notifications", &resource.Sweeper{
		Name: "aws_backup_vault_notifications",
		F:    testSweepBackupVaultNotifications,
	})
}

func testSweepBackupVaultNotifications(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}

	conn := client.(*AWSClient).backupconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &backup.ListBackupVaultsInput{}

	err = conn.ListBackupVaultsPages(input, func(page *backup.ListBackupVaultsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vault := range page.BackupVaultList {
			if vault == nil {
				continue
			}

			r := resourceAwsBackupVaultNotifications()
			d := r.Data(nil)
			d.SetId(aws.StringValue(vault.BackupVaultName))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Backup Vaults for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Backup Vault Notifications for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Backup Vault Notifications sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAwsBackupVaultNotification_basic(t *testing.T) {
	var vault backup.GetBackupVaultNotificationsOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_backup_vault_notifications.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		ErrorCheck:   testAccErrorCheck(t, backup.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultNotificationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultNotificationExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "backup_vault_events.#", "2"),
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

func TestAccAwsBackupVaultNotification_disappears(t *testing.T) {
	var vault backup.GetBackupVaultNotificationsOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_backup_vault_notifications.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		ErrorCheck:   testAccErrorCheck(t, backup.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupVaultNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultNotificationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultNotificationExists(resourceName, &vault),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsBackupVaultNotifications(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsBackupVaultNotificationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).backupconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_vault_notifications" {
			continue
		}

		input := &backup.GetBackupVaultNotificationsInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetBackupVaultNotifications(input)

		if err == nil {
			if aws.StringValue(resp.BackupVaultName) == rs.Primary.ID {
				return fmt.Errorf("Backup Plan notifications '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsBackupVaultNotificationExists(name string, vault *backup.GetBackupVaultNotificationsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).backupconn
		params := &backup.GetBackupVaultNotificationsInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetBackupVaultNotifications(params)
		if err != nil {
			return err
		}

		*vault = *resp

		return nil
	}
}

func testAccBackupVaultNotificationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

data "aws_iam_policy_document" "test" {
  policy_id = "__default_policy_ID"

  statement {
    actions = [
      "SNS:Publish",
    ]

    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["backup.amazonaws.com"]
    }

    resources = [
      "${aws_sns_topic.test.arn}",
    ]

    sid = "__default_statement_ID"
  }
}

resource "aws_sns_topic_policy" "test" {
  arn    = aws_sns_topic.test.arn
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_backup_vault_notifications" "test" {
  backup_vault_name   = aws_backup_vault.test.name
  sns_topic_arn       = aws_sns_topic.test.arn
  backup_vault_events = ["BACKUP_JOB_STARTED", "RESTORE_JOB_COMPLETED"]
}
`, rName)
}
