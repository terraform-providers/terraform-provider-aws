package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_dax_cluster", &resource.Sweeper{
		Name: "aws_dax_cluster",
		F:    testSweepDAXClusters,
	})
}

func testSweepDAXClusters(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).daxconn

	resp, err := conn.DescribeClusters(&dax.DescribeClustersInput{})
	if err != nil {
		// GovCloud (with no DAX support) has an endpoint that responds with:
		// InvalidParameterValueException: Access Denied to API Version: DAX_V3
		if testSweepSkipSweepError(err) || isAWSErr(err, "InvalidParameterValueException", "Access Denied to API Version: DAX_V3") {
			log.Printf("[WARN] Skipping DAX Cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving DAX clusters: %s", err)
	}

	if len(resp.Clusters) == 0 {
		log.Print("[DEBUG] No DAX clusters to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d DAX clusters", len(resp.Clusters))

	for _, cluster := range resp.Clusters {
		if !strings.HasPrefix(*cluster.ClusterName, "tf-") {
			continue
		}

		log.Printf("[INFO] Deleting DAX cluster %s", *cluster.ClusterName)
		_, err := conn.DeleteCluster(&dax.DeleteClusterInput{
			ClusterName: cluster.ClusterName,
		})
		if err != nil {
			return fmt.Errorf("Error deleting DAX cluster %s: %s", *cluster.ClusterName, err)
		}
	}

	return nil
}

func TestAccAWSDAXCluster_basic(t *testing.T) {
	var dc dax.Cluster
	rString := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDAXClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDAXClusterConfig(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDAXClusterExists("aws_dax_cluster.test", &dc),
					resource.TestMatchResourceAttr(
						"aws_dax_cluster.test", "arn", regexp.MustCompile("^arn:aws:dax:[\\w-]+:\\d+:cache/")),
					resource.TestMatchResourceAttr(
						"aws_dax_cluster.test", "cluster_name", regexp.MustCompile("^tf-\\w+$")),
					resource.TestMatchResourceAttr(
						"aws_dax_cluster.test", "iam_role_arn", regexp.MustCompile("^arn:aws:iam::\\d+:role/")),
					resource.TestCheckResourceAttr(
						"aws_dax_cluster.test", "node_type", "dax.r3.large"),
					resource.TestCheckResourceAttr(
						"aws_dax_cluster.test", "replication_factor", "1"),
					resource.TestCheckResourceAttr(
						"aws_dax_cluster.test", "description", "test cluster"),
					resource.TestMatchResourceAttr(
						"aws_dax_cluster.test", "parameter_group_name", regexp.MustCompile("^default.dax")),
					resource.TestMatchResourceAttr(
						"aws_dax_cluster.test", "maintenance_window", regexp.MustCompile("^\\w{3}:\\d{2}:\\d{2}-\\w{3}:\\d{2}:\\d{2}$")),
					resource.TestCheckResourceAttr(
						"aws_dax_cluster.test", "subnet_group_name", "default"),
					resource.TestMatchResourceAttr(
						"aws_dax_cluster.test", "nodes.0.id", regexp.MustCompile("^tf-[\\w-]+$")),
					resource.TestMatchResourceAttr(
						"aws_dax_cluster.test", "configuration_endpoint", regexp.MustCompile(":\\d+$")),
					resource.TestCheckResourceAttrSet(
						"aws_dax_cluster.test", "cluster_address"),
					resource.TestMatchResourceAttr(
						"aws_dax_cluster.test", "port", regexp.MustCompile("^\\d+$")),
				),
			},
		},
	})
}

func TestAccAWSDAXCluster_resize(t *testing.T) {
	var dc dax.Cluster
	rString := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDAXClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDAXClusterConfigResize_singleNode(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDAXClusterExists("aws_dax_cluster.test", &dc),
					resource.TestCheckResourceAttr(
						"aws_dax_cluster.test", "replication_factor", "1"),
				),
			},
			{
				Config: testAccAWSDAXClusterConfigResize_multiNode(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDAXClusterExists("aws_dax_cluster.test", &dc),
					resource.TestCheckResourceAttr(
						"aws_dax_cluster.test", "replication_factor", "2"),
				),
			},
			{
				Config: testAccAWSDAXClusterConfigResize_singleNode(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDAXClusterExists("aws_dax_cluster.test", &dc),
					resource.TestCheckResourceAttr(
						"aws_dax_cluster.test", "replication_factor", "1"),
				),
			},
		},
	})
}

func testAccCheckAWSDAXClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).daxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dax_cluster" {
			continue
		}
		res, err := conn.DescribeClusters(&dax.DescribeClustersInput{
			ClusterNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			// Verify the error is what we want
			if isAWSErr(err, dax.ErrCodeClusterNotFoundFault, "") {
				continue
			}
			return err
		}
		if len(res.Clusters) > 0 {
			return fmt.Errorf("still exist.")
		}
	}
	return nil
}

func testAccCheckAWSDAXClusterExists(n string, v *dax.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DAX cluster ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).daxconn
		resp, err := conn.DescribeClusters(&dax.DescribeClustersInput{
			ClusterNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return fmt.Errorf("DAX error: %v", err)
		}

		for _, c := range resp.Clusters {
			if *c.ClusterName == rs.Primary.ID {
				*v = *c
			}
		}

		return nil
	}
}

var baseConfig = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
       {
            "Effect": "Allow",
            "Principal": {
                "Service": "dax.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = "${aws_iam_role.test.id}"

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "dynamodb:*",
            "Effect": "Allow",
            "Resource": "*"
        }
    ]
}
EOF
}
`

func testAccAWSDAXClusterConfig(rString string) string {
	return fmt.Sprintf(`%s
		resource "aws_dax_cluster" "test" {
 		  cluster_name       = "tf-%s"
		  iam_role_arn       = "${aws_iam_role.test.arn}"
		  node_type          = "dax.r3.large"
		  replication_factor = 1
		  description        = "test cluster"

		  tags {
		    foo = "bar"
		  }
		}
		`, baseConfig, rString)
}

func testAccAWSDAXClusterConfigResize_singleNode(rString string) string {
	return fmt.Sprintf(`%s
		resource "aws_dax_cluster" "test" {
		  cluster_name       = "tf-%s"
		  iam_role_arn       = "${aws_iam_role.test.arn}"
		  node_type          = "dax.r3.large"
		  replication_factor = 1
		}
		`, baseConfig, rString)
}

func testAccAWSDAXClusterConfigResize_multiNode(rString string) string {
	return fmt.Sprintf(`%s
		resource "aws_dax_cluster" "test" {
		  cluster_name       = "tf-%s"
		  iam_role_arn       = "${aws_iam_role.test.arn}"
		  node_type          = "dax.r3.large"
		  replication_factor = 2
		}
		`, baseConfig, rString)
}
