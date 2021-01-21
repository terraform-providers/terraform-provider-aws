package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSRDSCluster_basic(t *testing.T) {
	clusterName := fmt.Sprintf("testaccawsrdscluster-basic-%s", acctest.RandString(10))
	dataSourceName := "data.aws_rds_cluster.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRdsClusterConfigBasic(clusterName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "backtrack_window", resourceName, "backtrack_window"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_identifier", resourceName, "cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "database_name", resourceName, "database_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_parameter_group_name", resourceName, "db_cluster_parameter_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_subnet_group_name", resourceName, "db_subnet_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "hosted_zone_id", resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_username", resourceName, "master_username"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Environment", resourceName, "tags.Environment"),
				),
			},
		},
	})
}

func testAccDataSourceAwsRdsClusterConfigBasic(clusterName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = "%[1]s"
  database_name                   = "mydb"
  db_cluster_parameter_group_name = "default.aurora5.6"
  db_subnet_group_name            = aws_db_subnet_group.test.name
  master_password                 = "mustbeeightcharacters"
  master_username                 = "foo"
  skip_final_snapshot             = true

  tags = {
    Environment = "test"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-rds-cluster-data-source-basic"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-rds-cluster-data-source-basic"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-rds-cluster-data-source-basic"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "%[1]s"
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}

data "aws_rds_cluster" "test" {
  cluster_identifier = aws_rds_cluster.test.cluster_identifier
}
`, clusterName))
}
