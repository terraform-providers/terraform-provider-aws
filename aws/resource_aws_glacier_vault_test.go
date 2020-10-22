package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_glacier_vault", &resource.Sweeper{
		Name: "aws_glacier_vault",
		F:    testSweepGlacierVaults,
	})
}

func testSweepGlacierVaults(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).glacierconn
	var sweeperErrs *multierror.Error

	err = conn.ListVaultsPages(&glacier.ListVaultsInput{}, func(page *glacier.ListVaultsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, vault := range page.VaultList {
			name := aws.StringValue(vault.VaultName)

			// First attempt to delete the vault's notification configuration in case the vault deletion fails.
			log.Printf("[INFO] Deleting Glacier Vault (%s) Notifications", name)
			_, err := conn.DeleteVaultNotifications(&glacier.DeleteVaultNotificationsInput{
				VaultName: aws.String(name),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Glacier Vault (%s) Notifications: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			log.Printf("[INFO] Deleting Glacier Vault: %s", name)
			_, err = conn.DeleteVault(&glacier.DeleteVaultInput{
				VaultName: aws.String(name),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Glacier Vault (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !isLast
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Glacier Vaults sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Glacier Vaults: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSGlacierVault_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_glacier_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlacierVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlacierVault_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName),
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

func TestAccAWSGlacierVault_full(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_glacier_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlacierVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlacierVault_full(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName),
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

func TestAccAWSGlacierVault_RemoveNotifications(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_glacier_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlacierVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlacierVault_full(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlacierVault_withoutNotification(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName),
					testAccCheckVaultNotificationsMissing(resourceName),
				),
			},
		},
	})
}

func testAccCheckGlacierVaultExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		glacierconn := testAccProvider.Meta().(*AWSClient).glacierconn
		out, err := glacierconn.DescribeVault(&glacier.DescribeVaultInput{
			VaultName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if out.VaultARN == nil {
			return fmt.Errorf("No Glacier Vault Found")
		}

		if *out.VaultName != rs.Primary.ID {
			return fmt.Errorf("Glacier Vault Mismatch - existing: %q, state: %q",
				*out.VaultName, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVaultNotificationsMissing(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		glacierconn := testAccProvider.Meta().(*AWSClient).glacierconn
		out, err := glacierconn.GetVaultNotifications(&glacier.GetVaultNotificationsInput{
			VaultName: aws.String(rs.Primary.ID),
		})

		if awserr, ok := err.(awserr.Error); ok && awserr.Code() != "ResourceNotFoundException" {
			return fmt.Errorf("Expected ResourceNotFoundException for Vault %s Notification Block but got %s", rs.Primary.ID, awserr.Code())
		}

		if out.VaultNotificationConfig != nil {
			return fmt.Errorf("Vault Notification Block has been found for %s", rs.Primary.ID)
		}

		return nil
	}

}

func testAccCheckGlacierVaultDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).glacierconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glacier_vault" {
			continue
		}

		input := &glacier.DescribeVaultInput{
			VaultName: aws.String(rs.Primary.ID),
		}
		if _, err := conn.DescribeVault(input); err != nil {
			// Verify the error is what we want
			if ae, ok := err.(awserr.Error); ok && ae.Code() == "ResourceNotFoundException" {
				continue
			}

			return err
		}
		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccGlacierVault_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_glacier_vault" "test" {
  name = "my_test_vault_%d"
}
`, rInt)
}

func testAccGlacierVault_full(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "glacier-sns-topic-%d"
}

resource "aws_glacier_vault" "test" {
  name = "my_test_vault_%d"

  notification {
    sns_topic = aws_sns_topic.aws_sns_topic.arn
    events    = ["ArchiveRetrievalCompleted", "InventoryRetrievalCompleted"]
  }

  tags = {
    Test = "Test1"
  }
}
`, rInt, rInt)
}

func testAccGlacierVault_withoutNotification(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "glacier-sns-topic-%d"
}

resource "aws_glacier_vault" "test" {
  name = "my_test_vault_%d"

  tags = {
    Test = "Test1"
  }
}
`, rInt, rInt)
}
