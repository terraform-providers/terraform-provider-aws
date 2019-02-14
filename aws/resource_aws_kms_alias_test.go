package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSKmsAlias_importBasic(t *testing.T) {
	resourceName := "aws_kms_alias.single"
	rInt := acctest.RandInt()
	kmsAliasTimestamp := time.Now().Format(time.RFC1123)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsSingleAlias(rInt, kmsAliasTimestamp),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKmsAlias_basic(t *testing.T) {
	rInt := acctest.RandInt()
	kmsAliasTimestamp := time.Now().Format(time.RFC1123)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsSingleAlias(rInt, kmsAliasTimestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists("aws_kms_alias.single"),
					resource.TestCheckResourceAttrSet("aws_kms_alias.single", "target_key_arn"),
				),
			},
			{
				Config: testAccAWSKmsSingleAlias_modified(rInt, kmsAliasTimestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists("aws_kms_alias.single"),
				),
			},
		},
	})
}

func TestAccAWSKmsAlias_name_prefix(t *testing.T) {
	rInt := acctest.RandInt()
	kmsAliasTimestamp := time.Now().Format(time.RFC1123)
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
		},
	})
}

func TestAccAWSKmsAlias_no_name(t *testing.T) {
	rInt := acctest.RandInt()
	kmsAliasTimestamp := time.Now().Format(time.RFC1123)
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
		},
	})
}

func TestAccAWSKmsAlias_multiple(t *testing.T) {
	rInt := acctest.RandInt()
	kmsAliasTimestamp := time.Now().Format(time.RFC1123)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsMultipleAliases(rInt, kmsAliasTimestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists("aws_kms_alias.one"),
					resource.TestCheckResourceAttrSet("aws_kms_alias.one", "target_key_arn"),
					testAccCheckAWSKmsAliasExists("aws_kms_alias.two"),
					resource.TestCheckResourceAttrSet("aws_kms_alias.two", "target_key_arn"),
				),
			},
		},
	})
}

func TestAccAWSKmsAlias_ArnDiffSuppress(t *testing.T) {
	rInt := acctest.RandInt()
	kmsAliasTimestamp := time.Now().Format(time.RFC1123)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsArnDiffSuppress(rInt, kmsAliasTimestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists("aws_kms_alias.bar"),
					resource.TestCheckResourceAttrSet("aws_kms_alias.bar", "target_key_arn"),
				),
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
resource "aws_kms_key" "one" {
    description = "Terraform acc test One %s"
    deletion_window_in_days = 7
}
resource "aws_kms_key" "two" {
    description = "Terraform acc test Two %s"
    deletion_window_in_days = 7
}

resource "aws_kms_alias" "name_prefix" {
	name_prefix = "alias/tf-acc-key-alias-%d"
	target_key_id = "${aws_kms_key.one.key_id}"
}

resource "aws_kms_alias" "nothing" {
	target_key_id = "${aws_kms_key.one.key_id}"
}

resource "aws_kms_alias" "single" {
    name = "alias/tf-acc-key-alias-%d"
    target_key_id = "${aws_kms_key.one.key_id}"
}`, timestamp, timestamp, rInt, rInt)
}

func testAccAWSKmsSingleAlias_modified(rInt int, timestamp string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "one" {
    description = "Terraform acc test One %s"
    deletion_window_in_days = 7
}
resource "aws_kms_key" "two" {
    description = "Terraform acc test Two %s"
    deletion_window_in_days = 7
}

resource "aws_kms_alias" "single" {
    name = "alias/tf-acc-key-alias-%d"
    target_key_id = "${aws_kms_key.two.key_id}"
}`, timestamp, timestamp, rInt)
}

func testAccAWSKmsMultipleAliases(rInt int, timestamp string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "single" {
    description = "Terraform acc test One %s"
    deletion_window_in_days = 7
}

resource "aws_kms_alias" "one" {
    name = "alias/tf-acc-alias-one-%d"
    target_key_id = "${aws_kms_key.single.key_id}"
}
resource "aws_kms_alias" "two" {
    name = "alias/tf-acc-alias-two-%d"
    target_key_id = "${aws_kms_key.single.key_id}"
}`, timestamp, rInt, rInt)
}

func testAccAWSKmsArnDiffSuppress(rInt int, timestamp string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "foo" {
    description = "Terraform acc test foo %s"
    deletion_window_in_days = 7
}

resource "aws_kms_alias" "bar" {
	name = "alias/tf-acc-key-alias-%d"
	target_key_id = "${aws_kms_key.foo.arn}"
}`, timestamp, rInt)
}
