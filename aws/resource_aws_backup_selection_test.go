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

func TestAccAwsBackupSelection_basic(t *testing.T) {
	var selection1 backup.GetBackupSelectionOutput
	resourceName := "aws_backup_selection.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupSelectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupSelectionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionExists(resourceName, &selection1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSBackupSelectionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsBackupSelection_disappears(t *testing.T) {
	var selection1 backup.GetBackupSelectionOutput
	resourceName := "aws_backup_selection.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupSelectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupSelectionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionExists(resourceName, &selection1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsBackupSelection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsBackupSelection_disappears_BackupPlan(t *testing.T) {
	var selection1 backup.GetBackupSelectionOutput
	resourceName := "aws_backup_selection.test"
	backupPlanResourceName := "aws_backup_plan.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupSelectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupSelectionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionExists(resourceName, &selection1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsBackupSelection(), resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsBackupPlan(), backupPlanResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsBackupSelection_withTags(t *testing.T) {
	var selection1 backup.GetBackupSelectionOutput
	resourceName := "aws_backup_selection.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupSelectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupSelectionConfigWithTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionExists(resourceName, &selection1),
					resource.TestCheckResourceAttr(resourceName, "selection_tag.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSBackupSelectionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsBackupSelection_withResources(t *testing.T) {
	var selection1 backup.GetBackupSelectionOutput
	resourceName := "aws_backup_selection.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupSelectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupSelectionConfigWithResources(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionExists(resourceName, &selection1),
					resource.TestCheckResourceAttr(resourceName, "resources.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSBackupSelectionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsBackupSelection_updateTag(t *testing.T) {
	var selection1, selection2 backup.GetBackupSelectionOutput
	resourceName := "aws_backup_selection.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupSelectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupSelectionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionExists(resourceName, &selection1),
				),
			},
			{
				Config: testAccBackupSelectionConfigUpdateTag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionExists(resourceName, &selection2),
					testAccCheckAwsBackupSelectionRecreated(t, &selection1, &selection2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSBackupSelectionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsBackupSelectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).backupconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_selection" {
			continue
		}

		input := &backup.GetBackupSelectionInput{
			BackupPlanId: aws.String(rs.Primary.Attributes["plan_id"]),
			SelectionId:  aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetBackupSelection(input)

		if err == nil {
			if *resp.SelectionId == rs.Primary.ID {
				return fmt.Errorf("Selection '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsBackupSelectionExists(name string, selection *backup.GetBackupSelectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s, %v", name, s.RootModule().Resources)
		}

		conn := testAccProvider.Meta().(*AWSClient).backupconn

		input := &backup.GetBackupSelectionInput{
			BackupPlanId: aws.String(rs.Primary.Attributes["plan_id"]),
			SelectionId:  aws.String(rs.Primary.ID),
		}

		output, err := conn.GetBackupSelection(input)

		if err != nil {
			return err
		}

		*selection = *output

		return nil
	}
}

func testAccCheckAwsBackupSelectionRecreated(t *testing.T,
	before, after *backup.GetBackupSelectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.SelectionId == *after.SelectionId {
			t.Fatalf("Expected change of Backup Selection IDs, but both were %s", *before.SelectionId)
		}
		return nil
	}
}

func testAccAWSBackupSelectionImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		id := fmt.Sprintf("%s|%s",
			rs.Primary.Attributes["plan_id"],
			rs.Primary.ID)

		return id, nil
	}
}

func testAccBackupSelectionConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {
}

data "aws_partition" "current" {
}

data "aws_region" "current" {
}

resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"
  }
}
`, rName)
}

func testAccBackupSelectionConfigBasic(rName string) string {
	return composeConfig(
		testAccBackupSelectionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_backup_selection" "test" {
  plan_id = aws_backup_plan.test.id

  name         = %[1]q
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo"
    value = "bar"
  }

  resources = [
    "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/"
  ]
}
`, rName))
}

func testAccBackupSelectionConfigWithTags(rName string) string {
	return composeConfig(
		testAccBackupSelectionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_backup_selection" "test" {
  plan_id = aws_backup_plan.test.id

  name         = %[1]q
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo"
    value = "bar"
  }

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "boo"
    value = "far"
  }

  resources = [
    "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/"
  ]
}
`, rName))
}

func testAccBackupSelectionConfigWithResources(rName string) string {
	return composeConfig(
		testAccBackupSelectionConfigBase(rName),
		fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_backup_selection" "test" {
  plan_id = aws_backup_plan.test.id

  name         = %[1]q
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo"
    value = "bar"
  }

  resources = aws_ebs_volume.test[*].arn
}
`, rName))
}

func testAccBackupSelectionConfigUpdateTag(rName string) string {
	return composeConfig(
		testAccBackupSelectionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_backup_selection" "test" {
  plan_id = aws_backup_plan.test.id

  name         = %[1]q
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo2"
    value = "bar2"
  }

  resources = [
    "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/"
  ]
}
`, rName))
}
