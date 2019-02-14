package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceCloudHsm2Cluster_basic(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_cluster.cluster"
	dataSourceName := "data.aws_cloudhsm_v2_cluster.default"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudHsm2ClusterDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cluster_state", "UNINITIALIZED"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_id", resourceName, "cluster_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_state", resourceName, "cluster_state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_id", resourceName, "security_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

var testAccCheckCloudHsm2ClusterDataSourceConfig = fmt.Sprintf(`
variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "cloudhsm2_test_vpc" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-aws_cloudhsm_v2_cluster-data-source-basic"
  }
}

resource "aws_subnet" "cloudhsm2_test_subnets" {
  count                   = 2
  vpc_id                  = "${aws_vpc.cloudhsm2_test_vpc.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = false
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-aws_cloudhsm_v2_cluster-data-source-basic"
  }
}

resource "aws_cloudhsm_v2_cluster" "cluster" {
  hsm_type = "hsm1.medium"  
  subnet_ids = ["${aws_subnet.cloudhsm2_test_subnets.0.id}", "${aws_subnet.cloudhsm2_test_subnets.1.id}"]
  tags = {
    Name = "tf-acc-aws_cloudhsm_v2_cluster-data-source-basic-%d"
  }
}

data "aws_cloudhsm_v2_cluster" "default" {
  cluster_id = "${aws_cloudhsm_v2_cluster.cluster.cluster_id}"
}
`, acctest.RandInt())
