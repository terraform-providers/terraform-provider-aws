package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsRdsCluster_basic(t *testing.T) {
	clusterName := fmt.Sprintf("testaccawsrdscluster-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRdsClusterConfigBasic(clusterName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_rds_cluster.rds_cluster_test", "cluster_identifier", clusterName),
					resource.TestCheckResourceAttr("data.aws_rds_cluster.rds_cluster_test", "database_name", "mydb"),
					resource.TestCheckResourceAttr("data.aws_rds_cluster.rds_cluster_test", "db_cluster_parameter_group_name", "default.aurora5.6"),
					resource.TestCheckResourceAttr("data.aws_rds_cluster.rds_cluster_test", "db_subnet_group_name", clusterName),
					resource.TestCheckResourceAttr("data.aws_rds_cluster.rds_cluster_test", "master_username", "foo"),
					resource.TestCheckResourceAttr("data.aws_rds_cluster.rds_cluster_test", "tags.%", "1"),
					resource.TestCheckResourceAttr("data.aws_rds_cluster.rds_cluster_test", "tags.Environment", "test"),
				),
			},
		},
	})
}

func testAccDataSourceAwsRdsClusterConfigBasic(clusterName string) string {
	return fmt.Sprintf(`resource "aws_rds_cluster" "rds_cluster_test" {
	cluster_identifier = "%s"
	database_name = "mydb"
	db_cluster_parameter_group_name = "default.aurora5.6"
	db_subnet_group_name = "${aws_db_subnet_group.test.name}"
	master_password = "mustbeeightcharacters"
	master_username = "foo"
	skip_final_snapshot = true
	tags {
		Environment = "test"
	}
}
resource "aws_vpc" "test" {
	cidr_block = "10.0.0.0/16"
	tags {
	  Name = "terraform-testacc-rds-cluster-data-source-basic"
	}
}
  
resource "aws_subnet" "a" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.0.0.0/24"
	availability_zone = "us-west-2a"
	tags {
		Name = "tf-acc-rds-cluster-data-source-basic"
	}
}
  
resource "aws_subnet" "b" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.0.1.0/24"
	availability_zone = "us-west-2b"
	tags {
		Name = "tf-acc-rds-cluster-data-source-basic"
	}
}
  
resource "aws_db_subnet_group" "test" {
	name = "%s"
	subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}

data "aws_rds_cluster" "rds_cluster_test" {
	cluster_identifier = "${aws_rds_cluster.rds_cluster_test.cluster_identifier}"
}`, clusterName, clusterName)
}
