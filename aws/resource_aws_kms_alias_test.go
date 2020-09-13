package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSKmsAlias_basic(t *testing.T) {
	rInt := acctest.RandInt()
	kmsAliasTimestamp := time.Now().Format(time.RFC1123)
	resourceName := "aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsSingleAlias(rInt, kmsAliasTimestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "target_key_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSKmsSingleAlias_modified(rInt, kmsAliasTimestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists(resourceName),
				),
			},
		},
	})
}

func TestAccAWSKmsAlias_name_prefix(t *testing.T) {
	rInt := acctest.RandInt()
	kmsAliasTimestamp := time.Now().Format(time.RFC1123)
	resourceName := "aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsSingleAlias(rInt, kmsAliasTimestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists("aws_kms_alias.name_prefix"),
					resource.TestCheckResourceAttrSet("aws_kms_alias.name_prefix", "target_key_arn"),
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

func TestAccAWSKmsAlias_no_name(t *testing.T) {
	rInt := acctest.RandInt()
	kmsAliasTimestamp := time.Now().Format(time.RFC1123)
	resourceName := "aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsSingleAlias(rInt, kmsAliasTimestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists("aws_kms_alias.nothing"),
					resource.TestCheckResourceAttrSet("aws_kms_alias.nothing", "target_key_arn"),
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

func TestAccAWSKmsAlias_multiple(t *testing.T) {
	rInt := acctest.RandInt()
	kmsAliasTimestamp := time.Now().Format(time.RFC1123)
	resourceName := "aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsMultipleAliases(rInt, kmsAliasTimestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists("aws_kms_alias.test"),
					resource.TestCheckResourceAttrSet("aws_kms_alias.test", "target_key_arn"),
					testAccCheckAWSKmsAliasExists("aws_kms_alias.test2"),
					resource.TestCheckResourceAttrSet("aws_kms_alias.test2", "target_key_arn"),
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

func TestAccAWSKmsAlias_ArnDiffSuppress(t *testing.T) {
	rInt := acctest.RandInt()
	kmsAliasTimestamp := time.Now().Format(time.RFC1123)
	resourceName := "aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsArnDiffSuppress(rInt, kmsAliasTimestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists("aws_kms_alias.test"),
					resource.TestCheckResourceAttrSet("aws_kms_alias.test", "target_key_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
				Config:             testAccAWSKmsArnDiffSuppress(rInt, kmsAliasTimestamp),
			},
		},
	})
}

func testAccCheckAWSKmsAliasDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).kmsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kms_alias" {
			continue
		}

		entry, err := findKmsAliasByName(conn, rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if entry != nil {
			return fmt.Errorf("KMS alias still exists:\n%#v", entry)
		}

		return nil
	}

	return nil
}

func testAccCheckAWSKmsAliasExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAWSKmsSingleAlias(rInt int, timestamp string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Terraform acc test One %s"
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  description             = "Terraform acc test Two %s"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "name_prefix" {
  name_prefix   = "alias/tf-acc-key-alias-%d"
  target_key_id = aws_kms_key.test.key_id
}

resource "aws_kms_alias" "nothing" {
  target_key_id = aws_kms_key.test.key_id
}

resource "aws_kms_alias" "test" {
  name          = "alias/tf-acc-key-alias-%d"
  target_key_id = aws_kms_key.test.key_id
}
`, timestamp, timestamp, rInt, rInt)
}

func testAccAWSKmsSingleAlias_modified(rInt int, timestamp string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Terraform acc test One %s"
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  description             = "Terraform acc test Two %s"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/tf-acc-key-alias-%d"
  target_key_id = aws_kms_key.test2.key_id
}
`, timestamp, timestamp, rInt)
}

func testAccAWSKmsMultipleAliases(rInt int, timestamp string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Terraform acc test One %s"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/tf-acc-alias-test-%d"
  target_key_id = aws_kms_key.test.key_id
}

resource "aws_kms_alias" "test2" {
  name          = "alias/tf-acc-alias-test2-%d"
  target_key_id = aws_kms_key.test.key_id
}
`, timestamp, rInt, rInt)
}

func testAccAWSKmsArnDiffSuppress(rInt int, timestamp string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Terraform acc test test %s"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/tf-acc-key-alias-%d"
  target_key_id = aws_kms_key.test.arn
}
`, timestamp, rInt)
}
