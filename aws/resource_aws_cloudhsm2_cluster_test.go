package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCloudHsm2Cluster_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudHsm2ClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsm2Cluster(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsm2ClusterExists("aws_cloudhsm_v2_cluster.cluster"),
					resource.TestCheckResourceAttrSet("aws_cloudhsm_v2_cluster.cluster", "cluster_id"),
					resource.TestCheckResourceAttrSet("aws_cloudhsm_v2_cluster.cluster", "vpc_id"),
					resource.TestCheckResourceAttrSet("aws_cloudhsm_v2_cluster.cluster", "security_group_id"),
					resource.TestCheckResourceAttrSet("aws_cloudhsm_v2_cluster.cluster", "cluster_state"),
				),
			},
			{
				ResourceName:            "aws_cloudhsm_v2_cluster.cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster_certificates", "tags"},
			},
		},
	})
}

func testAccAWSCloudHsm2Cluster() string {
	return fmt.Sprintf(`
variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "cloudhsm2_test_vpc" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-aws_cloudhsm_v2_cluster-resource-basic"
  }
}

resource "aws_subnet" "cloudhsm2_test_subnets" {
  count                   = 2
  vpc_id                  = "${aws_vpc.cloudhsm2_test_vpc.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = false
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-aws_cloudhsm_v2_cluster-resource-basic"
  }
}

resource "aws_cloudhsm_v2_cluster" "cluster" {
  hsm_type = "hsm1.medium"  
  subnet_ids = ["${aws_subnet.cloudhsm2_test_subnets.*.id[0]}", "${aws_subnet.cloudhsm2_test_subnets.*.id[1]}"]
  tags = {
    Name = "tf-acc-aws_cloudhsm_v2_cluster-resource-basic-%d"
  }
}
`, acctest.RandInt())
}

func testAccCheckAWSCloudHsm2ClusterDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudhsm_v2_cluster" {
			continue
		}
		cluster, err := describeCloudHsm2Cluster(testAccProvider.Meta().(*AWSClient).cloudhsmv2conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if cluster != nil && aws.StringValue(cluster.State) != "DELETED" {
			return fmt.Errorf("CloudHSM cluster still exists %s", cluster)
		}
	}

	return nil
}

func testAccCheckAWSCloudHsm2ClusterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudhsmv2conn
		it, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		_, err := describeCloudHsm2Cluster(conn, it.Primary.ID)

		if err != nil {
			return fmt.Errorf("CloudHSM cluster not found: %s", err)
		}

		return nil
	}
}
