package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCodeDeployDeploymentConfig_fleetPercent(t *testing.T) {
	var config1, config2 codedeploy.DeploymentConfigInfo

	rName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentConfigFleet(rName, 75),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists("aws_codedeploy_deployment_config.foo", &config1),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_config.foo", "minimum_healthy_hosts.0.type", "FLEET_PERCENT"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_config.foo", "minimum_healthy_hosts.0.value", "75"),
				),
			},
			{
				Config: testAccAWSCodeDeployDeploymentConfigFleet(rName, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists("aws_codedeploy_deployment_config.foo", &config2),
					testAccCheckAWSCodeDeployDeploymentConfigRecreated(&config1, &config2),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_config.foo", "minimum_healthy_hosts.0.type", "FLEET_PERCENT"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_config.foo", "minimum_healthy_hosts.0.value", "50"),
				),
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentConfig_hostCount(t *testing.T) {
	var config1, config2 codedeploy.DeploymentConfigInfo

	rName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentConfigHostCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists("aws_codedeploy_deployment_config.foo", &config1),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_config.foo", "minimum_healthy_hosts.0.type", "HOST_COUNT"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_config.foo", "minimum_healthy_hosts.0.value", "1"),
				),
			},
			{
				Config: testAccAWSCodeDeployDeploymentConfigHostCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentConfigExists("aws_codedeploy_deployment_config.foo", &config2),
					testAccCheckAWSCodeDeployDeploymentConfigRecreated(&config1, &config2),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_config.foo", "minimum_healthy_hosts.0.type", "HOST_COUNT"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_config.foo", "minimum_healthy_hosts.0.value", "2"),
				),
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

		if ae, ok := err.(awserr.Error); ok && ae.Code() == "DeploymentConfigDoesNotExistException" {
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
resource "aws_codedeploy_deployment_config" "foo" {
	deployment_config_name = "test-deployment-config-%s"
	minimum_healthy_hosts {
		type = "FLEET_PERCENT"
		value = %d
	}
}`, rName, value)
}

func testAccAWSCodeDeployDeploymentConfigHostCount(rName string, value int) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_deployment_config" "foo" {
	deployment_config_name = "test-deployment-config-%s"
	minimum_healthy_hosts {
		type = "HOST_COUNT"
		value = %d
	}
}`, rName, value)
}
