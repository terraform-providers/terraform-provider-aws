package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
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
	var vault glacier.DescribeVaultOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glacier_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlacierVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlacierVaultBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "glacier", regexp.MustCompile(`vaults/.+`)),
					resource.TestCheckResourceAttr(resourceName, "notification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "access_policy", ""),
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

func TestAccAWSGlacierVault_notification(t *testing.T) {
	var vault glacier.DescribeVaultOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glacier_vault.test"
	snsResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlacierVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlacierVaultNotificationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "notification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification.0.events.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "notification.0.sns_topic", snsResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlacierVaultBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "notification.#", "0"),
					testAccCheckVaultNotificationsMissing(resourceName),
				),
			},
			{
				Config: testAccGlacierVaultNotificationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "notification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification.0.events.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "notification.0.sns_topic", snsResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSGlacierVault_policy(t *testing.T) {
	var vault glacier.DescribeVaultOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glacier_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlacierVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlacierVaultPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestMatchResourceAttr(resourceName, "access_policy",
						regexp.MustCompile(`"Sid":"cross-account-upload".+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlacierVaultPolicyConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestMatchResourceAttr(resourceName, "access_policy",
						regexp.MustCompile(`"Sid":"cross-account-upload1".+`)),
				),
			},
			{
				Config: testAccGlacierVaultBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "access_policy", ""),
				),
			},
		},
	})
}

func TestAccAWSGlacierVault_tags(t *testing.T) {
	var vault glacier.DescribeVaultOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glacier_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlacierVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlacierVaultConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlacierVaultConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGlacierVaultConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGlacierVault_disappears(t *testing.T) {
	var vault glacier.DescribeVaultOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glacier_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlacierVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlacierVaultBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlacierVaultExists(resourceName, &vault),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGlacierVault(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGlacierVaultExists(name string, vault *glacier.DescribeVaultOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).glacierconn
		out, err := conn.DescribeVault(&glacier.DescribeVaultInput{
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

		*vault = *out

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

		conn := testAccProvider.Meta().(*AWSClient).glacierconn
		out, err := conn.GetVaultNotifications(&glacier.GetVaultNotificationsInput{
			VaultName: aws.String(rs.Primary.ID),
		})

		if !isAWSErr(err, glacier.ErrCodeResourceNotFoundException, "") {
			return fmt.Errorf("Expected ResourceNotFoundException for Vault %s Notification Block but got %s", rs.Primary.ID, err)
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
			if isAWSErr(err, glacier.ErrCodeResourceNotFoundException, "") {
				continue
			}

			return err
		}
		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccGlacierVaultBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_glacier_vault" "test" {
  name = %[1]q
}
`, rName)
}

func testAccGlacierVaultNotificationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_glacier_vault" "test" {
  name = %[1]q

  notification {
    sns_topic = aws_sns_topic.test.arn
    events    = ["ArchiveRetrievalCompleted", "InventoryRetrievalCompleted"]
  }
}
`, rName)
}

func testAccGlacierVaultPolicyConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_glacier_vault" "test" {
  name = %[1]q

  access_policy = <<EOF
{
    "Version":"2012-10-17",
    "Statement":[
       {
          "Sid":"cross-account-upload",
          "Principal": {
             "AWS": ["*"]
          },
          "Effect":"Allow",
          "Action": [
             "glacier:InitiateMultipartUpload",
             "glacier:AbortMultipartUpload",
             "glacier:CompleteMultipartUpload"
          ],
          "Resource": ["arn:${data.aws_partition.current.partition}:glacier:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:vaults/%[1]s"]
       }
    ]
}
EOF
}
`, rName)
}

func testAccGlacierVaultPolicyConfigUpdated(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_glacier_vault" "test" {
  name = %[1]q

  access_policy = <<EOF
{
    "Version":"2012-10-17",
    "Statement":[
       {
          "Sid":"cross-account-upload1",
          "Principal": {
             "AWS": ["*"]
          },
          "Effect":"Allow",
          "Action": [
             "glacier:UploadArchive",
             "glacier:InitiateMultipartUpload",
             "glacier:AbortMultipartUpload",
             "glacier:CompleteMultipartUpload"
          ],
          "Resource": ["arn:${data.aws_partition.current.partition}:glacier:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:vaults/%[1]s"]
       }
    ]
}
EOF
}
`, rName)
}

func testAccGlacierVaultConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_glacier_vault" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccGlacierVaultConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_glacier_vault" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
