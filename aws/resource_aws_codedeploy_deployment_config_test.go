package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCodeDeployDeploymentConfig_basic(t *testing.T) {
	var config1 codedeploy.DeploymentConfigInfo
	resourceName := "aws_codedeploy_deployment_config.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentConfigFleet(rName, 75),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists(resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "deployment_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentConfig_fleetPercent(t *testing.T) {
	var config1, config2 codedeploy.DeploymentConfigInfo
	resourceName := "aws_codedeploy_deployment_config.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentConfigFleet(rName, 75),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists(resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.type", "FLEET_PERCENT"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.value", "75"),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", "0"),
				),
			},
			{
				Config: testAccAWSCodeDeployDeploymentConfigFleet(rName, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists(resourceName, &config2),
					testAccCheckAWSCodeDeployDeploymentConfigRecreated(&config1, &config2),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.type", "FLEET_PERCENT"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.value", "50"),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentConfig_hostCount(t *testing.T) {
	var config1, config2 codedeploy.DeploymentConfigInfo
	resourceName := "aws_codedeploy_deployment_config.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentConfigHostCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists(resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.type", "HOST_COUNT"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.value", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", "0"),
				),
			},
			{
				Config: testAccAWSCodeDeployDeploymentConfigHostCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists(resourceName, &config2),
					testAccCheckAWSCodeDeployDeploymentConfigRecreated(&config1, &config2),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.type", "HOST_COUNT"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.value", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentConfig_trafficCanary(t *testing.T) {
	var config1, config2 codedeploy.DeploymentConfigInfo
	resourceName := "aws_codedeploy_deployment_config.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentConfigTrafficCanary(rName, 10, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists(resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.type", "TimeBasedCanary"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.0.interval", "10"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.0.percentage", "50"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", "0"),
				),
			},
			{
				Config: testAccAWSCodeDeployDeploymentConfigTrafficCanary(rName, 3, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists(resourceName, &config2),
					testAccCheckAWSCodeDeployDeploymentConfigRecreated(&config1, &config2),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.type", "TimeBasedCanary"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.0.interval", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.0.percentage", "10"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentConfig_trafficLinear(t *testing.T) {
	var config1, config2 codedeploy.DeploymentConfigInfo
	resourceName := "aws_codedeploy_deployment_config.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentConfigTrafficLinear(rName, 10, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists(resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.type", "TimeBasedLinear"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.0.interval", "10"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.0.percentage", "50"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", "0"),
				),
			},
			{
				Config: testAccAWSCodeDeployDeploymentConfigTrafficLinear(rName, 3, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists(resourceName, &config2),
					testAccCheckAWSCodeDeployDeploymentConfigRecreated(&config1, &config2),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.type", "TimeBasedLinear"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.0.interval", "3"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.0.percentage", "10"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSCodeDeployDeploymentConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codedeployconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codedeploy_deployment_config" {
			continue
		}

		resp, err := conn.GetDeploymentConfig(&codedeploy.GetDeploymentConfigInput{
			DeploymentConfigName: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, codedeploy.ErrCodeDeploymentConfigDoesNotExistException, "") {
			continue
		}

		if err == nil {
			if resp.DeploymentConfigInfo != nil {
				return fmt.Errorf("CodeDeploy deployment config still exists:\n%#v", *resp.DeploymentConfigInfo.DeploymentConfigName)
			}
		}

		return err
	}

	return nil
}

func testAccCheckAWSCodeDeployDeploymentConfigExists(name string, config *codedeploy.DeploymentConfigInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).codedeployconn

		resp, err := conn.GetDeploymentConfig(&codedeploy.GetDeploymentConfigInput{
			DeploymentConfigName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*config = *resp.DeploymentConfigInfo

		return nil
	}
}

func testAccCheckAWSCodeDeployDeploymentConfigRecreated(i, j *codedeploy.DeploymentConfigInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreateTime) == aws.TimeValue(j.CreateTime) {
			return errors.New("CodeDeploy Deployment Config was not recreated")
		}

		return nil
	}
}

func testAccAWSCodeDeployDeploymentConfigFleet(rName string, value int) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_deployment_config" "test" {
  deployment_config_name = %q

  minimum_healthy_hosts {
    type  = "FLEET_PERCENT"
    value = %d
  }
}
`, rName, value)
}

func testAccAWSCodeDeployDeploymentConfigHostCount(rName string, value int) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_deployment_config" "test" {
  deployment_config_name = %q

  minimum_healthy_hosts {
    type  = "HOST_COUNT"
    value = %d
  }
}
`, rName, value)
}

func testAccAWSCodeDeployDeploymentConfigTrafficCanary(rName string, interval, percentage int) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_deployment_config" "test" {
  deployment_config_name = %q
  compute_platform = "Lambda"

  traffic_routing_config {
    type = "TimeBasedCanary"

    time_based_canary {
      interval = %d
      percentage = %d
    }
  }
}
`, rName, interval, percentage)
}

func testAccAWSCodeDeployDeploymentConfigTrafficLinear(rName string, interval, percentage int) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_deployment_config" "test" {
  deployment_config_name = %q
  compute_platform = "Lambda"

  traffic_routing_config {
    type = "TimeBasedLinear"

    time_based_linear {
      interval = %d
      percentage = %d
    }
  }
}
`, rName, interval, percentage)
}
