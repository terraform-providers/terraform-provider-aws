package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudhsm_v2_cluster", &resource.Sweeper{
		Name: "aws_cloudhsm_v2_cluster",
		F:    testSweepCloudhsmv2Clusters,
	})
}

func testSweepCloudhsmv2Clusters(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).cloudhsmv2conn

	input := &cloudhsmv2.DescribeClustersInput{}

	err = conn.DescribeClustersPages(input, func(page *cloudhsmv2.DescribeClustersOutput, lastPage bool) bool {
		for _, cluster := range page.Clusters {
			clusterID := aws.StringValue(cluster.ClusterId)
			input := &cloudhsmv2.DeleteClusterInput{
				ClusterId: cluster.ClusterId,
			}

			for _, hsm := range cluster.Hsms {
				hsmID := aws.StringValue(hsm.HsmId)
				input := &cloudhsmv2.DeleteHsmInput{
					ClusterId: cluster.ClusterId,
					HsmId:     hsm.HsmId,
				}

				log.Printf("[INFO] Deleting CloudHSMv2 Cluster (%s) HSM: %s", clusterID, hsmID)
				_, err := conn.DeleteHsm(input)

				if err != nil {
					log.Printf("[ERROR] Error deleting CloudHSMv2 Cluster (%s) HSM (%s): %s", clusterID, hsmID, err)
					continue
				}

				if err := waitForCloudhsmv2HsmDeletion(conn, hsmID, 120*time.Minute); err != nil {
					log.Printf("[ERROR] Error waiting for CloudHSMv2 Cluster (%s) HSM (%s) deletion: %s", clusterID, hsmID, err)
				}
			}

			log.Printf("[INFO] Deleting CloudHSMv2 Cluster: %s", clusterID)
			_, err := conn.DeleteCluster(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting CloudHSMv2 Cluster (%s): %s", clusterID, err)
				continue
			}

			if err := waitForCloudhsmv2ClusterDeletion(conn, clusterID, 120*time.Minute); err != nil {
				log.Printf("[ERROR] Error waiting for CloudHSMv2 Cluster (%s) deletion: %s", clusterID, err)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudHSMv2 Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing CloudHSMv2 Clusters: %s", err)
	}

	return nil
}

func TestAccAWSCloudHsmV2Cluster_basic(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_cluster.cluster"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudHsmV2ClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2ClusterConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_id"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_state"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster_certificates"},
			},
		},
	})
}

func TestAccAWSCloudHsmV2Cluster_Tags(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudHsmV2ClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2ClusterConfigTags2("key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster_certificates"},
			},
			{
				Config: testAccAWSCloudHsmV2ClusterConfigTags1("key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
				),
			},
			{
				Config: testAccAWSCloudHsmV2ClusterConfigTags2("key1", "value1updated", "key3", "value3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key3", "value3"),
				),
			},
		},
	})
}

func testAccAWSCloudHsmV2ClusterConfigBase() string {
	return `
variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "cloudhsm_v2_test_vpc" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-aws_cloudhsm_v2_cluster-resource-basic"
  }
}

resource "aws_subnet" "cloudhsm_v2_test_subnets" {
  count                   = 2
  vpc_id                  = aws_vpc.cloudhsm_v2_test_vpc.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = false
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-aws_cloudhsm_v2_cluster-resource-basic"
  }
}
`
}

func testAccAWSCloudHsmV2ClusterConfig() string {
	return testAccAWSCloudHsmV2ClusterConfigBase() + `
resource "aws_cloudhsm_v2_cluster" "cluster" {
  hsm_type   = "hsm1.medium"
  subnet_ids = [aws_subnet.cloudhsm_v2_test_subnets.*.id[0], aws_subnet.cloudhsm_v2_test_subnets.*.id[1]]
}
`
}

func testAccAWSCloudHsmV2ClusterConfigTags1(tagKey1, tagValue1 string) string {
	return testAccAWSCloudHsmV2ClusterConfigBase() + fmt.Sprintf(`
resource "aws_cloudhsm_v2_cluster" "test" {
  hsm_type   = "hsm1.medium"
  subnet_ids = [aws_subnet.cloudhsm_v2_test_subnets.*.id[0], aws_subnet.cloudhsm_v2_test_subnets.*.id[1]]

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccAWSCloudHsmV2ClusterConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSCloudHsmV2ClusterConfigBase() + fmt.Sprintf(`
resource "aws_cloudhsm_v2_cluster" "test" {
  hsm_type   = "hsm1.medium"
  subnet_ids = [aws_subnet.cloudhsm_v2_test_subnets.*.id[0], aws_subnet.cloudhsm_v2_test_subnets.*.id[1]]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCheckAWSCloudHsmV2ClusterDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudhsm_v2_cluster" {
			continue
		}
		cluster, err := describeCloudHsmV2Cluster(testAccProvider.Meta().(*AWSClient).cloudhsmv2conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if cluster != nil && aws.StringValue(cluster.State) != cloudhsmv2.ClusterStateDeleted {
			return fmt.Errorf("CloudHSM cluster still exists %s", cluster)
		}
	}

	return nil
}

func testAccCheckAWSCloudHsmV2ClusterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudhsmv2conn
		it, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		_, err := describeCloudHsmV2Cluster(conn, it.Primary.ID)

		if err != nil {
			return fmt.Errorf("CloudHSM cluster not found: %s", err)
		}

		return nil
	}
}
