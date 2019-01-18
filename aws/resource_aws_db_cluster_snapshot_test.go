package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDBClusterSnapshot_basic(t *testing.T) {
	var dbClusterSnapshot rds.DBClusterSnapshot
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbClusterSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDbClusterSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbClusterSnapshotExists(resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttrSet(resourceName, "allocated_storage"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zones.#"),
					resource.TestMatchResourceAttr(resourceName, "db_cluster_snapshot_arn", regexp.MustCompile(`^arn:[^:]+:rds:[^:]+:\d{12}:cluster-snapshot:.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, "license_model"),
					resource.TestCheckResourceAttrSet(resourceName, "port"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_type", "manual"),
					resource.TestCheckResourceAttr(resourceName, "source_db_cluster_snapshot_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "status", "available"),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestMatchResourceAttr(resourceName, "vpc_id", regexp.MustCompile(`^vpc-.+`)),
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

func testAccCheckDbClusterSnapshotDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_cluster_snapshot" {
			continue
		}

		input := &rds.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeDBClusterSnapshots(input)
		if err != nil {
			if isAWSErr(err, rds.ErrCodeDBClusterSnapshotNotFoundFault, "") {
				continue
			}
			return err
		}

		if output != nil && len(output.DBClusterSnapshots) > 0 && output.DBClusterSnapshots[0] != nil && aws.StringValue(output.DBClusterSnapshots[0].DBClusterSnapshotIdentifier) == rs.Primary.ID {
			return fmt.Errorf("RDS DB Cluster Snapshot %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckDbClusterSnapshotExists(resourceName string, dbClusterSnapshot *rds.DBClusterSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set for %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		request := &rds.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(rs.Primary.ID),
		}

		response, err := conn.DescribeDBClusterSnapshots(request)
		if err != nil {
			return err
		}

		if response == nil || len(response.DBClusterSnapshots) == 0 || response.DBClusterSnapshots[0] == nil || aws.StringValue(response.DBClusterSnapshots[0].DBClusterSnapshotIdentifier) != rs.Primary.ID {
			return fmt.Errorf("RDS DB Cluster Snapshot %q not found", rs.Primary.ID)
		}

		*dbClusterSnapshot = *response.DBClusterSnapshots[0]

		return nil
	}
}

func testAccAwsDbClusterSnapshotConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
   Name = %q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = %q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %q
  subnet_ids = ["${aws_subnet.test.*.id}"]
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = %q
  db_subnet_group_name = "${aws_db_subnet_group.test.name}"
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_rds_cluster.test.id}"
  db_cluster_snapshot_identifier = %q
}
`, rName, rName, rName, rName, rName)
}
