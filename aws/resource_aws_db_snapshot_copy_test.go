package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSRdsSnapshotCopy_basic(t *testing.T) {
	var v rds.DBSnapshot
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRdsDbSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRdsDbSnapshotCopyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdsDbSnapshotCopyExists("aws_db_snapshot_copy.test", &v),
				),
			},
		},
	})
}

func TestAccAWSRdsDbSnapshotCopy_withRegions(t *testing.T) {
	var v rds.DBSnapshot
	rInt := acctest.RandInt()
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckRdsDbSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRdsDbSnapshotCopyConfigWithRegions(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdsDbSnapshotCopyExistsWithProviders("aws_db_snapshot_copy.test", &v, &providers),
				),
			},
		},
	})

}

func TestAccAWSRdsDbSnapshotCopy_disappears(t *testing.T) {
	var v rds.DBSnapshot
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRdsDbSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRdsDbSnapshotCopyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdsDbSnapshotCopyExists("aws_db_snapshot_copy.test", &v),
					testAccCheckRdsDbSnapshotCopyDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRdsDbSnapshotCopyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rds_db_snapshot_copy" {
			continue
		}

		resp, err := conn.DescribeDBSnapshots(&rds.DescribeDBSnapshotsInput{
			DBSnapshotIdentifier: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, "InvalidSnapshot.NotFound", "") {
			continue
		}

		if err == nil {
			for _, snapshot := range resp.DBSnapshots {
				if aws.StringValue(snapshot.DBSnapshotIdentifier) == rs.Primary.ID {
					return fmt.Errorf("RDS Snapshot still exists")
				}
			}
		}

		return err
	}

	return nil
}

func testAccCheckRdsDbSnapshotCopyDisappears(snapshot *rds.DBSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		_, err := conn.DeleteDBSnapshot(&rds.DeleteDBSnapshotInput{
			DBSnapshotIdentifier: snapshot.DBSnapshotIdentifier,
		})

		return err
	}
}

func testAccCheckRdsDbSnapshotCopyExists(n string, v *rds.DBSnapshot) resource.TestCheckFunc {
	providers := []*schema.Provider{testAccProvider}
	return testAccCheckRdsDbSnapshotCopyExistsWithProviders(n, v, &providers)
}

func testAccCheckRdsDbSnapshotCopyExistsWithProviders(n string, v *rds.DBSnapshot, providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		for _, provider := range *providers {
			// Ignore if Meta is empty, this can happen for validation providers
			if provider.Meta() == nil {
				continue
			}

			conn := provider.Meta().(*AWSClient).rdsconn

			request := &rds.DescribeDBSnapshotsInput{
				DBSnapshotIdentifier: aws.String(rs.Primary.ID),
			}

			response, err := conn.DescribeDBSnapshots(request)
			if err == nil {
				if response.DBSnapshots != nil && len(response.DBSnapshots) > 0 {
					*v = *response.DBSnapshots[0]
					return nil
				}
			}
		}
		return fmt.Errorf("Error finding RDS Snapshot %s", rs.Primary.ID)
	}
}

func testAccAwsRdsDbSnapshotCopyConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_db_instance" "bar" {
  allocated_storage = 10
  engine            = "MySQL"
  engine_version    = "5.6.35"
  instance_class    = "db.t2.micro"
  name              = "baz"
  password          = "barbarbarbar"
  username          = "foo"

  maintenance_window = "Fri:09:00-Fri:09:30"

  backup_retention_period = 0

  parameter_group_name = "default.mysql5.6"

  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = "${aws_db_instance.bar.id}"
  db_snapshot_identifier = "testsnapshot%d"
}

resource "aws_db_snapshot_copy" "test" {
        source_db_snapshot_identifier = "${aws_db_snapshot.test.db_snapshot_arn}"
        target_db_snapshot_identifier = "testsnapshot%d"
        source_region = "us-west-2"
}

`, rInt, rInt)
}

func testAccAwsRdsDbSnapshotCopyConfigWithRegions(rInt int) string {
	return fmt.Sprintf(`

provider "aws" {
	region = "us-west-2"
	alias  = "uswest2"
}

provider "aws" {
	region = "us-west-1"
	alias  = "uswest1"
}

resource "aws_db_instance" "bar" {
	provider          = "aws.uswest2"
	allocated_storage = 10
	engine = "MySQL"
	engine_version = "5.6.35"
	instance_class = "db.t2.micro"
	name = "baz"
	password = "barbarbarbar"
	username = "foo"

	maintenance_window = "Fri:09:00-Fri:09:30"

	backup_retention_period = 0

	parameter_group_name = "default.mysql5.6"

	skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  provider = "aws.uswest2"
  db_instance_identifier = "${aws_db_instance.bar.id}"
  db_snapshot_identifier = "testsnapshot%d"
}

resource "aws_db_snapshot_copy" "test" {
	provider           = "aws.uswest1"
	source_db_snapshot_identifier = "${aws_db_snapshot.test.db_snapshot_arn}"
	target_db_snapshot_identifier = "testsnapshot%d"
	source_region = "us-west-2"
}
`, rInt, rInt)
}
