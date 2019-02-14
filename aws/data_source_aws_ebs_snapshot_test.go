package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEbsSnapshotDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ebs_snapshot.test"
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEbsSnapshotDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEbsSnapshotDataSourceID(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "encrypted", resourceName, "encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_alias", resourceName, "owner_alias"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_id", resourceName, "volume_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_size", resourceName, "volume_size"),
				),
			},
		},
	})
}

func TestAccAWSEbsSnapshotDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ebs_snapshot.test"
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEbsSnapshotDataSourceConfigFilter,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEbsSnapshotDataSourceID(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
				),
			},
		},
	})
}

func TestAccAWSEbsSnapshotDataSource_MostRecent(t *testing.T) {
	dataSourceName := "data.aws_ebs_snapshot.test"
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEbsSnapshotDataSourceConfigMostRecent,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEbsSnapshotDataSourceID(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckAwsEbsSnapshotDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find snapshot data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot data source ID not set")
		}
		return nil
	}
}

const testAccCheckAwsEbsSnapshotDataSourceConfig = `
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  type = "gp2"
  size = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"
}

data "aws_ebs_snapshot" "test" {
  snapshot_ids = ["${aws_ebs_snapshot.test.id}"]
}
`

const testAccCheckAwsEbsSnapshotDataSourceConfigFilter = `
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  type = "gp2"
  size = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"
}

data "aws_ebs_snapshot" "test" {
  filter {
    name = "snapshot-id"
    values = ["${aws_ebs_snapshot.test.id}"]
  }
}
`

const testAccCheckAwsEbsSnapshotDataSourceConfigMostRecent = `
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  type = "gp2"
  size = 1
}

resource "aws_ebs_snapshot" "incorrect" {
  volume_id = "${aws_ebs_volume.test.id}"

  tags = {
    Name = "tf-acc-test-ec2-ebs-snapshot-data-source-most-recent"
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_snapshot.incorrect.volume_id}"

  tags = {
    Name = "tf-acc-test-ec2-ebs-snapshot-data-source-most-recent"
  }
}

data "aws_ebs_snapshot" "test" {
  most_recent = true

  filter {
    name   = "tag:Name"
    values = ["${aws_ebs_snapshot.test.tags.Name}"]
  }
}
`
