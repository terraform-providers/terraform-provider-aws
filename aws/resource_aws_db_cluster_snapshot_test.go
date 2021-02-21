package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_db_cluster_snapshot", &resource.Sweeper{
		Name: "aws_db_cluster_snapshot",
		F:    testSweepDbClusterSnapshots,
	})
}

func testSweepDbClusterSnapshots(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).rdsconn
	input := &rds.DescribeDBClusterSnapshotsInput{
		// "InvalidDBClusterSnapshotStateFault: Only manual snapshots may be deleted."
		Filters: []*rds.Filter{{
			Name:   aws.String("snapshot-type"),
			Values: aws.StringSlice([]string{"manual"}),
		}},
	}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.DescribeDBClusterSnapshots(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Cluster Snapshots sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving RDS DB Cluster Snapshots: %w", err))
			return sweeperErrs
		}

		for _, dbClusterSnapshot := range output.DBClusterSnapshots {
			id := aws.StringValue(dbClusterSnapshot.DBClusterSnapshotIdentifier)

			log.Printf("[INFO] Deleting RDS DB Cluster Snapshot: %s", id)
			_, err := conn.DeleteDBClusterSnapshot(&rds.DeleteDBClusterSnapshotInput{
				DBClusterSnapshotIdentifier: aws.String(id),
			})
			if isAWSErr(err, rds.ErrCodeDBClusterSnapshotNotFoundFault, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting RDS DB Cluster Snapshot (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.Marker) == "" {
			break
		}
		input.Marker = output.Marker
	}

	return sweeperErrs.ErrorOrNil()
}

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
					testAccMatchResourceAttrRegionalARN(resourceName, "db_cluster_snapshot_arn", "rds", regexp.MustCompile(fmt.Sprintf("cluster-snapshot:%s$", rName))),
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
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSDBClusterSnapshot_Tags(t *testing.T) {
	var dbClusterSnapshot rds.DBClusterSnapshot
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbClusterSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDbClusterSnapshotConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbClusterSnapshotExists(resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
			{
				Config: testAccAwsDbClusterSnapshotConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbClusterSnapshotExists(resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsDbClusterSnapshotConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbClusterSnapshotExists(resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q
}
`, rName)
}

func testAccAwsDbClusterSnapshotConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.*.id[0], aws_subnet.test.*.id[1]]
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAwsDbClusterSnapshotConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.*.id[0], aws_subnet.test.*.id[1]]
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
