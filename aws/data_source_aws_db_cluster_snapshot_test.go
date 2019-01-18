package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDbClusterSnapshotDataSource_DbClusterSnapshotIdentifier(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_db_cluster_snapshot.test"
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsDbClusterSnapshotDataSourceConfig_DbClusterSnapshotIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDbClusterSnapshotDataSourceExists(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "allocated_storage", resourceName, "allocated_storage"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_identifier", resourceName, "db_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine", resourceName, "engine"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", resourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "license_model", resourceName, "license_model"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttrSet(dataSourceName, "snapshot_create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_type", resourceName, "snapshot_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_db_cluster_snapshot_arn", resourceName, "source_db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_encrypted", resourceName, "storage_encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func TestAccAWSDbClusterSnapshotDataSource_DbClusterIdentifier(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_db_cluster_snapshot.test"
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsDbClusterSnapshotDataSourceConfig_DbClusterIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDbClusterSnapshotDataSourceExists(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "allocated_storage", resourceName, "allocated_storage"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_identifier", resourceName, "db_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine", resourceName, "engine"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", resourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "license_model", resourceName, "license_model"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttrSet(dataSourceName, "snapshot_create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_type", resourceName, "snapshot_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_db_cluster_snapshot_arn", resourceName, "source_db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_encrypted", resourceName, "storage_encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func TestAccAWSDbClusterSnapshotDataSource_MostRecent(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_db_cluster_snapshot.test"
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsDbClusterSnapshotDataSourceConfig_MostRecent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDbClusterSnapshotDataSourceExists(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
				),
			},
		},
	})
}

func testAccCheckAwsDbClusterSnapshotDataSourceExists(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Can't find data source: %s", dataSourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot data source ID not set")
		}
		return nil
	}
}

func testAccCheckAwsDbClusterSnapshotDataSourceConfig_DbClusterSnapshotIdentifier(rName string) string {
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

data "aws_db_cluster_snapshot" "test" {
  db_cluster_snapshot_identifier = "${aws_db_cluster_snapshot.test.id}"
}
`, rName, rName, rName, rName, rName)
}

func testAccCheckAwsDbClusterSnapshotDataSourceConfig_DbClusterIdentifier(rName string) string {
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

data "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier = "${aws_db_cluster_snapshot.test.db_cluster_identifier}"
}
`, rName, rName, rName, rName, rName)
}

func testAccCheckAwsDbClusterSnapshotDataSourceConfig_MostRecent(rName string) string {
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

resource "aws_db_cluster_snapshot" "incorrect" {
  db_cluster_identifier          = "${aws_rds_cluster.test.id}"
  db_cluster_snapshot_identifier = "%s-incorrect"
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_db_cluster_snapshot.incorrect.db_cluster_identifier}"
  db_cluster_snapshot_identifier = %q
}

data "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier = "${aws_db_cluster_snapshot.test.db_cluster_identifier}"
  most_recent           = true
}
`, rName, rName, rName, rName, rName, rName)
}
