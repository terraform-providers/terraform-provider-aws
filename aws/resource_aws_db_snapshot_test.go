package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_db_snapshot", &resource.Sweeper{
		Name: "aws_db_snapshot",
		F:    testSweepDbSnapshots,
		Dependencies: []string{
			"aws_db_instance",
		},
	})
}

func testSweepDbSnapshots(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).rdsconn
	input := &rds.DescribeDBSnapshotsInput{}
	var sweeperErrs error

	err = conn.DescribeDBSnapshotsPages(input, func(out *rds.DescribeDBSnapshotsOutput, lastPage bool) bool {
		if out == nil {
			return !lastPage
		}

		for _, dbSnapshot := range out.DBSnapshots {
			if dbSnapshot == nil {
				continue
			}

			id := aws.StringValue(dbSnapshot.DBSnapshotIdentifier)
			input := &rds.DeleteDBSnapshotInput{
				DBSnapshotIdentifier: dbSnapshot.DBSnapshotIdentifier,
			}

			if strings.HasPrefix(id, "rds:") {
				log.Printf("[INFO] Skipping RDS Automated DB Snapshot: %s", id)
				continue
			}

			log.Printf("[INFO] Deleting RDS DB Snapshot: %s", id)
			_, err := conn.DeleteDBSnapshot(input)

			if isAWSErr(err, rds.ErrCodeDBSnapshotNotFoundFault, "") {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting RDS DB Snapshot (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}
		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Snapshot sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing RDS DB Snapshots: %s", err)
	}

	return sweeperErrs
}

func TestAccAWSDBSnapshot_basic(t *testing.T) {
	var v rds.DBSnapshot
	resourceName := "aws_db_snapshot.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDbSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "db_snapshot_arn", "rds", regexp.MustCompile(`snapshot:.+`)),
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

func TestAccAWSDBSnapshot_tags(t *testing.T) {
	var v rds.DBSnapshot
	resourceName := "aws_db_snapshot.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDbSnapshotConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbSnapshotExists(resourceName, &v),
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
				Config: testAccAwsDbSnapshotConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsDbSnapshotConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSDBSnapshot_disappears(t *testing.T) {
	var v rds.DBSnapshot
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDbSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbSnapshotExists(resourceName, &v),
					testAccCheckDbSnapshotDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDbSnapshotDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_snapshot" {
			continue
		}

		request := &rds.DescribeDBSnapshotsInput{
			DBSnapshotIdentifier: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBSnapshots(request)

		if isAWSErr(err, rds.ErrCodeDBSnapshotNotFoundFault, "") {
			continue
		}

		if err == nil {
			for _, dbSnapshot := range resp.DBSnapshots {
				if aws.StringValue(dbSnapshot.DBSnapshotIdentifier) == rs.Primary.ID {
					return fmt.Errorf("AWS DB Snapshot is still exist: %s", rs.Primary.ID)
				}
			}
		}

		return err
	}

	return nil
}

func testAccCheckDbSnapshotExists(n string, v *rds.DBSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

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
		return fmt.Errorf("Error finding RDS DB Snapshot %s", rs.Primary.ID)
	}
}

func testAccCheckDbSnapshotDisappears(snapshot *rds.DBSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		if _, err := conn.DeleteDBSnapshot(&rds.DeleteDBSnapshotInput{
			DBSnapshotIdentifier: snapshot.DBSnapshotIdentifier,
		}); err != nil {
			return err
		}

		return nil
	}
}

func testAccAwsDbSnapshotConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 10
  engine                  = "MySQL"
  engine_version          = "5.6.35"
  instance_class          = "db.t2.micro"
  name                    = "baz"
  identifier              = %[1]q
  password                = "barbarbarbar"
  username                = "foo"
  maintenance_window      = "Fri:09:00-Fri:09:30"
  backup_retention_period = 0
  parameter_group_name    = "default.mysql5.6"
  skip_final_snapshot     = true
}`, rName)
}

func testAccAwsDbSnapshotConfig(rName string) string {
	return testAccAwsDbSnapshotConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.id
  db_snapshot_identifier = %[1]q
}
`, rName)
}

func testAccAwsDbSnapshotConfigTags1(rName, tag1Key, tag1Value string) string {
	return testAccAwsDbSnapshotConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.id
  db_snapshot_identifier = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccAwsDbSnapshotConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return testAccAwsDbSnapshotConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.id
  db_snapshot_identifier = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
