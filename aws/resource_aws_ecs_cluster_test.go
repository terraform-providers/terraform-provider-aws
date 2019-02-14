package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ecs_cluster", &resource.Sweeper{
		Name: "aws_ecs_cluster",
		F:    testSweepEcsClusters,
		Dependencies: []string{
			"aws_ecs_service",
		},
	})
}

func testSweepEcsClusters(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ecsconn

	err = conn.ListClustersPages(&ecs.ListClustersInput{}, func(page *ecs.ListClustersOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, clusterARNPtr := range page.ClusterArns {
			input := &ecs.DeleteClusterInput{
				Cluster: clusterARNPtr,
			}
			clusterARN := aws.StringValue(clusterARNPtr)

			log.Printf("[INFO] Deleting ECS Cluster: %s", clusterARN)
			_, err = conn.DeleteCluster(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting ECS Cluster (%s): %s", clusterARN, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping ECS Cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving ECS Clusters: %s", err)
	}

	return nil
}

func TestAccAWSEcsCluster_basic(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ecs", fmt.Sprintf("cluster/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsCluster_disappears(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
					testAccCheckAWSEcsClusterDisappears(&cluster1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEcsCluster_Tags(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsClusterConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcsClusterConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEcsClusterConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSEcsClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecs_cluster" {
			continue
		}

		out, err := conn.DescribeClusters(&ecs.DescribeClustersInput{
			Clusters: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		for _, c := range out.Clusters {
			if *c.ClusterArn == rs.Primary.ID && *c.Status != "INACTIVE" {
				return fmt.Errorf("ECS cluster still exists:\n%s", c)
			}
		}
	}

	return nil
}

func testAccCheckAWSEcsClusterExists(resourceName string, cluster *ecs.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).ecsconn

		input := &ecs.DescribeClustersInput{
			Clusters: []*string{aws.String(rs.Primary.ID)},
			Include:  []*string{aws.String(ecs.ClusterFieldTags)},
		}

		output, err := conn.DescribeClusters(input)

		if err != nil {
			return fmt.Errorf("error reading ECS Cluster (%s): %s", rs.Primary.ID, err)
		}

		for _, c := range output.Clusters {
			if aws.StringValue(c.ClusterArn) == rs.Primary.ID && aws.StringValue(c.Status) != "INACTIVE" {
				*cluster = *c
				return nil
			}
		}

		return fmt.Errorf("ECS Cluster (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSEcsClusterDisappears(cluster *ecs.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ecsconn

		input := &ecs.DeleteClusterInput{
			Cluster: cluster.ClusterArn,
		}

		if _, err := conn.DeleteCluster(input); err != nil {
			return fmt.Errorf("error deleting ECS Cluster (%s): %s", aws.StringValue(cluster.ClusterArn), err)
		}

		return nil
	}
}

func testAccAWSEcsClusterConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %q
}
`, rName)
}

func testAccAWSEcsClusterConfigTags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %q

  tags = {
    %q = %q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccAWSEcsClusterConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %q

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
