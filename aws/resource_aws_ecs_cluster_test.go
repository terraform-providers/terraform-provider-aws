package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ecs/finder"
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

	err = conn.ListClustersPages(&ecs.ListClustersInput{}, func(page *ecs.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, clusterARNPtr := range page.ClusterArns {
			clusterARN := aws.StringValue(clusterARNPtr)

			log.Printf("[INFO] Deleting ECS Cluster: %s", clusterARN)
			r := resourceAwsEcsCluster()
			d := r.Data(nil)
			d.SetId(clusterARN)
			err = r.Delete(d, client)
			if err != nil {
				log.Printf("[ERROR] Error deleting ECS Cluster (%s): %s", clusterARN, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping ECS Cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving ECS Clusters: %w", err)
	}

	return nil
}

func TestAccAWSEcsCluster_basic(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEcsCluster(), resourceName),
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
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
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

func TestAccAWSEcsCluster_SingleCapacityProvider(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	providerName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsClusterSingleCapacityProvider(rName, providerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
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

func TestAccAWSEcsCluster_CapacityProviders(t *testing.T) {
	var cluster ecs.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsClusterCapacityProviders(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccAWSEcsClusterCapacityProvidersReOrdered(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSEcsCluster_CapacityProvidersUpdate(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsClusterCapacityProvidersFargate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcsClusterCapacityProvidersFargateSpot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
				),
			},
			{
				Config: testAccAWSEcsClusterCapacityProvidersFargateBoth(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
				),
			},
		},
	})
}

func TestAccAWSEcsCluster_CapacityProvidersNoStrategy(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsClusterCapacityProvidersFargateNoStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcsClusterCapacityProvidersFargateSpotNoStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
				),
			},
		},
	})
}

func TestAccAWSEcsCluster_containerInsights(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ecs", fmt.Sprintf("cluster/%s", rName)),
				),
			},
			{
				Config: testAccAWSEcsClusterConfigContainerInsights(rName, "enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ecs", fmt.Sprintf("cluster/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "setting.*", map[string]string{
						"name":  "containerInsights",
						"value": "enabled",
					}),
				),
			},
			{
				Config: testAccAWSEcsClusterConfigContainerInsights(rName, "disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "setting.*", map[string]string{
						"name":  "containerInsights",
						"value": "disabled",
					}),
				),
			},
		},
	})
}

func TestAccAWSEcsCluster_configuration(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsClusterConfiguationConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.execute_command_configuration.0.kms_key_id", "aws_kms_key.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.0.logging", "OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.0.log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.0.log_configuration.0.cloud_watch_encryption_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.execute_command_configuration.0.log_configuration.0.cloud_watch_log_group_name", "aws_cloudwatch_log_group.test", "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcsClusterConfiguationConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.execute_command_configuration.0.kms_key_id", "aws_kms_key.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.0.logging", "OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.0.log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.0.log_configuration.0.cloud_watch_encryption_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.execute_command_configuration.0.log_configuration.0.cloud_watch_log_group_name", "aws_cloudwatch_log_group.test", "name"),
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

		out, err := finder.ClusterByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		for _, c := range out.Clusters {
			if aws.StringValue(c.ClusterArn) == rs.Primary.ID && aws.StringValue(c.Status) != "INACTIVE" {
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
		output, err := finder.ClusterByARN(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading ECS Cluster (%s): %w", rs.Primary.ID, err)
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

func testAccAWSEcsClusterCapacityProviderConfig(rName string) string {
	return testAccAWSEcsCapacityProviderConfigBase(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}
`, rName)
}

func testAccAWSEcsClusterSingleCapacityProvider(rName, providerName string) string {
	return testAccAWSEcsClusterCapacityProviderConfig(providerName) + fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = [aws_ecs_capacity_provider.test.name]

  default_capacity_provider_strategy {
    base              = 1
    capacity_provider = aws_ecs_capacity_provider.test.name
    weight            = 1
  }
}
`, rName)
}

func testAccAWSEcsClusterCapacityProviders(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE_SPOT", "FARGATE"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
    base              = 1
  }

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }
}
`, rName)
}

func testAccAWSEcsClusterCapacityProvidersReOrdered(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
    base              = 1
  }
}
`, rName)
}

func testAccAWSEcsClusterCapacityProvidersFargate(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    base              = 1
    capacity_provider = "FARGATE"
    weight            = 1
  }
}
`, rName)
}

func testAccAWSEcsClusterCapacityProvidersFargateSpot(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 1
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
  }
}
`, rName)
}

func testAccAWSEcsClusterCapacityProvidersFargateBoth(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 1
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
  }
}
`, rName)
}

func testAccAWSEcsClusterCapacityProvidersFargateNoStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE"]
}
`, rName)
}

func testAccAWSEcsClusterCapacityProvidersFargateSpotNoStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE_SPOT"]
}
`, rName)
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

func testAccAWSEcsClusterConfigContainerInsights(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
  setting {
    name  = "containerInsights"
    value = %[2]q
  }
}
`, rName, value)
}

func testAccAWSEcsClusterConfiguationConfig(rName string, enable bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q

  configuration {
    execute_command_configuration {
      kms_key_id = aws_kms_key.test.arn
      logging    = "OVERRIDE"

      log_configuration {
        cloud_watch_encryption_enabled = %[2]t
        cloud_watch_log_group_name     = aws_cloudwatch_log_group.test.name
      }
    }
  }
}
`, rName, enable)
}
