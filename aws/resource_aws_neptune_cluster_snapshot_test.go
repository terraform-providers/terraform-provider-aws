package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSNeptuneClusterSnapshot_basic(t *testing.T) {
	var dbClusterSnapshot neptune.DBClusterSnapshot
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_neptune_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNeptuneClusterSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNeptuneClusterSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNeptuneClusterSnapshotExists(resourceName, &dbClusterSnapshot),
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

func testAccCheckNeptuneClusterSnapshotDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).neptuneconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_cluster_snapshot" {
			continue
		}

		input := &neptune.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeDBClusterSnapshots(input)
		if err != nil {
			if isAWSErr(err, neptune.ErrCodeDBClusterSnapshotNotFoundFault, "") {
				continue
			}
			return err
		}

		if output != nil && len(output.DBClusterSnapshots) > 0 && output.DBClusterSnapshots[0] != nil && aws.StringValue(output.DBClusterSnapshots[0].DBClusterSnapshotIdentifier) == rs.Primary.ID {
			return fmt.Errorf("Neptune DB Cluster Snapshot %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckNeptuneClusterSnapshotExists(resourceName string, dbClusterSnapshot *neptune.DBClusterSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set for %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).neptuneconn

		input := &neptune.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeDBClusterSnapshots(input)
		if err != nil {
			return err
		}

		if output == nil || len(output.DBClusterSnapshots) == 0 || output.DBClusterSnapshots[0] == nil || aws.StringValue(output.DBClusterSnapshots[0].DBClusterSnapshotIdentifier) != rs.Primary.ID {
			return fmt.Errorf("Neptune DB Cluster Snapshot %q not found", rs.Primary.ID)
		}

		*dbClusterSnapshot = *output.DBClusterSnapshots[0]

		return nil
	}
}

func testAccAwsNeptuneClusterSnapshotConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier   = %q
  skip_final_snapshot  = true
}

resource "aws_neptune_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_neptune_cluster.test.id}"
  db_cluster_snapshot_identifier = %q
}
`, rName, rName)
}
