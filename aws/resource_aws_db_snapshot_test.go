package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDBSnapshot_basic(t *testing.T) {
	var v rds.DBSnapshot
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDbSnapshotConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbSnapshotExists("aws_db_snapshot.test", &v),
				),
			},
		},
	})
}

func TestAccAWSDBSnapshot_tags(t *testing.T) {
	var v rds.DBSnapshot
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDbSnapshotConfigTags1(rInt, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbSnapshotExists("aws_db_snapshot.test", &v),
					resource.TestCheckResourceAttr("aws_db_snapshot.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_db_snapshot.test", "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAwsDbSnapshotConfigTags2(rInt, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbSnapshotExists("aws_db_snapshot.test", &v),
					resource.TestCheckResourceAttr("aws_db_snapshot.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_db_snapshot.test", "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr("aws_db_snapshot.test", "tags.key2", "value2"),
				),
			},
		},
	})
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

func testAccAwsDbSnapshotConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_db_instance" "bar" {
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
	db_instance_identifier = "${aws_db_instance.bar.id}"
	db_snapshot_identifier = "testsnapshot%d"
}`, rInt)
}

func testAccAwsDbSnapshotConfigTags1(rInt int, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_db_instance" "bar" {
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
	db_instance_identifier = "${aws_db_instance.bar.id}"
	db_snapshot_identifier = "testsnapshot%d"

	tags = {
		%q = %q
	  }
	}`, rInt, tag1Key, tag1Value)
}

func testAccAwsDbSnapshotConfigTags2(rInt int, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_db_instance" "bar" {
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
	db_instance_identifier = "${aws_db_instance.bar.id}"
	db_snapshot_identifier = "testsnapshot%d"

	tags = {
		%q = %q
		%q = %q
	  }
	}`, rInt, tag1Key, tag1Value, tag2Key, tag2Value)
}
